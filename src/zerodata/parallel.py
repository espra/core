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
