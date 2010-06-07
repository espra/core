# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""
=============================
Ampify Zero AppEngine Backend
=============================

"""

import logging
import os
import sys

from cgi import FieldStorage
from time import time
from traceback import format_exception
from urllib import unquote as urlunquote

from google.appengine.api.capabilities import CapabilitySet
from google.appengine.ext import db

# Extend the sys.path to include the ``lib`` subdirectory.
sys.path.insert(0, 'lib')

from simplejson import dumps as json_encode, loads as json_decode
from pyutil.crypto import validate_tamper_proof_string

from config import *

# ------------------------------------------------------------------------------
# the age of ampify has begun!
# ------------------------------------------------------------------------------

AMPIFY_EPOCH = 1262790477000 # in milliseconds since the unix epoch

# ------------------------------------------------------------------------------
# i/o helpers
# ------------------------------------------------------------------------------

class DevNull(object):
    """Provide a file-like interface emulating /dev/null."""

    def __call__(self, *args, **kwargs):
        pass

    def flush(self):
        pass

    def log(self, *args, **kwargs):
        pass

    def write(self, input):
        pass

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

API_HANDLERS = {}
DEVNULL = DevNull()

API_REQUEST_KEYS = frozenset(['payload', 'sig'])
SSL_FLAGS = frozenset(['yes', 'on', '1'])

if os.environ.get('SERVER_SOFTWARE', '').startswith('Google'):
    RUNNING_ON_GOOGLE_SERVERS = True
else:
    RUNNING_ON_GOOGLE_SERVERS = False

OK = "Status: 200 OK\r\n"
ERROR = "Status: 500 Server Error\r\n"

CONTENT_TYPE = "Content-Type: text/plain; charset=utf-8\r\n"
LINE = '\r\n'
OK_HEADER = OK + CONTENT_TYPE + LINE
ERROR_HEADER = ERROR + CONTENT_TYPE + LINE

DISABLED = '{"error":"CapabilityError", "error_msg":"Datastore disabled"}'
NOTFOUND = '{"error":"NotFound", "error_msg":"Invalid API call"}'
NOTAUTHORISED = '{"error":"NotAuthorised", "error_msg":"Invalid API call"}'

UTF8 = 'utf-8'

VALID_HTTP_METHODS = frozenset(['GET', 'POST'])
VALID_REQUEST_CONTENT_TYPES = frozenset([
    '', 'application/x-www-form-urlencoded', 'multipart/form-data'
    ])

# ------------------------------------------------------------------------------
# app runner
# ------------------------------------------------------------------------------

def run_app(
    api=None,
    dict=dict,
    sys=sys,
    API_HANDLERS=API_HANDLERS,
    DEVNULL=DEVNULL,
    ERROR=ERROR,
    ERROR_HEADER=ERROR_HEADER,
    NOTFOUND=NOTFOUND,
    ):
    """The core application runner."""

    env = dict(os.environ)
    kwargs = {}

    sys._boot_stdout = sys.stdout
    sys.stdout = DEVNULL
    write = sys._boot_stdout.write

    try:

        http_method = env['REQUEST_METHOD']
        content_type = env.get('CONTENT-TYPE', '')
        api_method = env['PATH_INFO'][1:]

        if not 
        write()
        return

        args = [arg for arg in env['PATH_INFO'].split('/') if arg]
        if args:
            api = args[0]

        # return a NotFoundError if it doesn't look like a valid api call
        if (http_method != 'POST') or (api not in API_HANDLERS):
            write(ERROR_HEADER)
            write(NOTFOUND)
            return

        # force the request to be over SSL when on a production deployment
        if RUNNING_ON_GOOGLE_SERVERS and env.get('HTTPS') not in SSL_FLAGS:
            write(ERROR_HEADER)
            write(NOTAUTHORISED)
            return

        # we assume that the request is utf-8 encoded, but that the request
        # kwarg "keys" are in ascii and the kwarg values to be in utf-8
        if ';' in content_type:
            content_type = content_type.split(';', 1)[0]

        # parse the POST body if it exists and is of a known content type
        if content_type in VALID_REQUEST_CONTENT_TYPES:

            post_environ = env.copy()
            post_environ['QUERY_STRING'] = ''

            post_data = FieldStorage(
                environ=post_environ, fp=env['wsgi.input']
                ).list

            if post_data:
                for field in post_data:
                    key = field.name
                    if field.filename:
                        continue
                    if key not in API_REQUEST_KEYS:
                        continue
                    value = unicode(field.value, UTF8, 'strict')
                    kwargs[key] = value

        # check that there's a token and it validates
        if 0: # @/@
            signature = kwargs.pop('sig', None)
            if not signature:
                write(ERROR_HEADER)
                write(NOTAUTHORISED)
                return
            if not validate_tamper_proof_string(
                'token', token, key=API_KEY, timestamped=True
                ):
                logging.info("Unauthorised API Access Attempt: %r", token)
                write(UNAUTH)
                return

        handler, store_needed = api_definition

        try:
            # try and respond with the result of calling the api handler
            args = tuple(args)
            result = handler(*args, **kwargs)
            if result:
                write(OK_HEADER)
                write(json_encode(result))
            else:
                write(ERROR)
        except Exception, error:
            # log the error and return it as json
            logging.error(''.join(format_exception(*sys.exc_info())))
            write(ERROR_HEADER)
            write(json_encode({
                "error": error.__class__.__name__,
                "error_msg": str(error)
                }))

    except:
        # this shouldn't ever happen, but just in case...
        logging.critical(''.join(format_exception(*sys.exc_info())))
        write(ERROR_HEADER)
        write(json_encode({
            "error": error.__class__.__name__,
            "error_msg": str(error)
            }))

    finally:
        sys.stdout = sys._boot_stdout

def normalise(id, valid_chars=frozenset('abcdefghijklmnopqrstuvwxyz0123456789.-/')):
    r"normalise the id"
    id = '-'.join(id.replace('_', ' ').lower().split())

def foo():
    words = text.split()
    if len(words) > 5000:
        raise ValueError("The text is longer than 5,000 words!")

# ------------------------------------------------------------------------------
# the zerodatastore api
# ------------------------------------------------------------------------------

def get(foo, bar):
    return foo + bar

def put(ctx):
    return ctx * 2

def delete(ctx=None):
    """
    Delete the Item.

        >>> foo = delete()

    And blah.
    """

    a = 1/0
    return {
        "ok": time()
        }

class Ant(db.Model):
    legs = db.StringProperty()

    a = """
    Foo
    """

def query(ctx):
    key = db.Key.from_path('Ant', 1)
    entity = db.get(key)
    return {
        'id': repr(key),
        'legs': entity.legs
        }

# ------------------------------------------------------------------------------
# you can thank evangineer for this craziness ;p
# ------------------------------------------------------------------------------

# class Lease(db.Model):

#     id = db.StringProperty(name='i')
#     expires = db.DateTimeProperty(default=timedelta(seconds=45), name='e')

# /pecu @evangineer 500
# /pecu-allocated-total: + 500

# allocate_ids(Item, count)

# already exists:
# id1 = {'aspect': '/pecu-allocated-total', 'value': 1000}

# new allocation:
# msg = {'aspect': '/pecu', 'value': 500, 'ref': '@evangineer'}

#! source, new_id, target_id = get_lease('/pecu-allocated-total')

# source looks like:
# {'aspect': '/pecu-allocated-total', 'value': 1000, 'expires': now+45, 'key': 1, 'target': 3, 'other': None}

# new_id: 2
# target_id: 3

# if source:
#   source.value += msg.value
#   msg.key = new_id
#   msg.target = 
#!   save_as(new_id, source)
#!   save_as(target_id, msg)

# read / write / take


# <evangineer> hello

# Item:
#     id = 123
#     by = 'evangineer' # ? 73947
#     to = '#foo'
#     scope/cap = 0
#     value = 'hello'
#     aspect = 'default'

# Item:
#     id = 124
#     by = 'evangineer' # ? 73947
#     to = '374092796242946496297492'
#     value = 'hello'
#     aspect = 'default'

# class Token(db.Model):

#     ref = db.IntegerProperty()
#     read = db.BooleanProperty(default=False)
#     write = db.BooleanProperty(default=False)

# class Reference(db.Model):

#     key = db.ByteStringProperty()

# read
# write
# read + write

# "+lcl": 374092796242946496297492

def multiop(
    pre_id=None,
    pre_query=None,
    pre_cond=None, # == != in > < >= <= not in
    pre_cond_attr=None,
    pre_cond_val=None,
    pre_op=None, # incr, decr, push, pop, set, delitem
    pre_op_attr=None,
    pre_op_value=None,
    pre_op_return=None,
    put_id=None,
    put_val=None,
    post_id=None,
    post_query=None,
    post_cond=None,
    post_cond_attr=None,
    post_cond_val=None,
    post_op=None,
    post_op_attr=None,
    post_op_value=None,
    post_op_return=None,
    ):
    pass

# ------------------------------------------------------------------------------
# register the handlers
# ------------------------------------------------------------------------------

API_HANDLERS.update({
    'get': (get, False),
    'delete': (delete, True),
    'put': (put, True),
    'query': (query, True)
    })

# ------------------------------------------------------------------------------
# self runner -- app engine cached main() function
# ------------------------------------------------------------------------------

if DEBUG == 2:

    from cProfile import Profile
    from pstats import Stats

    def main():
        """Profiling main function."""

        profiler = Profile()
        profiler = profiler.runctx("run_app()", globals(), locals())
        iostream = StringIO()

        stats = Stats(profiler, stream=iostream)
        stats.sort_stats("time")  # or cumulative
        stats.print_stats(80)     # 80 == how many to print

        # optional:
        # stats.print_callees()
        # stats.print_callers()

        logging.info("Profile data:\n%s", iostream.getvalue())

else:

    main = run_app

# ------------------------------------------------------------------------------
# run in standalone mode
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
# foo
# No Copyright (-) 2009-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Parallel Query support."""

from hashlib import sha1

from demjson import decode as json_decode, encode as json_encode

from google.appengine.api import urlfetch
from google.appengine.api.apiproxy_stub_map import UserRPC
from google.appengine.api.datastore import _ToDatastoreError, Entity, Query as RawQuery
from google.appengine.api.datastore_types import Key
from google.appengine.datastore import datastore_index
from google.appengine.datastore.datastore_pb import QueryResult, NextRequest
from google.appengine.ext.db import Query
from google.appengine.runtime.apiproxy_errors import ApplicationError


class BaseQuery(Query):
    """Our BaseQuery subclass."""

    _cursor = None
    _prev = 0

    def execute(
        self, limit, offset, value, set_result, add_callback, deadline, on_complete
        ):

        self._limit = limit
        self._offset = offset
        self._value = value
        self._deadline = deadline
        self.set_result = set_result
        self.add_callback = add_callback
        self.on_complete = on_complete

        raw_query = self._get_query()
        if not isinstance(raw_query, RawQuery):
            raise ValueError(
                "IN and != MultiQueries are not allowed in a ParallelQuery."
                )
        self._buffer = []
        self.rpc_init(raw_query)

    def rpc_init(self, raw_query):

        rpc = UserRPC('datastore_v3', self._deadline)
        rpc.callback = lambda : self.rpc_callback(rpc)
        rpc.make_call(
            'RunQuery', raw_query._ToPb(self._limit, self._offset, self._limit),
            QueryResult()
            )
        self.add_callback(rpc.check_success)

    def rpc_next(self, request):

        rpc = UserRPC('datastore_v3', self._deadline)
        rpc.callback = lambda : self.rpc_callback(rpc)
        rpc.make_call('Next', request, QueryResult())
        self.add_callback(rpc.check_success)

    def rpc_callback(self, rpc):

        try:
            rpc.check_success()
        except ApplicationError, err:
            try:
                raise _ToDatastoreError(err)
            except datastore_errors.NeedIndexError, exc:
                yaml = datastore_index.IndexYamlForQuery(
                    *datastore_index.CompositeIndexForQuery(rpc.request)[1:-1])
                raise datastore_errors.NeedIndexError(
                    str(exc) + '\nThis query needs this index:\n' + yaml)

        response = rpc.response
        more = response.more_results()
        buffer = self._buffer
        buffer.extend(response.result_list())

        if more:
            if self._cursor is None:
                self._cursor = response.cursor()
            remaining = self._limit - len(buffer)
            if remaining and (remaining != self._prev):
                self._prev = remaining
                # logging.error("Requesting %r more for %r [%r]" % (remaining, self._value, len(buffer)))
                request = NextRequest()
                request.set_count(remaining)
                request.mutable_cursor().CopyFrom(self._cursor)
                return self.rpc_next(request)

        self.finish()

    def finish(self):

        try:
            if self._keys_only:
                results = [Key._FromPb(e.key()) for e in self._buffer[:self._limit]]
            else:
                results = [Entity._FromPb(e) for e in self._buffer[:self._limit]]
                if self._model_class is not None:
                    from_entity = self._model_class.from_entity
                    results = [from_entity(e) for e in results]
                else:
                    results = [class_for_kind(e.kind()).from_entity(e) for e in results]
        finally:
            del self._buffer[:]

        if self.on_complete:
            results = self.on_complete(results)
        self.set_result(self._value, results)


class ParallelQuery(object):
    """Parallel query object for doing Trust map queries."""

    def __init__(
        self, model_class=None, keys_only=False, query_key=None,
        cache_duration=5*60, namespace='pq', notify=True, limit=50, offset=0,
        deadline=None, on_complete=None
        ):
        self.model_class = model_class
        self.keys_only = keys_only
        self.query_key = query_key
        self.cache_duration = cache_duration
        self.namespace = namespace
        self.notify = notify
        self.limit = min(limit, 1000)
        self.offset = offset
        self.deadline = deadline
        self.on_complete = on_complete
        self.ops = []
        self.operate = self.ops.append
        self.callbacks = []
        self.results = {}

    def filter(self, property_operator, value):
        self.operate((0, (property_operator, value)))
        return self

    def order(self, property):
        self.operate((1, (property,)))
        return self

    def ancestor(self, ancestor):
        self.operate((2, (ancestor,)))
        return self

    def run(self, property_operator, values, hasher=sha1):

        if not isinstance(values, (list, tuple)):
            raise ValueError(
                "The values for for a ParallelQuery run need to be a list."
                )

        model_class = self.model_class
        keys_only = self.keys_only
        query_key = self.query_key
        limit = self.limit
        offset = self.offset
        deadline = self.deadline
        on_complete = self.on_complete
        ops = self.ops
        results = self.results
        set_result = results.__setitem__
        callbacks = self.callbacks
        add_callback = callbacks.append

        if query_key:
            key_prefix = '%s-%s-%s' % (
                hasher(query_key).hexdigest(), limit, offset
                )
            cache = memcache.get_multi(values, key_prefix, self.namespace)
        else:
            cache = {}

        for value in values:
            if value in cache:
                continue
            if limit == 0:
                set_result(value, [])
                continue
            query = BaseQuery(model_class, keys_only)
            for op, args in ops:
                if op == 0:
                    query.filter(*args)
                elif op == 1:
                    query.order(*args)
                elif op == 2:
                    query.ancestor(*args)
            query.filter(property_operator, value)
            query.execute(
                limit, offset, value, set_result, add_callback, deadline,
                on_complete
                )

        try:
            while callbacks:
                callback = callbacks.pop()
                callback()
            if query_key:
                unset_keys = memcache.set_multi(
                    results, cache_duration, key_prefix, self.namespace
                    )
                if notify:
                    set_keys = set(results).difference(set(unset_keys))
                    if set_keys:
                        rpc = urlfetch.create_rpc(deadline=10)
                        urlfetch.make_fetch_call(
                            rpc, notify, method='POST', payload=json_encode({
                                'key_prefix': key_prefix, 'keys': set_keys
                                })
                            )
            for key in cache:
                results[key] = cache[key]
            return self
        finally:
            del callbacks[:]


# def on_complete(results):
#     return [item.key().id() for item in results]

# query = ParallelQuery(
#   Item, query_key='/intentions #espians',
#   notify='http://notify.tentapp.com', on_complete=on_complete
#   )

# query.filter('aspect =', '/intention')
# query.filter('space =', 'espians')
# query.run('by =', [('evangineer', 'olasofia', 'sbp']))

# query.results <--- {'evangineer': [...], 'olasofia': [...]}
