# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import logging
import socket

from collections import deque
from time import time

from ampify.async import wrap_method

from tornado.ioloop import IOLoop
from tornado.iostream import IOStream

# ------------------------------------------------------------------------------
# Some Constants
# ------------------------------------------------------------------------------

Loop = IOLoop.instance()

NORMAL_COMMANDS = """
  APPEND AUTH BGREWRITEAOF BGSAVE CONFIG DBSIZE DECR DECRBY EXISTS EXPIRE
  FLUSHALL FLUSHDB GET GETSET HDEL HEXISTS HGET HGETALL HINCRBY HKEYS HLEN HMGET
  HMSET HSET HVALS INCR INCRBY INFO KEYS LASTSAVE LINDEX LLEN LPOP LPUSH LRANGE
  LREM LSET LTRIM MGET MOVE MSET MSETNX PING PSUBSCRIBE PUBLISH PUNSUBSCRIBE
  QUIT RANDOMKEY RENAME RENAMENX RPOP RPOPLPUSH RPUSH SADD SAVE SCARD SDIFF
  SDIFFSTORE SELECT SET SETEX SETNX SHUTDOWN SINTER SINTERSTORE SISMEMBER
  SLAVEOF SMEMBERS SMOVE SORT SPOP SRANDMEMBER SREM SUBSCRIBE SUBSTR SUNION
  SUNIONSTORE TTL TYPE UNSUBSCRIBE ZADD ZCARD ZCOUNT ZINCRBY ZINTERSTORE ZRANGE
  ZRANGEBYSCORE ZRANK ZREM ZREMRANGEBYRANK ZREMRANGEBYSCORE ZREVRANGE ZREVRANK
  ZSCORE ZUNIONSTORE
  """.strip().split()

# ------------------------------------------------------------------------------
# Utility Functions
# ------------------------------------------------------------------------------

def set_max_connections(value):
    Redis._max_cxns = value

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
    _max_cxns = None
    _open_cxns = 0
    _cxn = None
    _in_txn = 0
    _multi_wait = 0
    _in_progress = 0
    _opened = 0

    def __init__(self, host='', port=6379):
        self._addr = addr = (host, port)
        if addr not in self._global_cxns:
            self._cxns = self._global_cxns[addr] = set()
        else:
            self._cxns = self._global_cxns[addr]

    def handle_connection_close(self, cxn):
        if not cxn._discarded:
            cxn._discarded = 1
            Redis._open_cxns -= 1
            while cxn.queue:
                errback = cxn.queue.popleft()[-1]
                if errback:
                    try:
                        errback(socket.error("Connection closed."))
                    except Exception:
                        pass
            del cxn.queue
        self._cxns.discard(cxn)

    def close_connection(self):
        cxn, self._cxn = self._cxn, None
        if not cxn:
            return
        try:
            cxn.close()
        except socket.error:
            pass

    for _spec in [
        ('SEND_REQUEST', None, '', ''),
        ('DEL', 'delete', '', ''),
        ('BLPOP', """
        if not isinstance(args[-1], (int, float)): args = args + (0,)""", ""),
        ('BRPOP', """
        if not isinstance(args[-1], (int, float)): args = args + (0,)""", ""),
        ('MULTI', '', "txn = 1"),
        ('EXEC', 'execute', '', "multi = txn_end = 1"),
        ('DISCARD', '', "txn_end = 1"),
        ('WATCH', '', "txn = 1"),
        ('UNWATCH', '', "txn = 1"),
        ('MONITOR', '', "persist = 1"),
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
            _extra_3 = 'args[0], '
        else:
            _extra_1 = _extra_3 = ''
            _extra_2 = 'args = (%r,) + args' % _command

        exec(r"""def %s(self, %s*args, **kwargs):

        %s

        if not args:
            self.close_connection()
            raise ValueError("No arguments specified for redis call.")

        cxn = self._cxn
        if not cxn:
            cxns = self._cxns
            max_cxns = Redis._max_cxns
            if (not cxns) and max_cxns and Redis._open_cxns > max_cxns:
                closed = None
                for addr, open_cxns in self._global_cxns.iteritems():
                    if open_cxns:
                        closed = open_cxns.pop()
                        closed.close()
                        break
                if not closed:
                    if '_start' not in kwargs:
                        kwargs['_start'] = time()
                    now = time()
                    if (now - kwargs['_start']) > 60:   # @/@ configurable?
                        kwargs['_start'] = now
                        logging.warn(
                            "Redis connection starving [0x%%x] %%r"
                            %% (id(self), self._addr)
                            )
                    Loop.add_timeout(
                        now + 0.1,
                        lambda: self.%s.__raw__(self, %s *args[1:], **kwargs)
                    )
                    return self
            if cxns:
                cxn = self._cxn = cxns.pop()
            else:
                Redis._open_cxns += 1
                try:
                    sock = socket.socket(
                        socket.AF_INET, socket.SOCK_STREAM
                    )
                    sock.setsockopt(socket.SOL_TCP, socket.TCP_NODELAY, 1)
                    sock.connect(self._addr)
                    cxn = IOStream(sock)
                    cxn.queue = deque()
                    cxn._discarded = 0
                except Exception:
                    Redis._open_cxns -= 1
                    raise
                cxn._close_callback = lambda: self.handle_connection_close(cxn)
                self._cxn = cxn

        %s

        request = ['*%%i\r\n' %% len(args)]; out = request.append
        for arg in args:
            arg = str(arg)
            out('$%%i' %% len(arg))
            out('\r\n')
            out(arg)
            out('\r\n')

        try:
            cxn.write(''.join(request))
        except socket.error:
            self.close_connection()
            raise

        multi = txn = txn_end = persist = None

        %s

        callback = kwargs.pop('callback', None)
        errback = kwargs.pop('errback', None)
        cxn.queue.append([txn, txn_end, multi, persist, 1, callback, errback])
        self.handle_response()

        return self""" % (_name, _extra_1, _extra_2, _name, _extra_3, _before, _after))

        locals()[_name] = wrap_method(locals()[_name])

    del _command, _name, _before, _after, _spec, _extra_1, _extra_2, _extra_3

    def callback(self, result, err=None):
        self._in_progress = 0
        queue = self._cxn.queue
        if self._multi_wait:
            if err:
                self._multi_wait = None
                del self._multi_results, self._multi_result_left
            elif self._multi_result_left is None:
                self._multi_result_left = left = int(result[1:-2])
                if left == -1:
                    result = None
                    self._multi_wait = None
                    del self._multi_results, self._multi_result_left
                else:
                    return self._cxn.read_until('\r\n', self.handle_response)
            else:
                self._multi_result_left -= 1
                self._multi_results.append(result)
                if self._multi_result_left:
                    return self._cxn.read_until('\r\n', self.handle_response)
                else:
                    result = self._multi_results
                    self._multi_wait = None
                    del self._multi_results, self._multi_result_left
        txn, txn_end, multi, persist, stage, callback, errback = queue.popleft()
        if self._in_txn and txn_end:
            self._in_txn = 0
        cb = errback if err else callback
        Redis._opened += 1
        if cb:
            cb(result)
        if persist:
            queue.appendleft([txn, txn_end, multi, persist, 1, callback, errback])
        if queue:
            Loop.add_callback(self.handle_response)
        else:
            if not self._in_txn:
                self._cxns.add(self._cxn)
                self._cxn = None

    def errback(self, error):
        self.callback(error, 1)

    def handle_response(self, data=None):
        try:
            cxn = self._cxn
            txn, _, multi, _, stage, _, _ = cxn.queue[0]
            if data is None:
                if self._in_progress:
                    return
                if txn:
                    self._in_txn = 1
                self._in_progress = 1
                if multi:
                    self._multi_wait = 1
                    self._multi_results = []
                    self._multi_result_left = None
                    return cxn.read_until('\r\n', self.callback)
                return cxn.read_until('\r\n', self.handle_response)
            if stage == 1:
                opener = data[0]
                if opener == '+':
                    return self.callback(data[1:-2])
                if opener == ':':
                    return self.callback(int(data[1:-2]))
                if opener == '$':
                    length = int(data[1:-2])
                    if length == -1:
                        return self.callback(None)
                    cxn.queue[0][-3] = 2
                    return cxn.read_bytes(length+2, self.handle_response)
                if opener == '-':
                    return self.errback(RedisError(data[:-2]))
                if opener == '*':
                    self._results = []
                    self._result_left = int(data[1:-2])
                    cxn.queue[0][-3] = 3
                    return cxn.read_until('\r\n', self.handle_response)
                return self.errback(RedisError("Unknown response %r" % data))
            if stage == 2:
                return self.callback(data[:-2])
            if stage == 3:
                length = int(data[1:-2])
                if length == -1:
                    self._results.append(None)
                    self._result_left -= 1
                    if self._result_left:
                        return cxn.read_until('\r\n', self.handle_response)
                else:
                    cxn.queue[0][-3] = 4
                    return cxn.read_bytes(length+2, self.handle_response)
            if stage == 4:
                self._results.append(data[:-2])
                self._result_left -= 1
                if self._result_left:
                    cxn.queue[0][-3] = 3
                    return cxn.read_until('\r\n', self.handle_response)
            results = self._results
            del self._results, self._result_left
            return self.callback(results)
        except Exception:
            self.close_connection()
            raise


if __name__ == '__main__':

    from adisp import process

    redis = Redis()
    def handle_get(result):
        print "GOT:", result

    redis.set('name', 'tav', run=1)
    redis.get('name')(handle_get)

    N = 200
    set_max_connections(5)

    @process
    def test_set(i):
        redis = Redis()
        x = yield redis.send_request('set', 'foo', i)
        yield redis.multi()
        x = yield redis.set('foo', i)
        print x
        try:
            x = yield redis.get('foo', i)
        except Exception, error:
            print "aaas", error
        x = yield redis.incr('foos')
        print x
        x = yield redis.incr('foo')
        print repr(x)
        x = yield redis.execute()
        print repr(x)
        return
        try:
            x = yield redis.send_request('ping')
        except Exception, error:
            print 'got error', repr(error), "fa"
            return
        # x = yield redis.ping()
        print x
        #_ = yield redis.set('foo', i)
        # print _, i

    @process
    def test_get():
        redis = Redis()
        result = yield redis.send_request('get', 'foo')
        print "Got:", result

    def test_monitor():
        def handle_monitor_line(line):
            print "mon", line
        redis = Redis()
        redis.monitor()(handle_monitor_line)

    test_monitor()

    for i in xrange(N):
        test_set(i)
    else:
        print 'get'
        test_get()

    def print_max():
        print Redis._opened
        Loop.add_timeout(time() + 1, print_max)

    print_max()

    def on_event(result):
        print "GOT!!", result

    r = Redis()
    r.blpop('hmz')(callback=on_event)

    Loop.start()

# r = Redis('espians.com', 9094)

# def foo(i):
#     r.set('foo', i)
#     r.close_connection()

# from gevent import joinall
# from time import time

# start = time()
# joinall(
#     [spawn(foo, i) for i in range(1000)]
#     )
# print time() - start
#print r.get('foo')

# r.set('foo', 'bar')
