# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Redis."""

from gevent import core, spawn
from gevent import socket
from gevent.event import AsyncResult
from gevent.hub import Waiter


class RedisError(Exception):
    pass


class Redis(object):
    """Async redis client."""

    _global_cxns = {}
    _pool_size = 5
    _cxn = None
    _in_use = 0

    def __init__(self, host='', port=6379):
        self._addr = addr = (host, port)
        if addr not in self._global_cxns:
            self._cxns = self._global_cxns[addr] = [set(), set(), 0]
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
            del cxn._readable_fileobj
            cxn.close()
        except socket.error:
            pass
        self._cxns[2] -= 1

    for _spec in [
        ('SEND_REQUEST', None, '', ''),
        ('AUTH', '', ''),
        ('PING', '', ''),
        ('QUIT', '', ''),
        ('EXISTS', '', ''),
        ('DEL', 'delete', '', ''),
        ('TYPE', '', ''),
        ('KEYS', '', ''),
        ('RANDOMKEY', '', ''),
        ('RENAME', '', ''),
        ('RENAMENX', '', ''),
        ('DBSIZE', '', ''),
        ('EXPIRE', '', ''),
        ('TTL', '', ''),
        ('SELECT', '', ''),
        ('MOVE', '', ''),
        ('FLUSHDB', '', ''),
        ('FLUSHALL', '', ''),
        ('SET', '', ''),
        ('GET', '', ''),
        ('GETSET', '', ''),
        ('MGET', '', ''),
        ('SETNX', '', ''),
        ('SETEX', '', ''),
        ('MSET', '', ''),
        ('MSETNX', '', ''),
        ('INCR', '', ''),
        ('INCRBY', '', ''),
        ('DECR', '', ''),
        ('DECRBY', '', ''),
        ('APPEND', '', ''),
        ('SUBSTR', '', ''),
        ('RPUSH', '', ''),
        ('LPUSH', '', ''),
        ('LLEN', '', ''),
        ('LRANGE', '', ''),
        ('LTRIM', '', ''),
        ('LINDEX', '', ''),
        ('LSET', '', ''),
        ('LREM', '', ''),
        ('LPOP', '', ''),
        ('RPOP', '', ''),
        ('BLPOP', '', ''),
        ('BRPOP', '', ''),
        ('RPOPLPUSH', '', ''),
        ('SADD', '', ''),
        ('SREP', '', ''),
        ('SPOP', '', ''),
        ('SMOVE', '', ''),
        ('SCARD', '', ''),
        ('SISMEMBER', '', ''),
        ('SINTER', '', ''),
        ('SINTERSTORE', '', ''),
        ('SUNION', '', ''),
        ('SUNIONSTORE', '', ''),
        ('SDIFF', '', ''),
        ('SDIFFSTORE', '', ''),
        ('SMEMBERS', '', ''),
        ('SRANDMEMBER', '', ''),
        ('ZADD', '', ''),
        ('ZREM', '', ''),
        ('ZINCRBY', '', ''),
        ('ZRANK', '', ''),
        ('ZREVRANK', '', ''),
        ('ZRANGE', '', ''),
        ('ZREVRANGE', '', ''),
        ('ZRANGEBYSCORE', '', ''),
        ('ZCARD', '', ''),
        ('ZSCORE', '', ''),
        ('ZREMRANGEBYRANK', '', ''),
        ('ZREMRANGEBYSCORE', '', ''),
        ('ZUNIONSTORE', '', ''),
        ('ZINTERSTORE', '', ''),
        ('HSET', '', ''),
        ('HGET', '', ''),
        ('HMGET', '', ''),
        ('HMSET', '', ''),
        ('HINCRBY', '', ''),
        ('HEXISTS', '', ''),
        ('HDEL', '', ''),
        ('HLEN', '', ''),
        ('HKEYS', '', ''),
        ('HVALS', '', ''),
        ('HGETALL', '', ''),
        ('SORT', '', ''),
        ('MULTI', 'self._in_use = 1', ''),
        ('EXEC', 'execute', 'self._in_use = 0', """
        result = AsyncResult()
        spawn(self.handle_exec_response, cxn).link(result)
        return result.get()
        """),
        ('DISCARD', 'self._in_use = 0', ''),
        ('WATCH', 'self._in_use = 1', ''),
        ('UNWATCH', 'self._in_use = 1', ''),
        ('SUBSCRIBE', '', ''),
        ('UNSUBSCRIBE', '', ''),
        ('PUBLISH', '', ''),
        ('PSUBSCRIBE', '', ''),
        ('PUNSUBSCRIBE', '', ''),
        ('SAVE', '', ''),
        ('BGSAVE', '', ''),
        ('LASTSAVE', '', ''),
        ('SHUTDOWN', '', ''),
        ('BGREWRITEAOF', '', ''),
        ('INFO', '', ''),
        ('SLAVEOF', '', ''),
        ('CONFIG', '', ''),
        # ('MONITOR', '', ''),
        ]:
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
            cxns, waiters, size = self._cxns
            if cxns:
               cxn = self._cxn = cxns.pop()
            elif size < self._pool_size:
                cxn = self._cxn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                cxn.connect(self._addr)
                cxn._readable_fileobj = cxn.makefile('r')
                self._cxns[2] = size + 1
            else:
                waiter = Waiter()
                waiters.add(waiter)
                cxn = self._cxn = waiter.get()

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
            cxns, waiters, _ = self._cxns
            if waiters:
                waiter = waiters.pop()
                waiter.switch(cxn)
            else:
                cxns.add(cxn)

        %s

        result = AsyncResult()
        spawn(self.handle_response, cxn).link(result)
        return result.get()""" % (_name, _extra_1, _extra_2, _before, _after))

    del _command, _name, _before, _after, _spec, _extra_1, _extra_2

    def handle_response(self, cxn):
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
        except Exception:
            self.close_connection(cxn)
            raise
        raise RedisError("Unknown response type %r" % opener)

    def handle_exec_response(self, cxn):
        stream = cxn._readable_fileobj
        responses = []; out = responses.append
        handle_response = self.handle_response
        for i in xrange(int(stream.readline()[1:-2])):
            out(handle_response(cxn))
        return responses


if __name__ == '__main__':

    redis = Redis()
    print repr(redis.send_request('SET', 'foo', 'bar'))
    print redis.send_request('SET', 'foo', 'bar')
    print redis.send_request('SET', 'foo', 'bar')
    print redis.send_request('SET', 'foo', 'bar')
    print redis.send_request('SET', 'foo3', 'bar')
    print redis.send_request('SET', 'foo5', 'bar')

    print redis.set('bar', 1)
    print redis.decr('bar')
    print redis.decr('bar')
    print redis.decr('bar')
    print redis.incr('bar')
    print redis.incr('bar')

    print redis.info()

    print redis.multi()
    print redis.incr('bar')
    print redis.incr('bar')
    print redis.set('foawoa', '\x00aaa')
    print redis.get('foawoa')
    print redis.execute()

    # core.timer(1.2, release_connection, addr, cxns[1])
