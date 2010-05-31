# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

from gevent import sleep, socket, spawn
from gevent.event import AsyncResult
from gevent.queue import Queue

# ------------------------------------------------------------------------------
# Some Constants
# ------------------------------------------------------------------------------

NORMAL_COMMANDS = """
  APPEND AUTH BGREWRITEAOF BGSAVE CONFIG DBSIZE DECR DECRBY EXISTS EXPIRE
  FLUSHALL FLUSHDB GET GETSET HDEL HEXISTS HGET HGETALL HINCRBY HKEYS HLEN HMGET
  HMSET HSET HVALS INCR INCRBY INFO KEYS LASTSAVE LINDEX LLEN LPOP LPUSH LRANGE
  LREM LSET LTRIM MGET MOVE MSET MSETNX PING PSUBSCRIBE PUBLISH PUNSUBSCRIBE
  QUIT RANDOMKEY RENAME RENAMENX RPOP RPOPLPUSH RPUSH SADD SAVE SCARD SDIFF
  SDIFFSTORE SELECT SET SETEX SETNX SHUTDOWN SINTER SINTERSTORE SISMEMBER
  SLAVEOF SMEMBERS SMOVE SORT SPOP SRANDMEMBER SREP SUBSCRIBE SUBSTR SUNION
  SUNIONSTORE TTL TYPE UNSUBSCRIBE ZADD ZCARD ZINCRBY ZINTERSTORE ZRANGE
  ZRANGEBYSCORE ZRANK ZREM ZREMRANGEBYRANK ZREMRANGEBYSCORE ZREVRANGE ZREVRANK
  ZSCORE ZUNIONSTORE
  """.strip().split()

# ------------------------------------------------------------------------------
# Before/After Blocks
# ------------------------------------------------------------------------------

BLOCKING_BEFORE = """
        if not isinstance(args[-1], (int, float)): args = args + (0,)
        self._in_use = 1"""

BLOCKING_AFTER = """
        result = AsyncResult()
        spawn(self.handle_response, cxn).link(result)
        self._in_use = 0
        try:
            return result.get()
        finally:
            self._cxn = None
            self._cxns.add(cxn)"""

# ------------------------------------------------------------------------------
# Exceptions
# ------------------------------------------------------------------------------

class RedisError(Exception):
    pass

# ------------------------------------------------------------------------------
# The Redis Client
# ------------------------------------------------------------------------------

class Redis(object):
    """Async redis client."""

    _global_cxns = {}
    _max_cxns = 100
    _open_cxns = 0
    _cxn = None
    _in_use = 0

    def __init__(self, host='', port=6379):
        self._addr = addr = (host, port)
        if addr not in self._global_cxns:
            self._cxns = self._global_cxns[addr] = set()
        else:
            self._cxns = self._global_cxns[addr]

    def close_connection(self):
        cxn, self._cxn = self._cxn, None
        if not cxn:
            return
        try:
            try:
                cxn._readable_fileobj.close()
            except Exception:
                pass
            cxn.close()
        except socket.error:
            pass
        finally:
            self._open_cxns -= 1

    for _spec in [
        ('SEND_REQUEST', None, '', ''),
        ('DEL', 'delete', '', ''),
        ('BLPOP', BLOCKING_BEFORE, BLOCKING_AFTER),
        ('BRPOP', BLOCKING_BEFORE, BLOCKING_AFTER),
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
            cxns = self._cxns
            #while (not cxns) and self._open_cxns >= self._max_cxns:
            #    sleep(0.1)
            if cxns:
                cxn = self._cxn = cxns.pop()
            else:
                self._open_cxns += 1
                try:
                    cxn = socket.socket(
                        socket.AF_INET, socket.SOCK_STREAM
                    )
                    cxn.setsockopt(socket.SOL_TCP, socket.TCP_NODELAY, 1)
                    cxn.connect(self._addr)
                    cxn._readable_fileobj = cxn.makefile('r')
                except Exception:
                    self._open_cxns -= 1
                    raise
                self._cxn = cxn

        %s

        request = ['*%%i\r\n' %% len(args)]; out = request.append
        for arg in args:
            arg = str(arg)
            out('$%%i' %% len(arg))
            out('\r\n')
            out(arg)
            out('\r\n')

        cxn.sendall(''.join(request))
        #try:
        #    cxn.sendall(''.join(request))
        #except socket.error:
        #    raise
        #    self.close_connection()

        %s

        result = AsyncResult()
        spawn(self.handle_response).link(result)
        #res = result.get()
        # print "RESULT"
        #return res
        return result.get()""" % (_name, _extra_1, _extra_2, _before, _after))

    del _command, _name, _before, _after, _spec, _extra_1, _extra_2

    def handle_response(self, standalone=1):
        global x
        print "Handling", x
        y = x = x + 1
        cxn = self._cxn
        print repr(cxn)
        #if not cxn:
        #    raise IOError("Connection closed.")
        if 1:
        #try:
            print "Strt", y
            stream = cxn._readable_fileobj
            opener = stream.read(1)
            print "Strt-2", y
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
                if standalone:
                    raise RedisError(stream.readline()[:-2])
                return RedisError(stream.readline()[:-2])
            raise RedisError("Unknown response type %r" % opener)
        #except socket.error:
        #    pass
            #print "--------------------------------SOCKERR"
            #self.close_connection()
            #raise
        #finally:
        #    pass
            #if standalone and not self._in_use:
                #self._cxn = None
                #self._cxns.add(cxn)
        #print "Finished", y


    def handle_exec_response(self, cxn):
        try:
            stream = cxn._readable_fileobj
            responses = []; out = responses.append
            handle_response = self.handle_response
            for i in xrange(int(stream.readline()[1:-2])):
                out(handle_response(standalone=0))
        finally:
            self._cxn = None
            self._cxns.add(cxn)
        return responses


x = 0

r = Redis('espians.com', 9094)

def foo(i):
    r.set('foo', i)
    r.close_connection()

from gevent import joinall
from time import time

start = time()
joinall(
    [spawn(foo, i) for i in range(1000)]
    )
print time() - start
#print r.get('foo')

# r.set('foo', 'bar')
