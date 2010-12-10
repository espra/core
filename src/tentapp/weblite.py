# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""A sexy micro-framework for use with Google App Engine."""

import logging
import os
import sys

from BaseHTTPServer import BaseHTTPRequestHandler
from binascii import hexlify
from cgi import FieldStorage
from datetime import datetime
from hashlib import sha1
from md5 import md5
from os import urandom
from os.path import dirname, exists, join as join_path, getmtime
from posixpath import split as split_path
from re import compile as compile_regex, DOTALL, MULTILINE
from traceback import format_exception
from urllib import quote as urlquote, unquote as urlunquote
from urlparse import urljoin
from wsgiref.headers import Headers

from google.appengine.ext.blobstore import parse_blob_info
from google.appengine.runtime.apiproxy_errors import CapabilityDisabledError

# Extend the sys.path to include the parent and ``lib`` sibling directories.
sys.path.insert(0, dirname(__file__))
sys.path.insert(0, 'lib')

from cookie import SimpleCookie # note: this is our cookie and not Cookie...
from exception import html_format_exception
from mako.template import Template as MakoTemplate
from markdown import markdown
from pyutil.crypto import (
    create_tamper_proof_string, secure_string_comparison,
    validate_tamper_proof_string
    )

from pyutil.io import DEVNULL
from pyutil.jsonp import is_valid_jsonp_callback_value
from pyutil.sanitise import sanitise
from simplejson import dumps as json_encode

from webob import Request as WebObRequest # this import patches cgi.FieldStorage
                                          # to behave better for us too!

from config import (
    APPLICATION_TIMESTAMP, DEBUG, STATIC_PATH, SECURE_COOKIE_DURATION,
    SECURE_COOKIE_KEY, STATIC_HTTP_HOSTS, STATIC_HTTPS_HOSTS
    )

# ------------------------------------------------------------------------------
# Constants
# ------------------------------------------------------------------------------

COOKIE_KEY_NAMES = frozenset([
    'domain', 'expires', 'httponly', 'max-age', 'path', 'secure', 'version'
    ])

try:
    from config import SITE_CSS_FILE_BASE
except ImportError:
    SITE_CSS_FILE_BASE = '%ssite' % STATIC_PATH

SITE_CSS_PATH = "%s.css?%s" % (SITE_CSS_FILE_BASE, APPLICATION_TIMESTAMP)
SITE_CSS_IE_PATH = "%s.ie.css?%s" % (SITE_CSS_FILE_BASE, APPLICATION_TIMESTAMP)

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
""" % (SITE_CSS_PATH, SITE_CSS_PATH, SITE_CSS_IE_PATH)

ERROR_401 = ERROR_WRAPPER % """
  <div class="site-error">
    <h1>Not Authorized</h1>
    Your session may have expired or you may not have access.
    <ul>
      <li><a href="/">Return home</a></li>
      <li><a href="/login">Login</a></li>
    </ul>
  </div>
  """ # emacs"

ERROR_404 = ERROR_WRAPPER % """
  <div class="site-error">
    <h1>The item you requested was not found</h1>
    You may have clicked a dead link or mistyped the address. Some web addresses
    are case sensitive.
    <ul>
      <li><a href="/">Return home</a></li>
    </ul>
  </div>
  """ # emacs"

ERROR_500_BASE = ERROR_WRAPPER % """
  <div class="site-error">
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
  <div class="site-error">
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
    "Location: %s\r\n"
    )

RESPONSE_302 = (
    "Status: 302 Found\r\n"
    "Location: %s\r\n"
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
    STATIC_HTTP_HOSTS = STATIC_HTTPS_HOSTS = ['http://localhost:8080'] # filler

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
# Markdown-related Constants
# ------------------------------------------------------------------------------

SECURE_PREFIX = hexlify(urandom(18))

VALID_CSS_CLASSES = set([
    'c', 'cm', 'cp', 'cs', 'c1', 'err', 'g', 'gd', 'ge', 'gh', 'gi',
    'go', 'gp', 'gr', 'gs', 'gt', 'gu', 'k', 'kc', 'kd', 'kn', 'kp',
    'kr', 'kt', 'l', 'ld', 'm', 'mf', 'mh', 'mi', 'il', 'mo', 'bp',
    'n', 'na', 'nb', 'nc', 'nd', 'ne', 'nf', 'ni', 'nl', 'nn', 'no',
    'nx', 'nt', 'nv', 'vc', 'vg', 'vi', 'py', 'o', 'ow', 'p', 's',
    'sb', 'sc', 'sd', 'se', 'sh', 'si', 'sr', 'ss', 'sx', 's1', 's2',
    'w', 'x', 'css', 'rst', 'bash', 'yaml', 'codehilite'
    ])

# ------------------------------------------------------------------------------
# Markdown
# ------------------------------------------------------------------------------

# GitHub-flavoured markdown port adapted from http://gist.github.com/457617

replace_newlines = compile_regex(r'^[\w\<][^\n]*(\n+)', MULTILINE).sub
pre_extract = compile_regex(r'<pre>.*?</pre>', MULTILINE | DOTALL).sub
pre_insert = compile_regex(r'{gfm-extraction-([0-9a-f]{32})\}').sub

def gfm(text):

    # First, extract the pre blocks.
    extractions = {}

    def pre_extraction_callback(matchobj):
        digest = md5(matchobj.group(0) + str(len(extractions))).hexdigest()
        extractions[digest] = matchobj.group(0)
        return "{gfm-extraction-%s}" % digest

    text = pre_extract(pre_extraction_callback, text)

    # In very clear cases, let newlines become <br /> tags.
    def newline_callback(matchobj):
        if len(matchobj.group(1)) == 1:
            return matchobj.group(0).rstrip() + '  \n'
        else:
            return matchobj.group(0)

    text = replace_newlines(newline_callback, text)

    # And, finally, insert back the pre block extractions.
    def pre_insert_callback(matchobj):
        return '\n\n' + extractions[matchobj.group(1)]

    return pre_insert(pre_insert_callback, text)

def render_text(text, css=VALID_CSS_CLASSES, prefix=SECURE_PREFIX, trusted=0):
    if not isinstance(text, unicode):
        try:
            text = text.decode('utf-8')
        except UnicodeDecodeError:
            return "ERROR: text encoding."
    text = markdown(gfm(text), ['codehilite'])
    if trusted:
        return text
    return sanitise(text, secure_id_prefix=prefix, valid_css_classes=css)

# ------------------------------------------------------------------------------
# Static
# ------------------------------------------------------------------------------

HOST_SIZE = len(STATIC_HTTP_HOSTS)

if RUNNING_ON_GOOGLE_SERVERS:

    def STATIC(ctx, path, minify=0, secure=0, cache={}, host_size=HOST_SIZE):
        if ctx.ssl_mode:
            secure = 1
        if (path, minify, secure) in cache:
            return cache[(path, minify, secure)]
        if minify:
            path, filename = split_path(path)
            filename, ext = filename.rsplit('.', 1)
            if (not path) or (path == '/'):
                path = '%s%s.min.%s' % (path, filename, ext)
            else:
                path = '%s/%s.min.%s' % (path, filename, ext)
        if RUNNING_ON_GOOGLE_SERVERS:
            if secure:
                hosts = STATIC_HTTPS_HOSTS
            else:
                hosts = STATIC_HTTP_HOSTS
        else:
            return cache.setdefault(
                (path, minify, secure), (
                    "http://%s%s" %
                    (ctx.host, STATIC_PATH, path, APPLICATION_TIMESTAMP)
                    )
                )
        return cache.setdefault((path, minify, secure), "%s%s%s?%s" % (
            hosts[int('0x' + md5(path).hexdigest(), 16) % host_size],
            STATIC_PATH, path, APPLICATION_TIMESTAMP
            ))

else:

    def STATIC(ctx, path, minify=1, secure=0):
        return '%s://%s%s%s?%s' % (
            ctx.scheme, ctx.host, STATIC_PATH, path, APPLICATION_TIMESTAMP
            )

# ------------------------------------------------------------------------------
# Memcache
# ------------------------------------------------------------------------------

# Generate cache key/info for the render service call.
def cache_key_gen(ctx, cache_spec, name, *args, **kwargs):

    user = ''
    if cache_spec.get('user', True):
        user = ctx.username
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

try:
    from config import SSL_ONLY
except ImportError:
    SSL_ONLY = False

SERVICE_DEFAULT_CONFIG = {
    'admin': False,
    'anon': True,
    'blob': False,
    'cache': False,
    'cache_key': cache_key_gen,
    'cache_spec': dict(namespace=None, time=10, player=True, anon=True),
    'ssl': SSL_ONLY,
    'xsrf': False
    }

# The ``register_service`` decorator is used to turn a function into a service.
def register_service(name, renderers, **config):
    def __register_service(function):
        __config = SERVICE_DEFAULT_CONFIG.copy()
        __config.update(config)
        for _name in name.split():
            SERVICE_REGISTRY[_name] = (function, renderers, __config)
        return function
    return __register_service

# The default JSON renderer generates JSON-encoded output.
def json(ctx, **content):
    if 'Content-Type' not in ctx.response_headers:
        ctx.response_headers['Content-Type'] = 'application/json; charset=utf-8'
    callback = ctx.json_callback
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

    DEBUG = DEBUG
    STATIC = STATIC

    render_text = staticmethod(render_text)
    urlquote = staticmethod(urlquote)
    urlunquote = staticmethod(urlunquote)

    ajax_request = None
    json_callback = None
    end_pipeline = None

    _cookies_parsed = None
    _xsrf_token = None

    if SSL_ONLY:
        ssl_mode = 1
        scheme = 'https'

    def __init__(self, service, environ, ssl_mode, ssl_only=SSL_ONLY):
        self.service = service
        self.environ = environ
        self.host = environ['HTTP_HOST']
        self._status = (200, 'OK')
        self._raw_headers = []
        self._response_cookies = {}
        self.response_headers = Headers(self._raw_headers)
        if not ssl_only:
            self.ssl_mode = ssl_mode
            if ssl_mode:
                self.scheme = 'https'
            else:
                self.scheme = 'http'

    def set_response_status(self, code, message=None):
        if not message:
            message = HTTP_STATUS_MESSAGES.get(code, ["Server Error"])[0]
        self._status = (code, message)

    def _parse_cookies(self):
        cookies = {}
        cookie_data = self.environ.get('HTTP_COOKIE', '')
        if cookie_data:
            _parsed = SimpleCookie()
            _parsed.load(cookie_data)
            for name in _parsed:
                cookies[name] = _parsed[name].value
        self._request_cookies = cookies
        self._cookies_parsed = 1

    def get_cookie(self, name, default=''):
        if not self._cookies_parsed:
            self._parse_cookies()
        return self._request_cookies.get(name, default)

    def get_secure_cookie(self, name, key=SECURE_COOKIE_KEY, timestamped=True):
        if not self._cookies_parsed:
            self._parse_cookies()
        if name not in self._request_cookies:
            return
        return validate_tamper_proof_string(
            name, self._request_cookies[name], key, timestamped
            )

    def set_cookie(self, name, value, **kwargs):
        cookie = self._response_cookies.setdefault(name, {})
        cookie['value'] = value
        kwargs.setdefault('path', '/')
        if self.ssl_mode:
            kwargs.setdefault('secure', 1)
        for name, value in kwargs.iteritems():
            if value:
                cookie[name.lower()] = value

    def set_secure_cookie(
        self, name, value, key=SECURE_COOKIE_KEY,
        duration=SECURE_COOKIE_DURATION, **kwargs
        ):
        value = create_tamper_proof_string(name, value, key, duration)
        self.set_cookie(name, value, **kwargs)

    def append_to_cookie(self, name, value):
        cookie = self._response_cookies.setdefault(name, {})
        if 'value' in cookie:
            cookie['value'] = '%s:%s' % (cookie['value'], value)
        else:
            cookie['value'] = value

    def expire_cookie(self, name, **kwargs):
        if name in self._response_cookies:
            del self._response_cookies[name]
        kwargs.setdefault('path', '/')
        kwargs.update({'max_age': 0, 'expires': "Fri, 31-Dec-99 23:59:59 GMT"})
        self.set_cookie(name, '', **kwargs)

    def set_to_not_cache_response(self):
        headers = self.response_headers
        headers['Expires'] = "Fri, 31 December 1999 23:59:59 GMT"
        headers['Last-Modified'] = get_http_datetime()
        headers['Cache-Control'] = "no-cache, must-revalidate" # HTTP/1.1
        headers['Pragma'] =  "no-cache"                        # HTTP/1.0

    def compute_url(self, *args, **kwargs):
        return self.compute_url_for_host(self.host, *args, **kwargs)

    def compute_url_for_host(self, host, *args, **kwargs):
        out = self.scheme + '://' + host + '/' + '/'.join(
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

    @property
    def current_user(self):
        if not hasattr(self, '_current_user'):
            self._current_user = self.get_current_user()
        return self._current_user

    @property
    def is_admin(self):
        if not hasattr(self, '_is_admin'):
            self._is_admin = self.get_admin_status()
        return self._is_admin

    @property
    def site_url(self):
        if not hasattr(self, '_site_url'):
            self._site_url = self.scheme + '://' + self.host + '/'
        return self._site_url

    @property
    def url(self):
        if not hasattr(self, '_url'):
            self._url = self.site_url + self.environ['PATH_INFO']
        return self._url

    @property
    def url_with_qs(self):
        if not hasattr(self, '_url_with_qs'):
            env = self.environ
            query = env['QUERY_STRING']
            self._url_with_qs = (
                self.site_url + env['PATH_INFO'] + (
                    query and '?' or '') + query
                )
        return self._url_with_qs

    @property
    def username(self):
        if not hasattr(self, '_username'):
            self._username = self.get_username()
        return self._username

    @property
    def xsrf_token(self):
        if not self._xsrf_token:
            xsrf_token = self.get_secure_cookie('xsrf')
            if not xsrf_token:
                xsrf_token = hexlify(urandom(18))
                self.set_secure_cookie('xsrf', xsrf_token)
            self._xsrf_token = xsrf_token
        return self._xsrf_token

    from login import get_admin_status, get_current_user, get_username

    try:
        from login import get_login_url
    except ImportError:
        def get_login_url(self):
            return self.compute_url('login', return_to=self.url_with_qs)

# ------------------------------------------------------------------------------
# App Runner
# ------------------------------------------------------------------------------

def handle_http_request(
    dict=dict, isinstance=isinstance, sys=sys, urlunquote=urlunquote,
    unicode=unicode, get_response_headers=lambda: None
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
        ssl_mode = env.get('HTTPS') in SSL_FLAGS

        if http_method == 'OPTIONS':
            write(RESPONSE_OPTIONS)
            return

        if http_method not in SUPPORTED_HTTP_METHODS:
            write(RESPONSE_NOT_IMPLEMENTED)
            return

        _path_info = env['PATH_INFO']
        if isinstance(_path_info, unicode):
            _args = [arg for arg in _path_info.split(u'/') if arg]
        else:
            _args = [
                unicode(arg, 'utf-8', 'strict')
                for arg in _path_info.split('/') if arg
                ]

        if _args:
            service_name = _args[0]
            args = _args[1:]
        else:
            service_name = 'root'
            args = ()

        if service_name not in SERVICE_REGISTRY:
            router = handle_http_request.router
            if router:
                _service_info = router(env, _args)
                if not _service_info:
                    logging.error("No service found for: %s" % _path_info)
                    raise NotFound
                service_name, args = _service_info
            else:
                logging.error("Service not found: %s" % service_name)
                raise NotFound

        service, renderers, config = SERVICE_REGISTRY[service_name]
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

        # Parse the POST body if it exists and is of a known content type.
        if http_method == 'POST':

            content_type = env.get('CONTENT-TYPE', '')
            if ';' in content_type:
                content_type = content_type.split(';', 1)[0]

            if content_type in VALID_REQUEST_CONTENT_TYPES:

                post_environ = env.copy()
                post_environ['QUERY_STRING'] = ''

                post_data = FieldStorage(
                    environ=post_environ, fp=sys.stdin,
                    keep_blank_values=True
                    ).list or []

                for field in post_data:
                    key = field.name
                    if field.filename:
                        if config['blob']:
                            value = parse_blob_info(field)
                        else:
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

        ctx = Context(service_name, env, ssl_mode)

        def get_response_headers():
            # Figure out the HTTP headers for the response ``cookies``.
            cookie_output = SimpleCookie()
            for name, values in ctx._response_cookies.iteritems():
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
                raw_headers = ctx._raw_headers + [
                    ('Set-Cookie', ck.split(' ', 1)[-1])
                    for ck in str(cookie_output).split('\r\n')
                    ]
            else:
                raw_headers = ctx._raw_headers
            return '\r\n'.join('%s: %s' % (k, v) for k, v in raw_headers)

        if 'submit' in kwargs:
            del kwargs['submit']

        if 'callback' in kwargs:
            ctx.json_callback = kwargs.pop('callback')

        if env.get('HTTP_X_REQUESTED_WITH') == 'XMLHttpRequest':
            ctx.ajax_request = 1

        if config['ssl'] and RUNNING_ON_GOOGLE_SERVERS and not ssl_mode:
            raise NotFound

        if config['xsrf']:
            if 'xsrf' not in kwargs:
                raise AuthError("XSRF token not present.")
            provided_xsrf = kwargs.pop('xsrf')
            if not secure_string_comparison(provided_xsrf, ctx.xsrf_token):
                raise AuthError("XSRF tokens do not match.")

        if config['admin'] and not ctx.is_admin:
            raise NotFound

        if (not config['anon']) and (not ctx.current_user):
            raise AuthError("You need to be logged in.")

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
                logging.critical(''.join(format_exception(*sys.exc_info())))
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
                template = get_mako_template(renderer)
                content = template.render(
                    ctx=ctx, STATIC=ctx.STATIC, **content
                    )
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

        write('Status: %d %s\r\n' % ctx._status)
        write(get_response_headers())
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
        else:
            write(RESPONSE_302 % redirect.uri)
        headers = get_response_headers()
        if headers:
            write(headers)
            write('\r\n\r\n')
            return
        write('\r\n')
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
            error = ''
        response = ERROR_500_TRACEBACK % error
        write(RESPONSE_500_BASE % (len(response), response))
        return

    # We set ``sys.stdout`` back to the way we found it.
    finally:
        sys.stdout = sys._boot_stdout

handle_http_request.router = None

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
        'default_filters': ['unicode'],  # will be shared across instances
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
# Main Function
# ------------------------------------------------------------------------------

# The main function has to be called ``main`` in order to be cached by the App
# Engine runtime framework.
if DEBUG == 2:

    from cProfile import Profile
    from pstats import Stats
    from StringIO import StringIO

    def main():
        profiler = Profile()
        profiler = profiler.runctx("handle_http_request()", globals(), locals())
        iostream = StringIO()
        stats = Stats(profiler, stream=iostream)
        stats.sort_stats("time")
        stats.print_stats(80)
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
    stop_words=set(), xml_patterns=XML_PATTERNS,
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

def escape(s):
    return s.replace("&", "&amp;").replace("<", "&lt;").replace(
        ">", "&gt;").replace('"', "&quot;")

