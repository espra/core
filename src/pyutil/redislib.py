# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Minimalistic asynchronous Python client for Redis >= 1.2"""

import gevent
from gevent import socket


class RedisError(Exception): pass
class ResponseError(RedisError): pass


class Response(object):
    """Provides a parser for Redis responses."""
    def __init__(self, fp):
        self._fp = fp

    def __call__(self):
        resp_type = self._fp.read(1)
        if   resp_type == '-': return self.error_response()
        elif resp_type == '+': return self.inline_response()
        elif resp_type == '$': return self.bulk_response()
        elif resp_type == '*': return list(self.multi_bulk_response())
        elif resp_type == ':': return self.integer_response()
        return None

    def error_response(self):
        self._fp.read(4)
        error = self._fp.readline().strip()
        raise ResponseError(error)

    def inline_response(self):
        return self._fp.readline().strip()

    def bulk_response(self):
        length = int(self._fp.readline().strip())
        if length == -1:
            return None
        return self._fp.read(length+2).strip()

    def multi_bulk_response(self):
        length = int(self._fp.readline().strip())
        for x in xrange(length):
            if self._fp.read(1) == '$':
                yield self.bulk_response()

    def integer_response(self):
        return int(self._fp.readline().strip())


class Connection(object):
    """Provides a connection to a Redis server."""

    def __init__(self, host='localhost', port=6379, db=None):
        self._host = host
        self._port = port
        self._db = db
        self._socket = None

    def connect(self):
        if self._socket:
            return

        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.connect((self._host, self._port))
        self._file = self._socket.makefile('r')

        if self._db:
            self.send('SELECT', str(self._db))

    def send(self, in_commands):
        self.connect()

        if type(in_commands)==type(str()):
            commands = in_commands.split()
        else:
            commands = in_commands

        try:
            self._socket.sendall('*%d\r\n' % len(commands))
            for command in commands:
                self._socket.sendall('$%d\r\n' % len(str(command)))
                self._socket.sendall('%s\r\n' % command)
            return self.response()
        except socket.error:
            self.disconnect()
            return self.send(*commands)

    def response(self):
        resp = Response(self._file)
        return resp()

    def disconnect(self):
        try:
            self._socket.close()
        except socket.error:
            pass

        self._socket = None
        self._file = None

