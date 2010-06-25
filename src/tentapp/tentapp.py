# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""
===========================
Tent Collaboration Platform
===========================

Tent brings together IRC-like real-time collaboration with trust maps and
structured data.

"""

import logging
import os
import re
import sys

from BaseHTTPServer import BaseHTTPRequestHandler
from binascii import hexlify
from cgi import FieldStorage
from datetime import datetime
from hashlib import sha1
from md5 import md5
from os import urandom
from os.path import exists, join as join_path, getmtime
from posixpath import split as split_path
from time import time
from traceback import format_exception
from urllib import urlencode, quote as urlquote, unquote as urlunquote
from urlparse import urljoin
from wsgiref.headers import Headers

from google.appengine.ext import db
from google.appengine.runtime.apiproxy_errors import CapabilityDisabledError

# Extend the sys.path to include the ``lib`` subdirectory.
sys.path.insert(0, 'lib')

from cookie import SimpleCookie # note: this is our cookie and not Cookie...
from mako.template import Template as MakoTemplate
from pyutil.crypto import create_tamper_proof_string
from pyutil.crypto import validate_tamper_proof_string
from pyutil.exception import html_format_exception
from pyutil.io import DEVNULL
from pyutil.jsonp import is_valid_jsonp_callback_value
from pyutil.sanitise import match_valid_uri_scheme, sanitise
from simplejson import dumps as json_encode, loads as json_decode

from webob import Request as WebObRequest # this import patches cgi.FieldStorage
                                          # to behave better for us too!

from config import *

# ------------------------------------------------------------------------------
# Constants
# ------------------------------------------------------------------------------

# Item IDs are generated relative to the Ampify epoch in order to save space.
# This is, in turn, defined in terms of milliseconds since the Unix epoch.
AMPIFY_EPOCH = 1262790477000

COOKIE_KEY_NAMES = frozenset([
    'domain', 'expires', 'httponly', 'max-age', 'path', 'secure', 'version'
    ])

CSS_FILE_PATH = '%ssite.css?%s' % (STATIC_PATH, APPLICATION_TIMESTAMP)

ERROR_WRAPPER = """<!DOCTYPE html>
<html>
  <head>
    <title>Error!</title>
    <meta content="text/html; charset=utf-8" http-equiv="content-type" />
    <!--[if !IE]><!-->
      <link rel="stylesheet" type="text/css" href="%s" />
    <!--<![endif]-->
    <!--[if gte IE 8]>
      <link rel="stylesheet" type="text/css" href="%s" />
    <![endif]-->
    <!--[if lte IE 7]>
      <link rel="stylesheet" type="text/css" href="%s" />
    <![endif]-->
  </head>
  <body>
  %%s
  </body>
</html>
""" % (CSS_FILE_PATH, CSS_FILE_PATH, CSS_FILE_PATH)

ERROR_401 = ERROR_WRAPPER % """
  <div class="error">
    <h1>Not Authorized</h1>
    Your session may have expired or you may not have access.
    <ul>
      <li><a href="/">Return home</a></li>
      <li><a href="/login">Login</a></li>
    </ul>
  </div>
  """ # emacs"

ERROR_404 = ERROR_WRAPPER % """
  <div class="error">
    <h1>The item you requested was not found</h1>
    You may have clicked a dead link or mistyped the address. Some web addresses
    are case sensitive.
    <ul>
      <li><a href="/">Return home</a></li>
    </ul>
  </div>
  """ # emacs"

ERROR_500_BASE = ERROR_WRAPPER % """
  <div class="error">
    <h1>Sorry, something went wrong!</h1>
    There was an application error. This has been logged and will be resolved as
    soon as possible.
    <ul>
      <li><a href="/">Return home</a></li>
    </ul>
    %s
  </div>
  """ # emacs"

ERROR_500 = ERROR_500_BASE % ""

ERROR_500_TRACEBACK = ERROR_500_BASE % """
    <div class="traceback">%s</div>
  """ # emacs"

ERROR_503 = ERROR_WRAPPER % """
  <div class="error">
    <h1>Service Unavailable</h1>
    Google App Engine is currently down for a scheduled maintenance.
    Please try again later.
    <ul>
      <li><a href="/">Return home</a></li>
    </ul>
  </div>
  """ # emacs"

HTTP_STATUS_MESSAGES = BaseHTTPRequestHandler.responses

RESPONSE_NOT_IMPLEMENTED = "Status: 501 Not Implemented\r\n\r\n"

RESPONSE_OPTIONS = (
    "Status: 200 OK\r\n"
    "Allow: OPTIONS, GET, HEAD, POST\r\n\r\n"
    )

RESPONSE_301 = (
    "Status: 301 Moved Permanently\r\n"
    "Location: %s\r\n\r\n"
    )

RESPONSE_302 = (
    "Status: 302 Found\r\n"
    "Location: %s\r\n\r\n"
    )

RESPONSE_X = (
    "Status: %s\r\n"
    "Content-Type: text/html\r\n"
    "Content-Length: %s\r\n\r\n"
    "%s"
    )

RESPONSE_401 = (
    "Status: 401 Unauthorized\r\n"
    "WWW-Authenticate: Token realm='Service', error='token_expired'\r\n"
    "Content-Type: text/html\r\n"
    "Content-Length: %s\r\n\r\n"
    "%s"
    ) % (len(ERROR_401), ERROR_401)

RESPONSE_404 = RESPONSE_X % ("404 Not Found", len(ERROR_404), ERROR_404)
RESPONSE_500_BASE = RESPONSE_X % ("500 Server Error", "%s", "%s")
RESPONSE_500 = RESPONSE_X % ("500 Server Error", len(ERROR_500), ERROR_500)
RESPONSE_503 = RESPONSE_X % (
    "503 Service Unavailable", len(ERROR_503), ERROR_503
    )

RESPONSE_JSON_ERROR = (
    "Status: 500 Server Error\r\n"
    "Content-Type: application/json; charset=utf-8\r\n"
    "Content-Length: %s\r\n\r\n"
    "%s"
    )

if os.environ.get('SERVER_SOFTWARE', '').startswith('Google'):
    RUNNING_ON_GOOGLE_SERVERS = True
else:
    RUNNING_ON_GOOGLE_SERVERS = False
    TENT_HTTP_HOST = TENT_HTTPS_HOST = 'http://localhost:8080'
    COOKIE_DOMAIN_HTTP = COOKIE_DOMAIN_HTTPS = ''

SERVICE_REGISTRY = {}

SSL_FLAGS = frozenset(['yes', 'on', '1'])

SUPPORTED_HTTP_METHODS = frozenset(['GET', 'HEAD', 'POST'])

VALID_REQUEST_CONTENT_TYPES = frozenset([
    '', 'application/x-www-form-urlencoded', 'multipart/form-data'
    ])

# ------------------------------------------------------------------------------
# Exceptions
# ------------------------------------------------------------------------------

# Services can throw exceptions to return specifc HTTP response codes.
#
# All the errors subclass the ``BaseHTTPError``.
class BaseHTTPError(Exception):
    pass

# The ``Redirect`` exception is used to handle HTTP 301/302 redirects.
class Redirect(BaseHTTPError):
    def __init__(self, uri, permanent=False):
        self.uri = urljoin('', uri)
        self.permanent = permanent

# The ``HTTPContent`` is used to return the associated content.
class HTTPContent(BaseHTTPError):
    def __init__(self, content):
        self.content = content

# The ``AuthError`` is used to represent the 401 Not Authorized error.
class AuthError(BaseHTTPError):
    pass

# The ``NotFound`` is used to represent the classic 404 error.
class NotFound(BaseHTTPError):
    pass

# The ``HTTPError`` is used to represent all other response codes.
class HTTPError(BaseHTTPError):
    def __init__(self, code=500):
        self.code = code

# ------------------------------------------------------------------------------
# Models
# ------------------------------------------------------------------------------

class Alias(db.Model):
    v = db.IntegerProperty(default=0, name='v')

class Amp(db.Model):
    v = db.IntegerProperty(default=0, name='v')

class AccessToken(db.Model):
    v = db.IntegerProperty(default=0, name='v')
    created = db.DateTimeProperty(auto_now_add=True)
    accepted = db.BooleanProperty(default=False)
    control_ref = db.StringProperty(indexed=False)
    defined = db.StringListProperty()
    expires = db.IntegerProperty() # any number less than 1000 => "oneshot"
    read_ref = db.StringProperty(indexed=False)
    remove_ref = db.StringProperty(indexed=False)
    space = db.StringProperty() # *, +739, ~tav, ~tav/foo*, #blah, #*, @tav
    write_ref = db.StringProperty(indexed=False)
    write_aspects_ref = db.StringProperty(indexed=False)

class CloakToken(db.Model):
    v = db.IntegerProperty(default=0, name='v')
    created = db.DateTimeProperty(auto_now_add=True)
    accepted = db.BooleanProperty(default=False)
    aspects = db.StringListProperty()
    expires = db.IntegerProperty() # any number less than 1000 => "oneshot"
    space = db.StringProperty()

class Item(db.Model):
    v = db.IntegerProperty(default=0, name='v')
    created = db.DateTimeProperty(auto_now_add=True)

    # public

class Locker(db.Model):
    v = db.IntegerProperty(default=0, name='v')
    created = db.DateTimeProperty(auto_now_add=True)
    control_ref = db.StringProperty()
    name = db.StringProperty()
    pay_cost = db.StringProperty(default='')
    read_ref = db.StringProperty()
    title = db.StringProperty()
    write_ref = db.StringProperty()
    write_aspects = db.StringListProperty()
    write_aspects_ref = db.StringProperty()
    remove_ref = db.StringProperty()

class OAuthClient(db.Model):
    v = db.IntegerProperty(default=0, name='v')
    created = db.DateTimeProperty(auto_now_add=True)
    description = db.TextProperty()
    name = db.StringProperty(indexed=False)
    owner = db.StringProperty()
    redirect = db.StringProperty(indexed=False)
    secret = db.StringProperty(indexed=False)
    url = db.StringProperty(indexed=False)

class OAuthCode(db.Model):
    v = db.IntegerProperty(default=0, name='v')
    created = db.DateTimeProperty(auto_now_add=True)
    client = db.StringProperty()
    expires = db.IntegerProperty()
    redirect = db.StringProperty(indexed=False)
    scope = db.StringProperty(indexed=False)

class OAuthRequest(db.Model):
    v = db.IntegerProperty()
    created = db.DateTimeProperty(auto_now_add=True)
    client = db.StringProperty()
    redirect = db.StringProperty(indexed=False)
    scope = db.StringProperty(indexed=False)
    state = db.StringProperty(indexed=False)
    type = db.StringProperty(indexed=False)

class OAuthToken(db.Model):
    v = db.IntegerProperty(default=0, name='v')
    created = db.DateTimeProperty(auto_now_add=True)
    client = db.StringProperty()
    expires = db.IntegerProperty()
    scope = db.StringProperty()

class Trend(db.Model):
    v = db.IntegerProperty(default=0, name='v')

class User(db.Model):
    v = db.IntegerProperty(default=0, name='v')
    created = db.DateTimeProperty(auto_now_add=True)

# ------------------------------------------------------------------------------
# Parallel Query
# ------------------------------------------------------------------------------

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

from google.appengine.api import urlfetch
from google.appengine.api.apiproxy_stub_map import UserRPC
from google.appengine.api.datastore import _ToDatastoreError, Entity, Query as RawQuery
from google.appengine.api.datastore_types import Key
from google.appengine.datastore import datastore_index
from google.appengine.datastore.datastore_pb import QueryResult, NextRequest
from google.appengine.ext.db import Query
from google.appengine.runtime.apiproxy_errors import ApplicationError

# Our BaseQuery subclass.
class BaseQuery(Query):

    _cursor = None
    _prev = 0

    def execute(
        self, limit, offset, value, set_result, add_callback, deadline,
        on_complete
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
            'RunQuery',
            raw_query._ToPb(self._limit, self._offset, self._limit),
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
                    *datastore_index.CompositeIndexForQuery(rpc.request)[1:-1]
                    )
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
                results = [
                    Key._FromPb(e.key()) for e in self._buffer[:self._limit]
                    ]
            else:
                results = [
                    Entity._FromPb(e) for e in self._buffer[:self._limit]
                    ]
                if self._model_class is not None:
                    from_entity = self._model_class.from_entity
                    results = [from_entity(e) for e in results]
                else:
                    results = [
                        class_for_kind(e.kind()).from_entity(e)
                        for e in results
                        ]
        finally:
            del self._buffer[:]

        if self.on_complete:
            results = self.on_complete(results)
        self.set_result(self._value, results)


# Parallel query object for doing Trust map queries.
class ParallelQuery(object):

    def __init__(
        self, model_class=None, keys_only=False, query_key=None,
        cache_duration=5*60, namespace='pq', notify=True, limit=50,
        offset=0, deadline=None, on_complete=None
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

# ------------------------------------------------------------------------------
# Static
# ------------------------------------------------------------------------------

HOST_SIZE = len(STATIC_HTTP_HOSTS)

if RUNNING_ON_GOOGLE_SERVERS:

    def STATIC(path, minifiable=1, secure=0, cache={}, host_size=HOST_SIZE):
        if STATIC.ssl_mode:
            secure = 1
        if (path, minifiable, secure) in cache:
            return cache[(path, minifiable, secure)]
        if minifiable:
            path, filename = split_path(path)
            filename, ext = filename.rsplit('.', 1)
            if (not path) or (path == '/'):
                path = '%s%s.min.%s' % (path, filename, ext)
            else:
                path = '%s/%s.min.%s' % (path, filename, ext)
        if secure:
            hosts = STATIC_HTTPS_HOSTS
        else:
            hosts = STATIC_HTTP_HOSTS
        return cache.setdefault((path, minifiable, secure), "%s%s%s?%s" % (
            hosts[int('0x' + md5(path).hexdigest(), 16) % host_size],
            STATIC_PATH, path, APPLICATION_TIMESTAMP
            ))

else:

    def STATIC(path, minifiable=1, secure=0):
        return '%s%s?%s' % (STATIC_PATH, path, APPLICATION_TIMESTAMP)

STATIC.ssl_mode = None

# ------------------------------------------------------------------------------
# Memcache
# ------------------------------------------------------------------------------

# Generate cache key/info for the render service call.
def cache_key_gen(ctx, cache_spec, name, *args, **kwargs):

    user = ''
    if cache_spec.get('user', True):
        user = ctx.get_user_id()
        if (not cache_spec.get('anon', True)) and not user:
            return

    if cache_spec.get('ignore_args', False):
        args = ()

    if cache_spec.get('ignore_kwargs', False):
        kwargs = {}

    key = sha1(
        "%r-%r-%r" % (user, args, sorted(kwargs.iteritems()))
        ).hexdigest()

    namespace = cache_spec.get('namespace', None)
    if namespace is None:
        namespace = name

    return key, namespace, cache_spec.get('time', 20)

# ------------------------------------------------------------------------------
# Service Utilities
# ------------------------------------------------------------------------------

SERVICE_DEFAULT_CONFIG = {
    'admin_only': False,
    'anon_access': False,
    'cache': False,
    'cache_key': cache_key_gen,
    'cache_spec': dict(namespace=None, time=10, player=True, anon=True),
    'ssl_only': False,
    'token_required': False
    }

# The ``register_service`` decorator is used to turn a function into a service.
def register_service(name, renderers, **config):
    def __register_service(function):
        __config = SERVICE_DEFAULT_CONFIG.copy()
        __config.update(config)
        SERVICE_REGISTRY[name] = (function, renderers, __config)
        return function
    return __register_service

# The default JSON renderer generates JSON-encoded output.
def json(ctx, **content):
    if 'Content-Type' not in ctx.response_headers:
        ctx.response_headers['Content-Type'] = 'application/json; charset=utf-8'
    callback = ctx.callback
    if callback:
        if not is_valid_jsonp_callback_value(callback):
            raise ValueError(
                "%r is not an accepted callback parameter." % callback
                )
        return '%s(%s)' % (callback, json_encode(content))
    return json_encode(content)

# ------------------------------------------------------------------------------
# HTTP Utilities
# ------------------------------------------------------------------------------

# Return an HTTP header date/time string.
def get_http_datetime(timestamp=None):
    if timestamp:
        if not isinstance(timestamp, datetime):
            timestamp = datetime.fromtimestamp(timestamp)
    else:
        timestamp = datetime.utcnow()
    return timestamp.strftime('%a, %d %B %Y %H:%M:%S GMT') # %m

# ------------------------------------------------------------------------------
# Context
# ------------------------------------------------------------------------------

# The ``Context`` class encompasses the HTTP request/response. An instance,
# specific to the current request, is passed in as the first parameter to all
# service calls.
class Context(object):

    callback = None
    end_pipeline = False

    def __init__(self, request_cookies, ssl_mode):
        self.status = (200, 'OK')
        self.raw_headers = []
        self.request_cookies = request_cookies
        self.response_cookies = {}
        self.response_headers = Headers(self.raw_headers)
        self.ssl_mode = ssl_mode

    def set_response_status(self, code, message=None):
        if not message:
            message = HTTP_STATUS_MESSAGES.get(code, ["Server Error"])[0]
        self.status = (code, message)

    def get_cookie(self, name, default=''):
        return self.request_cookies.get(name, default)

    def get_secure_cookie(self, name, timestamped=True):
        if name not in self.cookies:
            return
        return validate_tamper_proof_string(
            name, self.request_cookies[name], timestamped
            )

    def set_cookie(self, name, value, **kwargs):
        cookie = self.response_cookies.setdefault(name, {})
        cookie['value'] = value
        kwargs.setdefault('path', '/')
        for name, value in kwargs.iteritems():
            if value:
                cookie[name.lower()] = value

    def set_secure_cookie(
        self, name, value, duration=TAMPER_PROOF_DEFAULT_DURATION.seconds,
        **kwargs
        ):
        value = create_tamper_proof_string(name, value, duration)
        self.set_cookie(name, value, **kwargs)

    def append_to_cookie(self, name, value):
        cookie = self.response_cookies.setdefault(name, {})
        if 'value' in cookie:
            cookie['value'] = '%s:%s' % (cookie['value'], value)
        else:
            cookie['value'] = value

    def expire_cookie(self, name, **kwargs):
        if name in self.response_cookies:
            del self.cookies[name]
        kwargs.setdefault('path', '/')
        kwargs.update({'max_age': 0, 'expires': "Fri, 31-Dec-99 23:59:59 GMT"})
        self.set_cookie(name, 'deleted', **kwargs) # @/@ 'deleted' or just '' ?

    def set_to_not_cache_response(self):
        headers = self.response_headers
        headers['Expires'] = "Fri, 31 December 1999 23:59:59 GMT"
        headers['Last-Modified'] = get_http_datetime()
        headers['Cache-Control'] = "no-cache, must-revalidate" # HTTP/1.1
        headers['Pragma'] =  "no-cache"                        # HTTP/1.0

    def compute_url(self, *args, **kwargs):
        host = TENT_HTTPS_HOST if self.ssl_mode else TENT_HTTP_HOST
        return self.compute_url_for_host(host, *args, **kwargs)

    def compute_url_for_host(self, host, *args, **kwargs):
        out = host + '/' + '/'.join(
            arg.encode('utf-8') for arg in args
            )
        if kwargs:
            out += '?'
            _set = 0
            _l = ''
            for key, value in kwargs.items():
                key = urlquote(key).replace(' ', '+')
                if value is None:
                    value = ''
                if isinstance(value, list):
                    for val in value:
                        if _set: _l = '&'
                        out += '%s%s=%s' % (
                            _l, key,
                            urlquote(val.encode('utf-8')).replace(' ', '+')
                            )
                        _set = 1
                else:
                    if _set: _l = '&'
                    out += '%s%s=%s' % (
                        _l, key, urlquote(value.encode('utf-8')).replace(' ', '+')
                        )
                    _set = 1
        return out

    auth_token = None
    valid_auth_token = False

    _is_admin_user = None
    _is_logged_in = None
    _user = None
    _user_id = None
    _session = None

    def get_session(self):
        if self._session is not None:
            return self._session
        token = self.session_token
        if not token:
            return
        try:
            session = Session.all().filter('e =', token).get()
            if not session:
                return
            if session.expires and (session.expires < datetime.now()):
                session.delete()
                return
        except:
            return
        self._session = session
        return session

    def get_user(self):
        if self._user is not None:
            return self._user
        user_id = self.get_user_id()
        if user_id:
            self._user = User.get_by_id(user_id)
            return self._user
        self._user = None
        return None

    def get_user_id(self):
        if self._user_id is not None:
            return self._user_id
        session = self.get_session()
        if session:
            self._user_id = session.user
            self._is_logged_in = True
            return self._user_id
        self._user_id = None
        self._is_logged_in = False
        return None

    def is_logged_in(self):
        return False
        if self._is_logged_in is not None:
            return self._is_logged_in
        if self.get_user_id():
            return True

    def is_admin_user(self):
        if self._is_admin_user is not None:
            return self._is_admin_user
        user_id = self.get_user_id()
        if not user_id:
            self._is_admin_user = False
            return False
        identity = Identity.all().filter(
            'a =', 'google'
            ).filter(
            'c =', user_id
            ).get()
        if identity.id in SITE_ADMINS:
            self._is_admin_user = True
            return True
        self._is_admin_user = False
        return False

# ------------------------------------------------------------------------------
# Templating
# ------------------------------------------------------------------------------

TEMPLATE_BUILTINS = {
    'DEBUG': DEBUG,
    'STATIC': STATIC,
    'urlencode': urlencode,
    'urlquote': urlquote,
    'page_title': 'Espra Tent',
    'slot': 'content_slot'
    }

# ------------------------------------------------------------------------------
# App Runner
# ------------------------------------------------------------------------------

def handle_http_request(
    dict=dict, isinstance=isinstance, sys=sys, urlunquote=urlunquote,
    unicode=unicode
    ):

    # We copy ``os.environ`` just in case App Engine doesn't reset any changes
    # we might make to it. We take quite a defensive approach as the underlying
    # App Engine behaviour may change at any time and we shouldn't really take
    # too much for granted.
    env = dict(os.environ)

    # We redirect the stdout to a ``/dev/null`` like interface so that any
    # accidental prints by libraries don't end up as a part of the response.
    sys._boot_stdout = sys.stdout
    sys.stdout = DEVNULL
    write = sys._boot_stdout.write

    try:

        http_method = env['REQUEST_METHOD']
        ssl_mode = STATIC.ssl_mode = env.get('HTTPS') in SSL_FLAGS

        if http_method == 'OPTIONS':
            write(RESPONSE_OPTIONS)
            return

        if http_method not in SUPPORTED_HTTP_METHODS:
            write(RESPONSE_NOT_IMPLEMENTED)
            return

        args = [
            unicode(arg, 'utf-8', 'strict')
            for arg in env['PATH_INFO'].split('/') if arg
            ]

        host = env['HTTP_HOST']

        if host.startswith('www.'):
            service_name = 'root.www'
        elif host.endswith('.appspot.com') and not ssl_mode:
            raise Redirect(
                TENT_HTTPS_HOST + env['PATH_INFO'] +
                (env['QUERY_STRING'] and "?%s" % env['QUERY_STRING'] or "")
                )
        else:
            if args:
                service_name = args.pop(0)
            else:
                service_name = 'root'



        if service_name not in SERVICE_REGISTRY:
            logging.error("Service not found: %s" % service_name)
            raise NotFound

        kwargs = {}

        for part in [
            sub_part
            for part in env['QUERY_STRING'].lstrip('?').split('&')
            for sub_part in part.split(';')
            ]:
            if not part:
                continue
            part = part.split('=', 1)
            if len(part) == 1:
                value = None
            else:
                value = part[1]
            key = urlunquote(part[0].replace('+', ' '))
            if value:
                value = unicode(
                    urlunquote(value.replace('+', ' ')), 'utf-8', 'strict'
                    )
            else:
                value = None
            if key in kwargs:
                _val = kwargs[key]
                if isinstance(_val, list):
                    _val.append(value)
                else:
                    kwargs[key] = [_val, value]
                continue
            kwargs[key] = value

        cookies = {}
        cookie_data = env.get('HTTP_COOKIE', '')

        if cookie_data:
            _parsed = SimpleCookie()
            _parsed.load(cookie_data)
            for name in _parsed:
                cookies[name] = _parsed[name].value

        # Parse the POST body if it exists and is of a known content type.
        if http_method == 'POST':

            content_type = env.get('CONTENT-TYPE', '')
            if ';' in content_type:
                content_type = content_type.split(';', 1)[0]

            if content_type in VALID_REQUEST_CONTENT_TYPES:

                post_environ = env.copy()
                post_environ['QUERY_STRING'] = ''

                post_data = FieldStorage(
                    environ=post_environ, fp=environ['wsgi.input'],
                    keep_blank_values=True
                    ).list or []

                for field in post_data:
                    key = field.name
                    if field.filename:
                        value = field
                    else:
                        value = unicode(field.value, 'utf-8', 'strict')
                    if key in kwargs:
                        _val = kwargs[key]
                        if isinstance(_val, list):
                            _val.append(value)
                        else:
                            kwargs[key] = [_val, value]
                        continue
                    kwargs[key] = value

        ctx = Context(cookies, ssl_mode)

        if 'submit' in kwargs:
            del kwargs['submit']

        if 'callback' in kwargs:
            ctx.callback = kwargs.pop('callback')

        # check that there's a token and it validates

        if 0:
            signature = kwargs.pop('sig', None)
            if not signature:
                write(ERROR_HEADER)
                write(NOTAUTHORIZED)
                return
            if not validate_tamper_proof_string(
                'token', token, key=API_KEY, timestamped=True
                ):
                logging.info("Unauthorized API Access Attempt: %r", token)
                write(UNAUTH)
                return

            session_token = None

            # @/@ this is insecure ...

            if 'token' in special_kwargs:
                session_token = special_kwargs['__token__']
            else:
                if '__token__' in cookies:
                    session_token = cookies['__token__']

            if session_token and '__csrf__' in special_kwargs:
                session = self.get_session()
                if session and (special_kwargs['__csrf__'] == session.csrf_key):
                    self.valid_auth_token = True

        service, renderers, config = SERVICE_REGISTRY[service_name]

        if config['ssl_only'] and RUNNING_ON_GOOGLE_SERVERS and not ssl_mode:
            raise NotFound

        if config['admin_only'] and not ctx.is_admin_user():
            raise NotFound

        if config['token_required'] and not ctx.valid_auth_token:
            raise AuthError

        if config['cache']:
            cache_info = config['cache_key'](
                ctx, config['cache_spec'], service_name, *args, **kwargs
                )
            if cache_info is not None:
                cache_key, cache_namespace, cache_time = cache_info
                output = memcache.get(cache_key, cache_namespace)
                if output is not None:
                    raise HTTPContent(output)

        # Try and respond with the result of calling the service.
        if renderers and renderers[-1] == json:
            try:
                content = service(ctx, *args, **kwargs)
            except BaseHTTPError:
                raise
            except Exception, error:
                response = json(
                    ctx, error=str(error),
                    error_type=error.__class__.__name__
                    )
                write(RESPONSE_JSON_ERROR % (len(response), response))
                return
        else:
            content = service(ctx, *args, **kwargs)

        for renderer in renderers:
            if ctx.end_pipeline:
                break
            if not isinstance(content, dict):
                content = {
                    'content': content
                    }
            if isinstance(renderer, str):
                kwargs = TEMPLATE_BUILTINS.copy()
                kwargs.update(content)
                template = get_mako_template(renderer)
                content = template.render(ctx=ctx, **kwargs)
            else:
                content = renderer(ctx, **content)

        if isinstance(content, unicode):
            content = content.encode('utf-8')

        if config['cache'] and cache_info is not None:
            memcache.set(
                cache_key, output, cache_time, namespace=cache_namespace
                )

        raise HTTPContent(content)

    # Return the content.
    except HTTPContent, payload:

        content = payload.content

        if 'Content-Type' not in ctx.response_headers:
            ctx.response_headers['Content-Type'] = 'text/html; charset=utf-8'

        ctx.response_headers['Content-Length'] = str(len(content))

        # Figure out the HTTP headers for the response ``cookies``.
        cookie_output = SimpleCookie()
        for name, values in ctx.response_cookies.iteritems():
            name = str(name)
            cookie_output[name] = values.pop('value')
            cur = cookie_output[name]
            for key, value in values.items():
                if key == 'max_age':
                    key = 'max-age'
                if key not in COOKIE_KEY_NAMES:
                    continue
                cur[key] = value

        if cookie_output:
            raw_headers = ctx.raw_headers + [
                ('Set-Cookie', ck.split(' ', 1)[-1])
                for ck in str(cookie_output).split('\r\n')
                ]
        else:
            raw_headers = ctx.raw_headers

        write('Status: %d %s\r\n' % ctx.status)
        write('\r\n'.join('%s: %s' % (k, v) for k, v in raw_headers))
        write('\r\n\r\n')

        if http_method != 'HEAD':
            write(content)

    # Handle 404s.
    except NotFound:
        write(RESPONSE_404)
        return

    # Handle 401s.
    except AuthError:
        write(RESPONSE_401)
        return

    # Handle HTTP 301/302 redirects.
    except Redirect, redirect:
        if redirect.permanent:
            write(RESPONSE_301 % redirect.uri)
            return
        write(RESPONSE_302 % redirect.uri)
        return

    # Handle other HTTP response codes.
    except HTTPError, error:
        write("Status: %s %s\r\n\r\n"
              % (error.code, HTTP_STATUS_MESSAGES[error.code]))
        return

    except CapabilityDisabledError:
        write(RESPONSE_503)
        return

    # Log any errors and return an HTTP 500 response.
    except Exception, error:
        logging.critical(''.join(format_exception(*sys.exc_info())))
        if DEBUG:
            error = ''.join(html_format_exception())
        else:
            error = escape("%s: %s" % (error.__class__.__name__, error))
        response = ERROR_500_TRACEBACK % error
        write(RESPONSE_500_BASE % (len(response), response))
        return

    # We set ``sys.stdout`` back to the way we found it.
    finally:
        sys.stdout = sys._boot_stdout

# ------------------------------------------------------------------------------
# Monkey Patches
# ------------------------------------------------------------------------------

# The ``mako`` templating system uses ``beaker`` to cache segments and this
# needs various patches to make appropriate use of Memcache as a cache backend
# on App Engine.
#
# First, the App Engine memcache client needs to be setup as the ``memcache``
# module.
import google.appengine.api.memcache as memcache

sys.modules['memcache'] = memcache

import beaker.container
import beaker.ext.memcached

# And then the beaker ``Value`` object itself needs to be patched.
class Value(beaker.container.Value):

    def get_value(self):
        stored, expired, value = self._get_value()
        if not self._is_expired(stored, expired):
            return value

        if not self.createfunc:
            raise KeyError(self.key)

        v = self.createfunc()
        self.set_value(v)
        return v

beaker.container.Value = Value
beaker.ext.memcached.verify_directory = lambda x: None

# ------------------------------------------------------------------------------
# Mako
# ------------------------------------------------------------------------------

# The ``mako`` templating system is used in Tent. It offers a reasonably
# flexible engine with pretty decent performance.
class MakoTemplateLookup(object):

    default_template_args = {
        'format_exceptions': False,
        'error_handler': None,
        'disable_unicode': False,
        'output_encoding': 'utf-8',
        'encoding_errors': 'strict',
        'input_encoding': 'utf-8',
        'module_directory': None,
        'cache_type': 'memcached',
        'cache_dir': '.',
        'cache_url': 'memcached://',
        'cache_enabled': True,
        'default_filters': ['unicode', 'h'],  # will be shared across instances
        'buffer_filters': [],
        'imports': None,
        'preprocessor': None
        }

    templates_directory = 'template'

    def __init__(self, **kwargs):
        self.template_args = self.default_template_args.copy()
        self.template_args.update(kwargs)
        self._template_cache = {}
        self._template_mtime_data = {}

    if DEBUG:

        def get_template(self, uri, kwargs=None):

            filepath = join_path(self.templates_directory, uri + '.mako')
            if not exists(filepath):
                raise IOError("Cannot find template %s.mako" % uri)

            template_time = getmtime(filepath)

            if ((template_time <= self._template_mtime_data.get(uri, 0)) and
                ((uri, kwargs) in self._template_cache)):
                return self._template_cache[(uri, kwargs)]

            if kwargs:
                _template_args = self.template_args.copy()
                _template_args.update(dict(kwargs))
            else:
                _template_args = self.template_args

            template = MakoTemplate(
                uri=uri, filename=filepath, lookup=self, **_template_args
                )

            self._template_cache[(uri, kwargs)] = template
            self._template_mtime_data[uri] = template_time

            return template

    else:

        def get_template(self, uri, kwargs=None):

            if (uri, kwargs) in self._template_cache:
                return self._template_cache[(uri, kwargs)]

            filepath = join_path(self.templates_directory, uri + '.mako')
            if not exists(filepath):
                raise IOError("Cannot find template %s.mako" % uri)

            if kwargs:
                _template_args = self.template_args.copy()
                _template_args.update(dict(kwargs))
            else:
                _template_args = self.template_args

            template = MakoTemplate(
                uri=uri, filename=filepath, lookup=self, **_template_args
                )

            return self._template_cache.setdefault((uri, kwargs), template)

    def adjust_uri(self, uri, relativeto):
        return uri

mako_template_lookup = MakoTemplateLookup()
get_mako_template = mako_template_lookup.get_template

def call_mako_template(template, **kwargs):
    return template.render(**kwargs)

def render_mako_template(template_name, **kwargs):
    return get_mako_template(template_name).render(**kwargs)

# ------------------------------------------------------------------------------
# OAuth
# ------------------------------------------------------------------------------

# This implements a fairly conformant implementation of `OAuth 2.0
# <http://tools.ietf.org/html/draft-ietf-oauth-v2-08>`_.
#
# We diverge from the still-being-defined standard by not supporting HTTP Basic
# Authentication or the Authorization Request Header Field for the Client. It
# just complicates the implementation needlessly -- not to mention the fact that
# most OAuth 2.0 clients don't even support them!
#
# We also make life simpler for developers by providing additional informative
# error messages (with an HTTP 500 response status) and relaxing a few REQUIRED
# fields:
#
# * the ``type`` parameter doesn't need to be specified for the end-user
#   authorisation endpoint -- a default of ``web_server`` is assumed.
#
# * the ``grant_type`` parameter doesn't need to be specified for the token
#   endpoint -- a default of ``authorization_code`` is assumed.
#
# And like other implementations, e.g. Facebook, we use the more sensible
# ``access_token`` parameter name as opposed to ``oauth_token`` as the spec
# states for the purpose of validating access to protected resources. Hopefully
# the spec will be updated to reflect this before it is finalised.
@register_service('oauth.authorize', ['oauth.authorize', 'main'])
def oauth_authorize(
    ctx, client_id=None, redirect_uri=None, scope=None, state=None,
    type='web_server'
    ):

    if not client_id:
        raise ValueError("The client_id parameter was not specified.")

    if not redirect_uri:
        raise ValueError("The redirect_uri parameter was not specified.")

    if type not in ['web_server', 'user_agent']:
        raise ValueError("Unknown OAuth request type: %r." % type)

    client = OAuthClient.get_by_key_name(client_id)
    if not client:
        raise ValueError(
            "Could not find records for the given client_id, '%s'."
            % client_id
            )

    # We enforce strict equality of the redirect URI and the use of HTTPS if
    # it's a ``user_agent`` request. Otherwise, as long as the client hasn't
    # specified a redirect URI, it can pass in wherever it wants.
    if type == 'web_server':
        if client.redirect and redirect_uri != client.redirect:
            raise ValueError(
                "The registered redirect URI and the redirect_uri parameter "
                "do not match."
                )
    else:
        if RUNNING_ON_GOOGLE_SERVERS and not ctx.ssl_mode:
            raise ValueError(
                "A user_agent request can only be conducted over HTTPS."
                )
        if redirect_uri != client.redirect:
            raise ValueError(
                "The registered redirect URI and the redirect_uri parameter "
                "do not match."
                )
        if not redirect_uri.startswith('https://'):
            raise ValueError(
                "The redirect_uri needs to be HTTPS for a user_agent request."
                )

    scope = scope or ctx.get_user_id()
    request_id = hexlify(urandom(12))

    request = OAuthRequest(
        key_name=request_id,
        client=client_id,
        redirect=redirect_uri,
        scope=scope,
        state=state,
        type=type
        )

    db.put(request)

    return {
        'request_id': request_id,
        'request_type': type,
        'scopes': scope.split(' ')
        }

@register_service('oauth.redirect', [])
def oauth_redirect(grant, request_id, limit=None):

    request = OAuthRequest.get_by_key_name(request_id)
    if not request:
        raise ValueError(
            "Could not find records for the OAuth request."
            )

    if request.state:
        response = {'state': state}
    else:
        response = {}

    ua_request = request.type == 'user_agent'

    if grant == 'yes':
        token_id = hexlify(urandom(12))
        if limit:
            expiry = int(86400 * limit) # number of days
        if ua_request:
            token = OAuthToken(
                key_name=token_id, client=client_id, scope=request.scope
                )
            response['access_token'] = token_id
            if limit:
                response['expires_in'] = str(expiry)
                token.expires = time() + expiry
        else:
            token = OAuthCode(
                key_name=token_id, client=client_id,
                scope=request.scope, redirect=request.redirect
                )
            if limit:
                token.expires = time() + expiry
            response['code'] = token_id
    else:
        response['error'] = 'user_denied'

    uri = request.redirect
    if ua_request:
        if '#' not in uri:
            uri += '#'

    if '?' in uri:
        uri += '&%s' % urlencode(response)
    else:
        uri += '?%s' % urlencode(response)

    try:
        db.delete(request)
    finally:
        raise Redirect(uri)

@register_service('oauth.token', [json], ssl_only=1)
def oauth_token(
    ctx, client_id=None, client_secret=None, code=None,
    grant_type='authorization_code', redirect_uri=None,
    scope=None, username=None, password=None, assertion_type=None,
    assertion=None, refresh_token=None
    ):

    ctx.response_headers['Cache-Control'] = 'no-store'

    if not client_id:
        raise ValueError("The client_id parameter was not specified.")

    if not client_secret:
        raise ValueError("The client_secret parameter was not specified.")

    client = OAuthClient.get_by_key_name(client_id)
    if not client:
        raise ValueError(
            "Could not find records for the given client_id, '%s'."
            % client_id
            )

    if client_secret != client.secret:
        ctx.set_response_status(400)
        return {
            'error': 'invalid_client_credentials', 'error_type': 'OAuthError'
            }

    if grant_type == "authorization_code":
        if not code:
            raise ValueError("The code parameter was not specified.")
        code = OAuthCode.get_by_key_name(code)
        if not code:
            raise ValueError(
                "Could not find a record for the OAuth code."
                )
        if code.redirect != redirect_uri:
            ctx.set_response_status(400)
            return {
                'error': 'redirect_uri_mismatch', 'error_type': 'OAuthError'
                }
        now = int(time())
        expires = code.expires
        if expires and expires <= now:
            ctx.set_response_status(400)
            return {
                'error': 'authorization_expired', 'error_type': 'OAuthError'
                }
        token_id = hexlify(urandom(12))
        token = OAuthToken(
            key_name=token_id, client=client_id,
            expires=expires, scope=code.scope
            )
        response = {
            'access_token': token_id,
            'scope': code.scope
            }
        if expires:
            response['expires_in'] = expires - now
        return response
    elif grant_type == "user_basic_credentials":
        if not username:
            raise ValueError("The username parameter was not specified.")
        if not password:
            raise ValueError("The password parameter was not specified.")
        user = User.get_by_key_name(sha1(username).hexdigest())
        if not user:
            raise ValueError(
                "Could not find a record for the given username."
                )
        if user.password != password: # @/@ update to actual check later
            ctx.set_response_status(400)
            return {
                'error': 'invalid_user_credentials', 'error_type': 'OAuthError'
                }
        token_id = hexlify(urandom(12))
        token = OAuthToken(
            key_name=token_id, client=client_id, scope=username
            )
        return {
            'access_token': token_id,
            'scope': username
            }
    elif grant_type == "assertion":
        ctx.set_response_status(400)
        return {
            'error': 'invalid_assertion', 'error_type': 'OAuthError'
            }
    elif grant_type == "none":
        # The spec doesn't actually say what should happen in the case when
        # ``grant_type`` is ``none``, i.e. when it's acting on behalf of itself.
        pass
    elif grant_type == "refresh_token":
        if not refresh_token:
            raise ValueError("The refresh_token parameter was not specified.")
        ctx.set_response_status(400)
        return {
            'error': 'authorization_expired', 'error_type': 'OAuthError'
            }
    else:
        ctx.set_response_status(400)
        return {'error': 'unknown_format', 'error_type': 'OAuthError'}

# ------------------------------------------------------------------------------
# Main Function
# ------------------------------------------------------------------------------

# The main function has to be called ``main`` in order to be cached by the App
# Engine runtime framework.
if DEBUG == 2:

    from cProfile import Profile
    from pstats import Stats

    # This particular main function wraps the real runner within a profiler and
    # dumps the profiled statistics to the logs for later inspection.
    def main():

        profiler = Profile()
        profiler = profiler.runctx("handle_http_request()", globals(), locals())
        iostream = StringIO()

        stats = Stats(profiler, stream=iostream)
        stats.sort_stats("time")  # or cumulative
        stats.print_stats(80)     # 80 == how many to print

        # optional:
        # stats.print_callees()
        # stats.print_callers()

        logging.info("Profile data:\n%s", iostream.getvalue())

else:

    main = handle_http_request

# ------------------------------------------------------------------------------
# Misc
# ------------------------------------------------------------------------------

# find_all_words = compile_regex(
#     r'[^\s!\"#$%&()*+,-./:;<=>?@\[\\^_`{|}~]*'
#     ).findall

# find_all_words = compile_regex(r'(?u)\w+').findall # (?L)\w+

find_all_words = compile_regex(r'(?u)\w+').findall # (?L)\w+

XML_PATTERNS = (
    # cdata
    compile_regex(r'<!\[CDATA\[((?:[^\]]+|\](?!\]>))*)\]\]>').sub,
    # comment
    compile_regex(r'<!--((?:[^-]|(?:-[^-]))*)-->').sub,
    # pi
    compile_regex(r'<\?(\S+)[\t\n\r ]+(([^\?]+|\?(?!>))*)\?>').sub,
    # doctype
    compile_regex(r'(?m)(<!DOCTYPE[\t\n\r ]+\S+[^\[]+?(\[[^\]]+?\])?\s*>)').sub,
    # entities
    compile_regex(r'&[A-Za-z]+;').sub,
    # tag
    compile_regex(r'(?ms)<[^>]+>').sub,

    # re.compile(r'<[^<>]*>').sub,

    )

def harvest_words(
    text, strip=True, min_word_length=1,
    stop_words=STOP_WORDS, xml_patterns=XML_PATTERNS,
    find_words_in_text=find_all_words
    ):

    if strip:
        for replace_xml in xml_patterns:
            text = replace_xml(' ', text)

    text = text.lower() # @/@ handle i18n ??
    words = set(); add_word = words.add

    for word in find_words_in_text(text):

        while word.startswith("'"):
            word = word[1:]
        while word.endswith("'"):
            word = word[:-1]

        if (len(word) > min_word_length) and (word not in stop_words):
            add_word(word)

    return list(words)

# is digit or alpha
# multiple word groups
# .ti = 3 property

#replace_links = re.compile(r'[^\\]\[(.*?)\]', re.DOTALL).sub

replace_links = re.compile(r'\[(.*?)\]', re.DOTALL).sub

def escape(s):
    return s.replace(
        "&", "&amp;"
        ).replace(
        "<", "&lt;"
        ).replace(
        ">", "&gt;"
        ).replace(
        '"', "&quot;"
        )

text = """

hmz, [foo] and [bar|yes] oi [skype:foo|call me] hehe and okay or \[skype:ignore|yes] okay
yes, [http://google.com|google] and an attack [http://www.google.com " onclick="foo"|evil
link]

""" # emacs "

def handle_links(content):
    link = content.group(1)
    if '|' in link:
        uri, name = link.split('|', 1)
        uri = uri.strip()
        name = name.strip()
    else:
        uri = name = link.strip()
    if not match_valid_uri_scheme(uri):
        return content.group()
    return '<a href="%s">%s</a>' % (escape(uri), escape(name))

# print replace_links(handle_links, text)

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

def normalise(id, valid_chars=frozenset('abcdefghijklmnopqrstuvwxyz0123456789.-/')):
    r"normalise the id"
    id = '-'.join(id.replace('_', ' ').lower().split())

def foo():
    words = text.split()
    if len(words) > 5000:
        raise ValueError("The text is longer than 5,000 words!")

@register_service('test', [json])
def test(ctx, a, b):
    return "f: %s" % (a + b)

@register_service('moop', ['main'])
def root(ctx):
    return "Hello World."

@register_service('root', ['main'], anon_access=1)
def root(ctx):
    1/0
    raise CapabilityDisabledError
    return "Hello World."

# default to seeing trusted view of a locker according to the space link was
# clicked from

# ------------------------------------------------------------------------------
# Run In Standalone Mode
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
