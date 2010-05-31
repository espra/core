# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

from gevent import socket, spawn
from gevent.event import AsyncResult
from gevent.queue import Queue


NORMAL_COMMANDS = """
  APPEND AUTH BGREWRITEAOF BGSAVE BLPOP BRPOP CONFIG DBSIZE DECR DECRBY EXISTS
  EXPIRE FLUSHALL FLUSHDB GET GETSET HDEL HEXISTS HGET HGETALL HINCRBY HKEYS
  HLEN HMGET HMSET HSET HVALS INCR INCRBY INFO KEYS LASTSAVE LINDEX LLEN LPOP
  LPUSH LRANGE LREM LSET LTRIM MGET MOVE MSET MSETNX PING PSUBSCRIBE PUBLISH
  PUNSUBSCRIBE QUIT RANDOMKEY RENAME RENAMENX RPOP RPOPLPUSH RPUSH SADD SAVE
  SCARD SDIFF SDIFFSTORE SELECT SET SETEX SETNX SHUTDOWN SINTER SINTERSTORE
  SISMEMBER SLAVEOF SMEMBERS SMOVE SORT SPOP SRANDMEMBER SREP SUBSCRIBE SUBSTR
  SUNION SUNIONSTORE TTL TYPE UNSUBSCRIBE ZADD ZCARD ZINCRBY ZINTERSTORE ZRANGE
  ZRANGEBYSCORE ZRANK ZREM ZREMRANGEBYRANK ZREMRANGEBYSCORE ZREVRANGE ZREVRANK
  ZSCORE ZUNIONSTORE
  """.strip().split()


class RedisError(Exception):
    pass


class Redis(object):
    """Async redis client."""

    _global_cxns = {}
    _pool_size = 2
    _cxn = None
    _in_use = 0

    def __init__(self, host='', port=6379):
        self._addr = addr = (host, port)
        if addr not in self._global_cxns:
            self._cxns = self._global_cxns[addr] = [Queue(), 0]
        else:
            self._cxns = self._global_cxns[addr]

    def close_connection(self, cxn=None):
        if not cxn:
            cxn, self._cxn = self._cxn, None
        try:
            try:
                cxn._readable_fileobj.close()
            except Exception:
                pass
            # del cxn._readable_fileobj
            cxn.close()
        except socket.error:
            pass
        self._cxns[1] -= 1

    for _spec in [
        ('SEND_REQUEST', None, '', ''),
        ('DEL', 'delete', '', ''),
        ('MULTI', 'self._in_use = 1', ''),
        ('EXEC', 'execute', '', """
        result = AsyncResult()
        spawn(self.handle_exec_response, cxn).link(result)
        self._in_use = 0
        return result.get()
        """),
        ('DISCARD', 'self._in_use = 0', ''),
        ('WATCH', 'self._in_use = 1', ''),
        ('UNWATCH', 'self._in_use = 1', ''),
        # ('MONITOR', '', ''),
        ] + [(cmd, '', '') for cmd in NORMAL_COMMANDS]:
        if len(_spec) == 3:
            _command, _before, _after = _spec
            _name = _command.lower()
        else:
            _command, _name, _before, _after = _spec
        if not _name:
            _name = 'send_request'
            _extra_1 = 'cmd, '
            _extra_2 = 'args = (cmd,) + args'
        else:
            _extra_1 = ''
            _extra_2 = 'args = (%r,) + args' % _command

        exec(r"""def %s(self, %s*args):
        %s

        if not args:
            raise ValueError("No arguments specified for redis call.")

        cxn = self._cxn
        if not cxn:
            queue, size = self._cxns
            if size < self._pool_size:
                self._cxns[1] = size + 1
                cxn = self._cxn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                cxn.connect(self._addr)
                cxn._readable_fileobj = cxn.makefile('r')
            else:
                cxn = self._cxn = queue.get()

        %s

        request = ['*%%i\r\n' %% len(args)]; out = request.append
        for arg in args:
            arg = str(arg)
            out('$%%i' %% len(arg))
            out('\r\n')
            out(arg)
            out('\r\n')

        try:
            cxn.sendall(''.join(request))
        except socket.error:
            self.close_connection()
            raise

        if not self._in_use:
            self._cxn = None
            self._cxns[0].put(cxn)

        %s

        result = AsyncResult()
        spawn(self.handle_response, cxn).link(result)
        return result.get()""" % (_name, _extra_1, _extra_2, _before, _after))

    del _command, _name, _before, _after, _spec, _extra_1, _extra_2

    def handle_response(self, cxn, standalone=1):
        stream = cxn._readable_fileobj
        try:
            opener = stream.read(1)
            if opener == '+':
                return stream.readline()[:-2]
            if opener == ':':
                return int(stream.readline()[:-2])
            if opener == '$':
                length = int(stream.readline()[:-2])
                if length == '-1':
                    return None
                return stream.read(length+2)[:-2]
            if opener == '*':
                result = []; out = result.append
                readline = stream.readline; read = stream.read
                for i in xrange(int(readline()[:-2])):
                    length = int(readline()[1:-2])
                    if length == '-1':
                        out(None)
                    out(read(length+2)[:-2])
                return result
            if opener == '-':
                raise RedisError(stream.readline()[:-2])
            raise RedisError("Unknown response type %r" % opener)
        except socket.error:
            self.close_connection(cxn)
            raise
        finally:
            if standalone and not self._in_use:
                self._cxn = None
                self._cxns[0].put(cxn)

    def handle_exec_response(self, cxn):
        try:
            stream = cxn._readable_fileobj
            responses = []; out = responses.append
            handle_response = self.handle_response
            for i in xrange(int(stream.readline()[1:-2])):
                out(handle_response(cxn, standalone=0))
        finally:
            self._cxn = None
            self._cxns[0].put(cxn)
        return responses
