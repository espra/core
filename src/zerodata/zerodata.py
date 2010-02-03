# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Zerodata Framework."""

import logging
import os
import sys

from cgi import FieldStorage
from time import time
from traceback import format_exception
from urllib import unquote as urlunquote

from google.appengine.api.capabilities import CapabilitySet
from demjson import encode as json_encode, decode as json_decode

from pyutil.crypto import validate_tamper_proof_string

from config import *

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

DEVNULL = DevNull()
HTTP_HANDLERS = {}
SSL_FLAGS = frozenset(['yes', 'on', '1'])

if os.environ.get('SERVER_SOFTWARE', '').startswith('Google'):
    RUNNING_ON_GOOGLE_SERVERS = True
else:
    RUNNING_ON_GOOGLE_SERVERS = False

OK = "Status: 200 OK\r\n"
UNAUTH = "Status: 401 Unauthorized\r\n"
ERROR = "Status: 500 Internal Server Error\r\n"

CONTENT_TYPE = "Content-Type: text/html; charset=utf-8\r\n"
LINE = '\r\n'
OK_HEADER = OK + CONTENT_TYPE + LINE
ERROR_HEADER = ERROR + CONTENT_TYPE + LINE

DISABLED = '{"error": "CapabilityError", "error_msg": "%s disabled"}'
UTF8 = 'utf-8'

VALID_HTTP_METHODS = frozenset(['GET', 'POST'])
VALID_REQUEST_CONTENT_TYPES = frozenset([
    '', 'application/x-www-form-urlencoded', 'multipart/form-data'
    ])

# ------------------------------------------------------------------------------
# app runner
# ------------------------------------------------------------------------------

def run_app():
    """The core application runner."""

    env = dict(os.environ)

    sys._boot_stdout = sys.stdout
    sys.stdout = DEVNULL
    write = sys._boot_stdout.write

    try:

        http_method = env['REQUEST_METHOD']
        path = env['PATH_INFO']
        query = env['QUERY_STRING']
        content_type = env.get('CONTENT-TYPE', '')

        # we assume that the request is utf-8 encoded, but that the request
        # arg/kwarg "keys" are in ascii and the kwarg values to be in utf-8
        args = [arg for arg in path.split('/') if arg]
        kwargs = {}

        # parse the query string
        for part in [
            sub_part
            for part in query.lstrip('?').split('&')
            for sub_part in part.split(';')
            ]:
            if not part:
                continue
            part = part.split('=', 1)
            if len(part) == 1:
                continue
            key = urlunquote(part[0].replace('+', ' '))
            value = part[1]
            if value:
                value = unicode(
                    urlunquote(value.replace('+', ' ')),
                    UTF8, 'strict'
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

        # parse the POST body if it exists and is of a known content type
        if http_method == 'POST':

            if ';' in content_type:
                content_type = content_type.split(';', 1)[0]

            if content_type in VALID_REQUEST_CONTENT_TYPES:

                post_environ = env.copy()
                post_environ['QUERY_STRING'] = ''

                post_data = FieldStorage(
                    environ=post_environ, fp=env['wsgi.input'],
                    keep_blank_values=True
                    ).list

                if post_data:
                    for field in post_data:
                        key = field.name
                        if field.filename:
                            value = field
                        else:
                            value = unicode(field.value, UTF8, 'strict')
                        if key in kwargs:
                            _val = kwargs[key]
                            if isinstance(_val, list):
                                _val.append(value)
                            else:
                                kwargs[key] = [_val, value]
                            continue
                        kwargs[key] = value

        # force the request to be over SSL only when deployed
        if RUNNING_ON_GOOGLE_SERVERS and env.get('HTTPS') not in SSL_FLAGS:
            write(UNAUTH)
            return

        # do we have handlers for the request's http method?
        if http_method not in VALID_HTTP_METHODS:
            write(ERROR)
            return

        if args:
            api_name = args.pop(0)
            api_definition = HTTP_HANDLERS.get(api_name)
            check_token = True
        else:
            api_definition = HTTP_HANDLERS.get('/')
            check_token = False

        # do we have an api handler for the requested api?
        if not api_definition:
            write(ERROR)
            return

        if check_token:
            pass

        handler, store_needed = api_definition

        # check if the datastore and memcache services are available
        if store_needed:
            disabled = None
            if not CapabilitySet('datastore_v3', capabilities=['write']).is_enabled():
                disabled = 'datastore'
            elif not CapabilitySet('memcache', methods=['set']).is_enabled():
                disabled = 'memcache'
            if disabled:
                write(ERROR_HEADER)
                write(DISABLED % disabled)
                return

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
        logging.critical(''.join(format_exception(*sys.exc_info())))

    finally:
        sys.stdout = sys._boot_stdout

# ------------------------------------------------------------------------------
# the zerodata api
# ------------------------------------------------------------------------------

def get(foo, bar):
    return foo + bar

def put(ctx):
    return ctx * 2

def delete(ctx=None):
    a = 1/0
    return {
        "ok": time()
        }

def query(ctx):
    pass

def invalidate(ctx):
    pass

# multiop

def root(*args, **kwargs):
    return {
        'foo': time()
        }

# ------------------------------------------------------------------------------
# register the handlers
# ------------------------------------------------------------------------------

HTTP_HANDLERS.update({
    '/': (root, False),
    'get': (get, False),
    'delete': (delete, True),
    'invalidate': (invalidate, False),
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
