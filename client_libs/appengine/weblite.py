# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""WebLite Framework."""

import logging
import os
import sys

from BaseHTTPServer import BaseHTTPRequestHandler
from cgi import FieldStorage, parse_qsl as parse_query_string
from datetime import datetime
from hashlib import sha1
from os.path import dirname, exists, join as join_path, getmtime, realpath
from pprint import pprint
from re import compile as compile_regex
from StringIO import StringIO
from time import time
from traceback import format_exception
from types import FunctionType
from urllib import urlencode, quote as urlquote, unquote as urlunquote
from urlparse import urljoin
from wsgiref.headers import Headers

# ------------------------------------------------------------------------------
# extend sys.path
# ------------------------------------------------------------------------------

APP_ROOT = dirname(realpath(__file__))
THIRD_PARTY_LIBS_PATH = join_path(APP_ROOT, 'third_party')

if THIRD_PARTY_LIBS_PATH not in sys.path:
    sys.path.insert(0, THIRD_PARTY_LIBS_PATH)

# ------------------------------------------------------------------------------
# import other libraries
# ------------------------------------------------------------------------------

from cookie import SimpleCookie # note: this is our cookie and not Cookie...
from demjson import encode as json_encode
from genshi.core import Markup
from genshi.template import MarkupTemplate, NewTextTemplate
from mako.template import Template as MakoTemplate
from webob import Request as WebObRequest # this import patches cgi.FieldStorage
                                          # to behave better for us too!

from google.appengine.api import users

from pyutil.exception import HTMLExceptionFormatter
from pyutil.validation import validate

from pyutil.crypto import (
    create_tamper_proof_string, validate_tamper_proof_string
    )

from model import Player

# ------------------------------------------------------------------------------
# default base konfig
# ------------------------------------------------------------------------------

try:
    from updated import APPLICATION_TIMESTAMP
except ImportError:
    APPLICATION_TIMESTAMP = time()

DEFAULT_TEMPLATE_MODE = 'mako'
LIVE_DEPLOYMENT = False
SITE_HTTP_URL = None
STATIC_PATH = '/.static/'
STATIC_HOSTS = None

DEBUG = None
SITE_MAIN_TEMPLATE = None
STATIC = None
TAMPER_PROOF_DEFAULT_DURATION = None
SITE_ADMINS = set()

# ------------------------------------------------------------------------------
# default error templates
# ------------------------------------------------------------------------------

ERROR_401_TEMPLATE = '<div class="error"><h1>YouPlease log in</div>'

ERROR_404_TEMPLATE = """
  <div class="error">
    <h1>The page you requested was not found</h1>
    You may have clicked a dead link or mistyped the address. Some web addresses
    are case sensitive.
    <ul>
      <li><a href="/">Return home</a></li>
      <li><a href="#">Go back to the previous page</a></li>
    </ul>
  </div>
  """ # emacs"

ERROR_500_TEMPLATE = """
  <div class="error">
    <h1>Ooops, something went wrong!</h1>
    There was an application error. This has been logged and will be resolved as
    soon as possible.
    <ul>
      <li><a href="/">Return home</a></li>
      <li><a href="#">Go back to the previous page</a></li>
    </ul>
    <div class="traceback">%s</div>
  </div>
  """ # emacs"

NETWORK_ERROR_MESSAGE = u"""
  Please take a deep breath, visualise floating in a warm sea ...
  then, when mentally refreshed, try again.
  """

# ------------------------------------------------------------------------------
# interpolated konfig
# ------------------------------------------------------------------------------

_GENERATED_CONFIG_TEMPLATE = """

from datetime import timedelta
from posixpath import split as split_path

APPLICATION_ID = os.environ.get('APPLICATION_ID')
SITE_MAIN_TEMPLATE = '%s:site' % DEFAULT_TEMPLATE_MODE

if os.environ.get('SERVER_SOFTWARE', '').startswith('Google'):

    RUNNING_ON_GOOGLE_SERVERS = True
    DEBUG = False

    if SITE_HTTP_URL is None:
        if LIVE_DEPLOYMENT:
            SITE_HTTP_URL = 'http://www.%s' % SITE_DOMAIN
        else:
            SITE_HTTP_URL = 'http://devsite.%s' % SITE_DOMAIN

    SITE_HTTPS_URL = 'https://%s.appspot.com' % APPLICATION_ID

else:

    RUNNING_ON_GOOGLE_SERVERS = False

    try:
        import emulate_production_mode
    except:
        DEBUG = True
    else:
        DEBUG = False

    SITE_HTTP_URL = 'http://localhost:8080'
    SITE_HTTPS_URL = SITE_HTTP_URL
    STATIC_HOSTS = [SITE_HTTP_URL]

if DEBUG:

    STATIC_HOST = SITE_HTTP_URL

    def STATIC(path, minifiable=False, secure=False):
        return '/.static/%s?%s' % (path, time())

else:

    if STATIC_HOSTS is None:
        STATIC_HOST = 'http://static1.%s' % SITE_DOMAIN
        STATIC_HOSTS = [
            'http://static1.%s' % SITE_DOMAIN,
            'http://static2.%s' % SITE_DOMAIN,
            'http://static3.%s' % SITE_DOMAIN
            ]
    else:
        STATIC_HOST = STATIC_HOSTS[0]

    def STATIC(path, minifiable=False, secure=False, cache={}, len_hosts=len(STATIC_HOSTS)):
        if STATIC.ctx and STATIC.ctx.ssl_mode:
            secure = True
        if (path, minifiable, secure) in cache:
            return cache[(path, minifiable, secure)]
        if minifiable:
            path, filename = split_path(path)
            if (not path) or (path == '/'):
                path = '%smin.%s' % (path, filename)
            else:
                path = '%s/min.%s' % (path, filename)
        if secure:
            return cache.setdefault((path, minifiable, secure), "%s%s%s?%s" % (
                SITE_HTTPS_URL, STATIC_PATH, path, APPLICATION_TIMESTAMP
                ))
        return cache.setdefault((path, minifiable, secure), "%s%s%s?%s" % (
            STATIC_HOSTS[int('0x' + sha1(path).hexdigest(), 16) % len_hosts],
            STATIC_PATH, path, APPLICATION_TIMESTAMP
            ))

    STATIC.ctx = None

CLEANUP_BATCH_SIZE = 100
EXPIRATION_WINDOW = timedelta(seconds=60*60*1) # 1 hour

"""

# ------------------------------------------------------------------------------
# load the siteconfig file
# ------------------------------------------------------------------------------

_site_config_file = open('config.py')
_site_config = _site_config_file.read() % {
    'include_base_config': _GENERATED_CONFIG_TEMPLATE
    }
_site_config_file.close()

exec(_site_config)

# ------------------------------------------------------------------------------
# exseptions
# ------------------------------------------------------------------------------

class Error(Exception):
    """Base Weblite Exception."""

class Redirect(Error):
    """
    Redirection Error.

    This is used to handle both internal and HTTP redirects.
    """

    def __init__(self, uri, method=None, permanent=False):
        self.uri = uri
        self.method = method
        self.permanent = permanent

class UnAuth(Error):
    """Unauthorised."""

class NotFound(Error):
    """404."""

class ServiceNotFound(NotFound):
    """Service 404."""

def format_traceback(type, value, traceback, limit=200):
    """Pretty print a traceback in HTML format."""

    return HTMLExceptionFormatter(limit).formatException(type, value, traceback)

# ------------------------------------------------------------------------------
# i/o helpers
# ------------------------------------------------------------------------------

class DevNull:
    """Provide a file-like interface emulating /dev/null."""

    def __call__(self, *args, **kwargs):
        pass

    def flush(self):
        pass

    def log(self, *args, **kwargs):
        pass

    def write(self, input):
        pass

DEVNULL = DevNull()

# ------------------------------------------------------------------------------
# wsgi
# ------------------------------------------------------------------------------

SSL_ENABLED_FLAGS = frozenset(['yes', 'on', '1'])

def run_wsgi_app(application, ssl_enabled_flags=SSL_ENABLED_FLAGS):
    """Run a WSGI ``application`` inside a CGI environment."""

    environ = dict(os.environ)

    environ['wsgi.errors'] = sys.stderr
    environ['wsgi.input'] = sys.stdin
    environ['wsgi.multiprocess'] = False
    environ['wsgi.multithread'] = False
    environ['wsgi.run_once'] = True
    environ['wsgi.version'] = (1, 0)

    if environ.get('HTTPS') in ssl_enabled_flags:
        environ['wsgi.url_scheme'] = 'https'
    else:
        environ['wsgi.url_scheme'] = 'http'

    sys._boot_stdout = sys.stdout
    sys.stdout = DEVNULL
    write = sys._boot_stdout.write

    try:
        result = application(environ, start_response)
        if result is not None:
            for data in result:
                write(data)
    finally:
        sys.stdout = sys._boot_stdout

def start_response(status, response_headers, exc_info=None):
    """Initialise a WSGI response with the given status and headers."""

    if exc_info:
        try:
            raise exc_info[0], exc_info[1], exc_info[2]
        finally:
            exc_info = None # bye-bye sirkular ref

    write = sys._boot_stdout.write
    write("Status: %s\r\n" % status)

    for name, val in response_headers:
        write("%s: %s\r\n" % (name, val))

    write('\r\n')

    return write

# ------------------------------------------------------------------------------
# general http util
# ------------------------------------------------------------------------------

def get_http_datetime(timestamp=None):
    """Return an HTTP header date/time string."""

    if timestamp:
        if not isinstance(timestamp, datetime):
            timestamp = datetime.fromtimestamp(timestamp)
    else:
        timestamp = datetime.utcnow()

    return timestamp.strftime('%a, %d %B %Y %H:%M:%S GMT') # %m

# ------------------------------------------------------------------------------
# http response
# ------------------------------------------------------------------------------

HTTP_STATUS_MESSAGES = BaseHTTPRequestHandler.responses

class Response(object):
    """HTTP Response."""

    def __init__(self):
        self.cookies = {}
        self.status = []
        self.stream = StringIO()
        self.write = self.stream.write
        self.raw_headers = []
        self.headers = Headers(self.raw_headers)
        self.set_header = self.headers.__setitem__

    def set_response_status(self, code, message=None):
        if not message:
            if not HTTP_STATUS_MESSAGES.has_key(code):
                raise Error('Invalid HTTP status code: %d' % code)
            message = HTTP_STATUS_MESSAGES[code][0]
        self.status[:] = (code, message)

    def clear_response(self):
        self.stream.seek(0)
        self.stream.truncate(0)

    def set_status_and_clear_response(self, code):
        self.set_response_status(code)
        self.clear_response()

    def set_cookie(self, name, value, **kwargs):
        cookie = self.cookies.setdefault(name, {})
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
        cookie = self.cookies.setdefault(name, {})
        if 'value' in cookie:
            cookie['value'] = '%s:%s' % (cookie['value'], value)
        else:
            cookie['value'] = value

    def expire_cookie(self, name, **kwargs):
        if name in self.cookies:
            del self.cookies[name]
        kwargs.setdefault('path', '/')
        kwargs.update({'max_age': 0, 'expires': "Fri, 31-Dec-99 23:59:59 GMT"})
        self.set_cookie(name, 'deleted', **kwargs) # @/@ 'deleted' or just '' ?

    def set_to_not_cache_response(self):
        headers = self.headers
        headers['Expires'] = "Fri, 31 December 1999 23:59:59 GMT"
        headers['Last-Modified'] = get_http_datetime()
        headers['Cache-Control'] = "no-cache, must-revalidate" # HTTP/1.1
        headers['Pragma'] =  "no-cache"                        # HTTP/1.0

# ------------------------------------------------------------------------------
# kookie support
# ------------------------------------------------------------------------------

COOKIE_KEY_NAMES = frozenset([
    'domain', 'expires', 'httponly', 'max-age', 'path', 'secure', 'version'
    ])

def get_cookie_headers_to_write(cookies, valid_keys=COOKIE_KEY_NAMES):
    """Return HTTP response headers for the given ``cookies``."""

    output = SimpleCookie()

    for name, values in cookies.iteritems():

        name = str(name)
        output[name] = values.pop('value')
        cur = output[name]

        for key, value in values.items():
            if key == 'max_age':
                key = 'max-age'
            # elif key == 'comment':
            #     # encode rather than throw an exception
            #     v = quote(v.encode('utf-8'), safe="/?:@&+")
            if key not in valid_keys:
                continue
            cur[key] = value

    return str(output)

# ------------------------------------------------------------------------------
# kore wsgi applikation
# ------------------------------------------------------------------------------

HTTP_HANDLERS = {}
register_http_handler = HTTP_HANDLERS.__setitem__

def Application(environ, start_response, handlers=HTTP_HANDLERS):
    """Core WSGI Application."""

    env_copy = dict(environ)
    response = Response()

    http_method = environ['REQUEST_METHOD']

    try:

        if http_method in handlers:
            if DEBUG:
                response.headers['Cache-Control'] = 'no-cache' # @/@
            response.set_response_status(200)
            handlers[http_method](environ, response)
        else:
            response.set_status_and_clear_response(501)

    except Redirect, redirect:

        # internal redirekt
        if redirect.method:
            env_copy['REQUEST_METHOD'] = redirect.method
            if '?' in redirect.uri:
                (env_copy['PATH_INFO'],
                 env_copy['QUERY_STRING']) = redirect.uri.split('?', 1)
            else:
                env_copy['PATH_INFO'] = redirect.uri
                env_copy['QUERY_STRING'] = ''
            return Application(
                env_copy, start_response, handlers
                )

        # external redirekt
        if redirect.permanent:
            response.set_response_status(301)
        else:
            response.set_response_status(302)

        response.headers['Location'] = str(
            urljoin('', redirect.uri)
            )
        response.clear_response()

    except Exception:

        response.set_status_and_clear_response(500)
        lines = ''.join(format_exception(*sys.exc_info()))
        logging.error(lines)

        if DEBUG:
            response.headers['Content-Type'] = 'text/plain'
            response.write(lines)

        return

    content = response.stream.getvalue()

    if isinstance(content, unicode):
        content = content.encode('utf-8')
    elif response.headers.get('Content-Type', '').endswith('; charset=utf-8'):
        try:
            content.decode('utf-8')
        except UnicodeError, error:
            logging.warning('Response written is not UTF-8: %s', error)

    response.headers['Content-Length'] = str(len(content))

    raw_headers = response.raw_headers + [
        ('Set-Cookie', ck.split(' ', 1)[-1])
        for ck in get_cookie_headers_to_write(response.cookies).split('\r\n')
        ]

    write = start_response('%d %s' % tuple(response.status), raw_headers)

    if http_method != 'HEAD':
        write(content)

    response.stream.close()

    return [''] # @/@ why do we have this instead of None ??

# ------------------------------------------------------------------------------
# http request objekt
# ------------------------------------------------------------------------------

VALID_CHARSETS = frozenset(['utf-8'])
find_charset = compile_regex(r'(?i);\s*charset=([^;]*)').search

VALID_REQUEST_CONTENT_TYPES = frozenset([
    '', 'application/x-www-form-urlencoded', 'multipart/form-data'
    ])

class Context(object):
    """Context API -- Encompasses an HTTP Request."""

    Redirect = Redirect
    UnAuth = UnAuth
    NotFound = NotFound

    auth_token = None
    post_data = None
    request_charset = 'utf-8'
    return_render = None
    valid_auth_token = False

    _current_player = None
    _current_player_id = None
    _current_session = None
    _is_admin_user = None
    _is_logged_in = None

    create_google_logout_url = staticmethod(users.create_logout_url)
    create_google_login_url = staticmethod(users.create_login_url)
    is_current_user_admin = staticmethod(users.is_current_user_admin)
    get_current_google_user = staticmethod(users.get_current_user)

    def __init__(
        self, environ, response, parse_query_string=parse_query_string,
        find_charset=find_charset, urlunquote=urlunquote
        ):

        self.request_method = environ['REQUEST_METHOD']

        self.environ = environ
        self.response = response

        self.append_to_cookie = response.append_to_cookie
        self.expire_cookie = response.expire_cookie
        self.set_cookie = response.set_cookie
        self.set_secure_cookie = response.set_secure_cookie
        self.clear_response = response.clear_response
        self.response_headers = response.headers
        self.set_response_header = response.set_header
        self.set_response_status = response.set_response_status
        self.set_status_and_clear_response = response.set_status_and_clear_response
        self.set_to_not_cache_response = response.set_to_not_cache_response

        path = environ['PATH_INFO']
        query = environ['QUERY_STRING']
        scheme = environ['wsgi.url_scheme']
        port = environ['SERVER_PORT']

        self.req_scheme = scheme
        self.ssl_mode = (scheme == 'https')
        self.site_uri = (
            scheme + '://' + environ['SERVER_NAME'] + ((
                (scheme == 'http' and port != '80') or 
                (scheme == 'https' and port != '443')
                ) and ':%s' % port or '')
            )

        self.uri = self.site_uri + path
        self.uri_with_qs = self.uri + (query and '?' or '') + query

        request_content_type = environ.get('CONTENT-TYPE', '')

        if request_content_type:
            match = find_charset(request_content_type)
            if match:
                match = match.group(1).lower()
                if match in VALID_CHARSETS:
                    self.request_charset = match

        self.request_args = tuple(
            unicode(arg, self.request_charset, 'strict')
            for arg in path.split('/') if arg
            )

        self.request_flags = flags = set()
        self.request_kwargs = kwargs = {}
        self.special_kwargs = special_kwargs = {}

        _val = None

        for part in [
            sub_part
            for part in query.lstrip('?').split('&')
            for sub_part in part.split(';')
            ]:
            if not part:
                continue
            part = part.split('=', 1)
            if len(part) == 1:
                flags.add(urlunquote(part[0].replace('+', ' ')))
                continue
            key = urlunquote(part[0].replace('+', ' '))
            value = part[1]
            if value:
                value = unicode(
                    urlunquote(value.replace('+', ' ')),
                    self.request_charset, 'strict'
                    )
            else:
                value = None
            if key.startswith('__') and key.endswith('__'):
                _kwargs = special_kwargs
            else:
                _kwargs = kwargs
            if key in _kwargs:
                _val = _kwargs[key]
                if isinstance(_val, list):
                    _val.append(value)
                else:
                    _kwargs[key] = [_val, value]
                continue
            _kwargs[key] = value

        self.cookies = cookies = {}
        cookie_data = environ.get('HTTP_COOKIE', '')

        if cookie_data:
            _parsed = SimpleCookie()
            _parsed.load(cookie_data)
            for name in _parsed:
                cookies[name] = _parsed[name].value

        if self.request_method == 'POST':

            if ';' in request_content_type:
                request_content_type = request_content_type.split(';', 1)[0]

            if request_content_type in VALID_REQUEST_CONTENT_TYPES:

                post_environ = environ.copy()
                post_environ['QUERY_STRING'] = ''

                post_data = self.post_data = FieldStorage(
                    environ=post_environ, fp=environ['wsgi.input'],
                    keep_blank_values=True
                    ).list

                if post_data:

                    for field in post_data:
                        key = field.name
                        if field.filename:
                            value = field
                        else:
                            value = unicode(field.value, self.request_charset, 'strict')
                        if key.startswith('__') and key.endswith('__'):
                            _kwargs = special_kwargs
                        else:
                            _kwargs = kwargs
                        if key in _kwargs:
                            _val = _kwargs[key]
                            if isinstance(_val, list):
                                _val.append(value)
                            else:
                                _kwargs[key] = [_val, value]
                            continue
                        _kwargs[key] = value

        session_token = None

        # @/@ this is insecure ...

        if '__token__' in special_kwargs:
            # if not self.ssl_mode:
            #     raise Error("It is insecure to use an auth token over a non-SSL connection.")
            session_token = special_kwargs['__token__']
        else:
            if '__token__' in cookies:
                session_token = cookies['__token__']

        self.session_token = session_token

        if session_token and '__csrf__' in special_kwargs:
            session = self.get_current_session()
            if session and (special_kwargs['__csrf__'] == session.csrf_key):
                self.valid_auth_token = True

        self._pre_render_hooks = []
        self._post_render_hooks = []
        self.render_stack = []

    def _set_pre_render_hook(self, function):
        if function not in self._pre_render_hooks:
            self._pre_render_hooks.append(function)

    pre_render_hook = property(
        lambda self: self._pre_render_hooks,
        _set_pre_render_hook
        )

    def _set_post_render_hook(self, function):
        if function not in self._post_render_hooks:
            self._post_render_hooks.append(function)

    post_render_hook = property(
        lambda self: self._post_render_hooks,
        _set_post_render_hook
        )

    @property
    def service_name(self):
        if self.render_stack:
            return self.render_stack[-1][0].name
        return None

    @property
    def render_format(self):
        if self.render_stack:
            return self.render_stack[-1][1]
        return None

    def compute_site_uri(self, *args, **kwargs):
        return self.compute_site_uri_for_host(self.site_uri, *args, **kwargs)

    def compute_site_uri_for_host(self, host, *args, **kwargs):

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

    def get_cookie(self, name, default=''):
        return self.cookies.get(name, default)

    def get_secure_cookie(self, name, timestamped=True):
        if name not in self.cookies:
            return
        return validate_tamper_proof_string(
            name, self.cookies[name], timestamped
            )

    def get_current_session(self):
        if self._current_session is not None:
            return self._current_session
        token = self.session_token
        if not token:
            self._current_session = None
            return None
        try:
            session = Session.all().filter('e =', token).get()
            if session.expires and (session.expires < datetime.now()):
                session.delete()
                session = None
        except:
            session = None
        self._current_session = session
        return session

    def get_current_player(self):
        if self._current_player is not None:
            return self._current_player
        player_id = self.get_current_player_id()
        if player_id:
            self._current_player = Player.get_by_id(player_id)
            return self._current_player
        self._current_player = None
        return None

    def get_current_player_id(self):
        if self._current_player_id is not None:
            return self._current_player_id
        session = self.get_current_session()
        if session:
            self._current_player_id = session.player
            self._is_logged_in = True
            return self._current_player_id
        self._current_player_id = None
        self._is_logged_in = False
        return None

    def is_logged_in(self):
        if self._is_logged_in is not None:
            return self._is_logged_in
        if self.get_current_player_id():
            return True

    def is_admin_user(self):

        if self._is_admin_user is not None:
            return self._is_admin_user

        player_id = self.get_current_player_id()

        if not player_id:
            self._is_admin_user = False
            return False

        identity = Identity.all().filter(
            'a =', 'google'
            ).filter(
            'c =', player_id
            ).get()

        if identity.id in SITE_ADMINS:
            self._is_admin_user = True
            return True

        self._is_admin_user = False
        return False

    def get_request_object(self):
        return WebObRequest(self.environ)

    def only_allow_admins(self):
        if not self.is_current_user_admin():
            raise self.Redirect(self.create_google_login_url(self.uri))

    def redirect(self, dest):
        raise self.Redirect(dest)

    def pretty_print(self, object):
        stream = StringIO()
        pprint(object, stream)
        self.response.write(stream.getvalue())

    def out(self, arg):
        if isinstance(arg, str):
            self.response.write(arg)
        elif isinstance(arg, unicode):
            self.response.write(arg.encode('utf-8'))
        else:
            self.response.write(str(arg))

# ------------------------------------------------------------------------------
# http request handlers
# ------------------------------------------------------------------------------

SUPPORTED_HTTP_METHODS = ('OPTIONS', 'GET', 'HEAD', 'POST', 'DELETE')

def handle_http_request(
    environ, response, html_formats=('html', 'index.html'),
    json_formats=('json', 'index.json'), rss_formats=('rss', 'index.rss')
    ):
    """Handle generic HTTP requests."""

    ctx = Context(environ, response)
    args, kwargs = ctx.request_args, ctx.request_kwargs

    STATIC.ctx = ctx

    if 'submit' in kwargs:
        del kwargs['submit']

    service_name = 'site.root_object'
    format = None

    if args and args[0].startswith('a/'):
        service_name = args[0][2:]
        args = args[1:]

    if '__format__' in ctx.special_kwargs:
        format = ctx.special_kwargs['__format__']
        if format in json_formats:
            format = 'json'
        elif format in rss_formats:
            format = 'rss'
        elif format in html_formats:
            format = 'html'

    try:
        output = render_service(ctx, service_name, format, *args, **kwargs)
    except NotFound, msg:
        response.set_response_status(404)
        # render_service('core.404')
        output = ERROR_404_TEMPLATE
        slot = 'content_slot'
    except UnAuth, msg:
        response.set_response_status(401)
        output = ERROR_401_TEMPLATE % msg
        slot = 'content_slot'
    except Redirect:
        raise
    except Exception:
        response.set_response_status(500)
        logging.error(''.join(format_exception(*sys.exc_info())))
        output = ERROR_500_TEMPLATE % ''.join(format_traceback(*sys.exc_info()))
        slot = 'content_slot'

    # call_service('core.send_events')
    ctx.out(output)

def handle_http_options_request(environ, response):
    """Handle an HTTP OPTIONS request."""

    return response.set_header(
        'Allow', ', '.join(HTTP_HANDLERS.keys())
        )

register_http_handler('GET', handle_http_request)
register_http_handler('HEAD', handle_http_request)
register_http_handler('POST', handle_http_request)
register_http_handler('OPTIONS', handle_http_options_request)

# ------------------------------------------------------------------------------
# default template builtins
# ------------------------------------------------------------------------------

BUILTINS = {
    'DEBUG': DEBUG,
    'STATIC': STATIC,
    'Markup': Markup,
    'content_slot': '',
    'urlencode': urlencode,
    'urlquote': urlquote,
    'validate': validate
    }

class Raw(unicode):
    """A raw return value from services."""

    def __new__(klass, value='', finish=False):
        self = unicode.__new__(klass, value)
        self.finish = finish
        return self

# ------------------------------------------------------------------------------
# patch beaker to support app engine memcache
# ------------------------------------------------------------------------------

import google.appengine.api.memcache as memcache

sys.modules['memcache'] = memcache

def _patch_beaker():
    """Patch Beaker to use Memcache."""

    import beaker.ext.memcached
    from beaker.synchronization import null_synchronizer

    beaker.ext.memcached.verify_directory = lambda x: None
    beaker.ext.memcached.MemcachedNamespaceManager.get_creation_lock = lambda x, y: null_synchronizer()

    import beaker.container

    class Value(beaker.container.Value):
        def get_value(self):
            self.namespace.acquire_read_lock()
            try:
                ### start hack
                has_value = False
                try:
                    value = self.__get_value()
                    has_value = True
                    if not self._is_expired():
                        return value
                except (TypeError, KeyError):
                    # guard against un-mutexed backends raising KeyError
                    pass
                ### end hack
                if not self.createfunc:
                    raise KeyError(self.key)
            finally:
                self.namespace.release_read_lock()
            has_createlock = False
            creation_lock = self.namespace.get_creation_lock(self.key)
            if has_value:
                if not creation_lock.acquire(wait=False):
                    return value
                else:
                    has_createlock = True
            if not has_createlock:
                creation_lock.acquire()
            try:
                # see if someone created the value already
                self.namespace.acquire_read_lock()
                try:
                    if self.has_value():
                        try:
                            value = self.__get_value()
                            if not self._is_expired():
                                return value
                        except KeyError:
                            # guard against un-mutexed backends raising KeyError
                            pass
                finally:
                    self.namespace.release_read_lock()
                v = self.createfunc()
                self.set_value(v)
                return v
            finally:
                creation_lock.release()

    beaker.container.Value = Value

_patch_beaker()

# ------------------------------------------------------------------------------
# genshi template handlers
# ------------------------------------------------------------------------------

GENSHI_TEMPLATE_CACHE = {}

if DEBUG:

    GENSHI_MTIME_DATA = {}

    def get_genshi_template(
        name, klass=MarkupTemplate, roots=(join_path(APP_ROOT, 'service'),)
        ):

        for root in roots:
            filepath = join_path(root, *name.split('.')) + '.genshi'
            if not exists(filepath):
                filepath = None
                continue

        if not filepath:
            raise IOError("Cannot find template gensi:%s" % name)

        template_time = getmtime(filepath)

        if ((template_time <= GENSHI_MTIME_DATA.get(name, 0)) and
            (name in GENSHI_TEMPLATE_CACHE)):
            return GENSHI_TEMPLATE_CACHE[name]

        try:
            template = klass(open(filepath, 'U'), filepath, name)
        except IOError:
            raise IOError(
                "Cannot find template genshi%s:%s" %
                (((klass == NewTextTemplate) and '-text' or ''), name)
                )

        GENSHI_TEMPLATE_CACHE[name] = template
        GENSHI_MTIME_DATA[name] = template_time

        return template

else:

    def get_genshi_template(
        name, klass=MarkupTemplate, roots=(join_path(APP_ROOT, 'service'),)
        ):

        if name in GENSHI_TEMPLATE_CACHE:
            return GENSHI_TEMPLATE_CACHE[name]

        for root in roots:
            filepath = join_path(root, *name.split('.')) + '.genshi'
            if not exists(filepath):
                filepath = None
                continue

        if not filepath:
            raise IOError("Cannot find template gensi:%s" % name)

        try:
            template = klass(open(filepath, 'U'), filepath, name)
        except IOError:
            raise IOError(
                "Cannot find template genshi%s:%s" %
                (((klass == NewTextTemplate) and '-text' or ''), name)
                )

        return GENSHI_TEMPLATE_CACHE.setdefault(name, template)

def call_genshi_template(template, template_mode='xhtml', **kwargs):
    return template.generate(**kwargs).render(template_mode)

def render_genshi_template(template_name, **kwargs):
    return get_genshi_template(template_name).generate(**kwargs).render('xhtml')

# ------------------------------------------------------------------------------
# mako template handlers
# ------------------------------------------------------------------------------

class MakoTemplateLookup(object):
    """Lookup Mako Templates."""

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

    def __init__(self, **kwargs):
        self.template_args = self.default_template_args.copy()
        self.template_args.update(kwargs)
        self._template_cache = {}
        self._template_mtime_data = {}

    TEMPLATE_ROOTS = (join_path(APP_ROOT, 'service'),)

    if DEBUG:

        # @/@ perhaps should handle case where template.filename == None

        def get_template(self, uri, kwargs=None):

            for root in self.TEMPLATE_ROOTS:
                filepath = join_path(root, *uri.split('.')) + '.mako'
                if exists(filepath):
                    break
            else:
                filepath = None

            if not filepath:
                raise IOError("Cannot find template mako:%s" % uri)

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

            for root in self.TEMPLATE_ROOTS:
                filepath = join_path(root, *uri.split('.')) + '.mako'
                if exists(filepath):
                    break
            else:
                filepath = None

            if not filepath:
                raise IOError("Cannot find template mako:%s" % uri)

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

MAKO_TEMPLATE_LOOKUP = MakoTemplateLookup()

get_mako_template = MAKO_TEMPLATE_LOOKUP.get_template

def call_mako_template(template, **kwargs):
    return template.render(**kwargs)

def render_mako_template(template_name, **kwargs):
    return get_mako_template(template_name).render(**kwargs)

# ------------------------------------------------------------------------------
# template getter
# ------------------------------------------------------------------------------

def get_template(
    name, template_cache, template_kwargs, default_mode=DEFAULT_TEMPLATE_MODE
    ):
    """Get a named template."""

    if (not DEBUG) and (name, template_kwargs) in template_cache:
        return template_cache[(name, template_kwargs)]

    template = name.split(':', 1)

    if len(template) == 1:
        template_type = DEFAULT_TEMPLATE_MODE
        template = template[0]
    else:
        template_type, template = template

    if template_type == 'mako':
        tmpl = get_mako_template(template, template_kwargs)
        template = tmpl.render

    elif template_type.startswith('genshi'):

        if template_type == 'genshi':
            tmpl_class = MarkupTemplate
            default_format = 'xhtml'
        elif template_type == 'genshi-text':
            tmpl_class = NewTextTemplate
            default_format = 'text'
        else:
            raise IOError("Cannot find template %r" % name)

        tmpl = get_genshi_template(template, tmpl_class)

        if template_kwargs:
            _template_kwargs = dict(template_kwargs)
        else:
            _template_kwargs = {'method': default_format}

        template = lambda **kwargs : tmpl.generate(**kwargs).render(
            **_template_kwargs
            )

    else:
        raise IOError("Cannot find template %r" % name)

    template_cache[(name, template_kwargs)] = template

    return template

# ------------------------------------------------------------------------------
# default views -- format renderers
# ------------------------------------------------------------------------------

def default_html_view(
    ctx, output, template=None, page_title='', slot='content_slot',
    slot_template=SITE_MAIN_TEMPLATE, slot_set=False, builtins=BUILTINS,
    get_template=get_template, template_kwargs=None, slot_template_kwargs=None,
    template_cache={}
    ):
    """Render an HTML output of the given input."""

    if 'Content-Type' not in ctx.response.headers:
        ctx.response.headers['Content-Type'] = 'text/html; charset=utf-8'

    if isinstance(output, Raw):
        if output.finish:
            return output

    else:

        if isinstance(template, basestring):
            template = get_template(template, template_cache, template_kwargs)

        kwargs = builtins.copy()

        if template:
            kwargs.update(output)
            output = template(ctx=ctx, **kwargs)

    if '__slot__' in ctx.special_kwargs:
        slot = ctx.special_kwargs['__slot__']
        if slot == '0' or not slot:
            slot = None
        else:
            slot_set = True

    if slot and slot_template and (
        (not ctx.environ.get('HTTP_X_REQUESTED_WITH')) or slot_set
        ):

        if isinstance(slot_template, basestring):
            slot_template = get_template(
                slot_template, template_cache, slot_template_kwargs
                )

        kwargs = builtins.copy()
        kwargs[slot] = output
        kwargs['page_title'] = page_title
        output = slot_template(ctx=ctx, **kwargs)

    return output

def default_cli_view(ctx, output):
    """Render a command line output of the given input."""

    return output

def default_ul_view(ctx, output):
    """Render an ueberlein output of the given input."""

    return output

def default_rss_view(ctx, output):
    """Render an RSS output of the given input."""

    if 'Content-Type' not in ctx.response.headers:
        ctx.response.headers['Content-Type'] = 'application/rss+xml; charset=utf-8'

    return output

def default_json_view(ctx, output):
    """Render a JSON output of the given input."""

    if 'Content-Type' not in ctx.response.headers:
        ctx.response.headers['Content-Type'] = 'application/json; charset=utf-8'

    if '__callback__' in ctx.special_kwargs:
        return '%s(%s)' % (
            ctx.special_kwargs['__callback__'],
            json_encode(output)
            )

    return json_encode(output)

DEFAULT_VIEWS = dict(
    cli=default_cli_view,
    html=default_html_view,
    json=default_json_view,
    rss=default_rss_view,
    ul=default_ul_view
    )

DEFAULT_HELP = u"No help is available for this service."
DEFAULT_PREVIEW = u"No preview is available for this service."

# ------------------------------------------------------------------------------
# utility klasses to define views
# ------------------------------------------------------------------------------

class View(dict):
    """A container for view formats."""

    def __init__(self, **formats):
        for name, value in formats.iteritems():
            self[name] = value

    def __getattr__(self, key):
        if key in self:
            return self.__getitem__(key)
        try:
            return self.__dict__.__getitem__(key)
        except KeyError:
            raise AttributeError(
                "%r has no attribute %r" % (self.__class__.__name__, key)
                )

    def __setattr__(self, key, value):
        if key not in dict.__dict__: # maybe not the best way?
            return self.__setitem__(key, value)
        return self.__dict__.__setitem__(key, value)

class Render(object):
    """A container for a format-specific renderer and options."""

    def __init__(self, renderer=None, *args, **kwargs):
        self.renderer = renderer
        self.args = args
        self.kwargs = kwargs

    def __call__(self, ctx, output):
        return self.renderer(ctx, output, *self.args, **self.kwargs)

# ------------------------------------------------------------------------------
# kache key funktion
# ------------------------------------------------------------------------------

def cache_key_gen(ctx, cache_spec, name, format, *args, **kwargs):
    """Generate cache key/info for the render service call."""

    player = ''

    if cache_spec.get('player', True):
        player = ctx.get_current_player_id()
        if (not cache_spec.get('anon', True)) and not player:
            return

    if cache_spec.get('ignore_args', False):
        args = ()

    if cache_spec.get('ignore_kwargs', False):
        kwargs = {}

    key = sha1(
        "%r-%r-%r-%r" % (player, format, args, sorted(kwargs.iteritems()))
        ).hexdigest()

    namespace = cache_spec.get('namespace', None)
    if namespace is None:
        namespace = name

    return key, namespace, cache_spec.get('time', 20)

# ------------------------------------------------------------------------------
# kore servise objekt
# ------------------------------------------------------------------------------

class Service(object):
    """A service object."""

    __services__ = {}
    __views__ = {}
    __help__ = {}
    __preview__ = {}

    _loaded_modules = set()
    _unknown_services = set()

    def __init__(
        self, function, name, views=None, cache=False, cache_key=cache_key_gen,
        cache_spec=dict(namespace=None, time=10, player=True, anon=True),
        default_format='html', inheritable_formats=('html',),
        help=DEFAULT_HELP, preview=DEFAULT_PREVIEW,
        admin_only=False, safe_html=False, ssl_only=False, token_required=True,
        validators=None
        ):

        if name in self.__services__:
            raise Error("Service already exists: %r" % name)
        else:
            self.__services__[name] = self
            self.__help__[name] = help
            self.__preview__[name] = preview

        if isinstance(views, basestring):
            views = {'html': Render(template=views)}
        elif isinstance(views, FunctionType):
            views = {'html': Render(renderer=views)}

        self.__views__[name] = service_views = {}

        if views:
            for format, renderer in views.iteritems():
                if isinstance(renderer, Render):
                    if renderer.renderer is None:
                        renderer.renderer = DEFAULT_VIEWS[format]
                service_views[format] = renderer

        for format in inheritable_formats:
            if (format in DEFAULT_VIEWS) and (format not in service_views):
                service_views[format] = DEFAULT_VIEWS[format]

        if validators is None:
            self.function = function
        else:
            self.function = validators(function)

        self.name = name

        self.admin_only = admin_only
        self.cache = cache
        self.cache_key = cache_key
        self.cache_spec = cache_spec
        self.default_format = default_format
        self.safe_html = safe_html
        self.ssl_only = ssl_only
        self.token_required = token_required

    def __call__(self, ctx, *args, **kwargs):
        if self.ssl_only and not ctx.ssl_mode:
            raise Error(u"This service %r is only callable over SSL." % self.name)
        if self.token_required and not ctx.valid_auth_token:
            raise Error(u"Invalid authorisation token for service %r." % self.name)
        if self.admin_only and not ctx.is_admin_user():
            raise Error("You need to be authenticated as a site admin for service %r." % self.name)
        return self.function(ctx, *args, **kwargs)

    @classmethod
    def get_service(klass, name, prefix='service.'):
        if name in klass.__services__:
            return klass.__services__[name]
        if name in klass._unknown_services:
            raise ServiceNotFound("Cannot find service: %r" % name)
        service_prefix = name.split('.', 1)[0]
        for mod in (service_prefix, name):
            if not mod:
                continue
            mod = prefix + mod
            if mod in klass._loaded_modules:
                continue
            try:
                __import__(mod)
            except ImportError, errmsg:
                # @/@ this may suppress similarly named modules in extreme cases
                if not mod.endswith(errmsg[0].split()[-1]):
                    raise
            else:
                klass._loaded_modules.add(mod)
                if name in klass.__services__:
                    return klass.__services__[name]
        klass._unknown_services.add(name)
        raise ServiceNotFound("Cannot find service: %r" % name)

    @classmethod
    def call_service(klass, ctx, name, *args, **kwargs):
        return klass.get_service(name)(ctx, *args, **kwargs)

    @classmethod
    def render_service(klass, ctx, __name, __format=None, *__args, **__kwargs):

        service = klass.get_service(__name)
        views = klass.__views__[__name]

        if __format is None:
            __format = service.default_format

        if __format not in views:
            # raise Error(_(u"Unknown format {0} for service {1}", __format, __name))
            raise Error(u"Unknown format %r for service %r" % (__format, __name))

        if service.cache:
            cache_info = service.cache_key(
                ctx, service.cache_spec, __name, __format, *__args, **__kwargs
                )
            if cache_info is not None:
                cache_key, cache_namespace, cache_time = cache_info
                output = memcache.get(cache_key, cache_namespace)
                if output is not None:
                    return output

        ctx.render_stack.append((service, __format, __args, __kwargs))
        ctx.return_render = None

        try:
            output = service(ctx, *__args, **__kwargs)
            if output is None:
                output = {}
            if ctx.return_render:
                return output
            for hook in ctx._pre_render_hooks[:]:
                output = hook(ctx, output)
                if ctx.return_render:
                    return output
            output = views[__format](ctx, output)
            if ctx.return_render:
                return output
            for hook in ctx._post_render_hooks[:]:
                output = hook(ctx, output)
                if ctx.return_render:
                    return output
            if service.cache and cache_info is not None:
                memcache.set(
                    cache_key, output, cache_time, namespace=cache_namespace
                    )
            return output
        finally:
            ctx.render_stack.pop()

    @classmethod
    def exec_service_command(klass, command):
        pass

    @classmethod
    def register_service(klass, *args, **kwargs):
        """Decorate a function with service-enabled behaviour."""

        def __register_service(function):
            return klass(function, *args, **kwargs)
        return __register_service

get_service = Service.get_service
call_service = Service.call_service
render_service = Service.render_service
exec_service_command = Service.exec_service_command
register_service = Service.register_service

# ------------------------------------------------------------------------------
# text indexing
# ------------------------------------------------------------------------------

STOP_WORDS = frozenset(['the', 'of', 'to', 'and'])

# STOP_WORDS = {
#     'en': frozenset([
#         'a', 'about', 'according', 'accordingly', 'affected', 'affecting',
#         # 'after',
#         'again', 'against', 'all', 'almost', 'already', 'also', 'although',
#         'always', 'am', 'among', 'an', 'and', 'any', 'anyone', 'apparently', 'are',
#         'arise', 'as', 'aside', 'at',
#         # 'away',
#         'be', 'became', 'because', 'become',
#         'becomes', 'been', 'before', 'being', 'between', 'both', 'briefly', 'but',
#         'by', 'came', 'can', 'cannot', 'certain', 'certainly', 'could', 'did', 'do',
#         'does', 'done', 'during', 'each', 'either', 'else', 'etc', 'ever', 'every',
#         'following', 'for', 'found', 'from', 'further', 'gave', 'gets', 'give',
#         'given', 'giving', 'gone', 'got', 'had', 'hardly', 'has', 'have', 'having',
#         'here', 'how', 'however', 'i', "i'm", 'if', 'in', 'into', 'is', 'it', 'its',
#         "it's", 'itself',
#         # 'just',
#         'keep', 'kept', 'knowledge', 'largely', 'like', 'made', 'mainly',
#         'make', 'many', 'might', 'more', 'most', 'mostly', 'much', 'must', 'nearly',
#         'necessarily', 'neither', 'next', 'no', 'none', 'nor', 'normally', 'not',
#         'noted', 'now', 'obtain', 'obtained', 'of', 'often', 'on', 'only', 'or',
#         'other', 'our', 'out', 'owing', 'particularly', 'past', 'perhaps', 'please',
#         'poorly', 'possible', 'possibly', 'potentially', 'predominantly', 'present',
#         'previously', 'primarily', 'probably', 'prompt', 'promptly', 'put',
#         'quickly', 'quite', 'rather', 'readily', 'really', 'recently', 'regarding',
#         'regardless', 'relatively', 'respectively', 'resulted', 'resulting',
#         'results', 'said', 'same', 'seem', 'seen', 'several', 'shall', 'should',
#         'show', 'showed', 'shown', 'shows', 'significantly', 'similar', 'similarly',
#         'since', 'slightly', 'so', 'some', 'sometime', 'somewhat', 'soon',
#         'specifically',
#         # 'state',
#         'states', 'strongly', 'substantially',
#         'successfully', 'such', 'sufficiently', 'than', 'that', 'the', 'their',
#         'theirs', 'them', 'then', 'there', 'therefore', 'these', 'they', 'this',
#         'those', 'though', 'through', 'throughout', 'to', 'too', 'toward', 'under',
#         'unless', 'until', 'up', 'upon', 'use', 'used', 'usefully', 'usefulness',
#         'using', 'usually', 'various', 'very', 'was', 'we', 'were', 'what', 'when',
#         'where', 'whether', 'which', 'while', 'who', 'whose', 'why', 'widely',
#         'will', 'with', 'within', 'without', 'would', 'yet', 'you'
#         ])
#     }

# find_all_words = compile_regex(
#     r'[^\s!\"#$%&()*+,-./:;<=>?@\[\\^_`{|}~]*'
#     ).findall

# find_all_words = compile_regex(r'(?u)\w+').findall # (?L)\w+

find_all_words = compile_regex(r'(?u)\w+').findall # (?L)\w+

HTML_PATTERNS = (
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
    text, strip_html=True, min_word_length=1,
    stop_words=STOP_WORDS, html_patterns=HTML_PATTERNS,
    find_words_in_text=find_all_words
    ):
    """
    Harvest words from the given ``text``.

      >>> text = "hello <tag>&nbsp;world. here! there, is a ain't 'so' \
      ...   it \"great hello '   '?"

    """ # emacs'

    if strip_html:
        for replace_html in html_patterns:
            text = replace_html(' ', text)

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

# ------------------------------------------------------------------------------
# plex link syntax
# ------------------------------------------------------------------------------

replace_links = compile_regex(r'[^\\]\[\[(.*?)[^\\]\]\]').sub

def _handle_links(content):
    pass

def handle_links(content):
    return replace_links(content, _handle_links)

# youtube
# service with multiple possible internal service calls with diff formats
# error handling

# stringify_entire_input_as_one=False,
# non_evaluate_code_blocks=('input',)
# anon=1

@register_service('hello', token_required=False)
def hello(ctx, *args, **kwargs):

    if ctx.render_format == 'html':
        ctx.post_render_hook = lambda ctx, result: "<strong>Hello</strong><br />%s" % result

    return {
        'args': args,
        'kwargs': kwargs
        }

# ------------------------------------------------------------------------------
# self runner -- app engine cached main() function
# ------------------------------------------------------------------------------

if DEBUG == 2:

    from cProfile import Profile
    from pstats import Stats

    def runner():
        run_wsgi_app(Application)

    def main():
        """Profiling main function."""

        profiler = Profile()
        profiler = profiler.runctx("runner()", globals(), locals())
        iostream = StringIO()

        stats = Stats(profiler, stream=iostream)
        stats.sort_stats("time")  # or cumulative
        stats.print_stats(80)     # 80 == how many to print

        # optional:
        # stats.print_callees()
        # stats.print_callers()

        logging.info("Profile data:\n%s", iostream.getvalue())

else:

    def main():
        """Default main function."""

        run_wsgi_app(Application)

# ------------------------------------------------------------------------------
# run in standalone mode
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
