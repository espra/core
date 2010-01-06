# Released into the Public Domain by tav <tav@espians.com>

"""Site configuration."""

import os

from datetime import timedelta
from hashlib import sha1
from os.path import join as join_path, dirname
from posixpath import split as split_path
from time import time

try:
    from updated import APPLICATION_TIMESTAMP
except ImportError:
    APPLICATION_TIMESTAMP = time()

from tentapp import APP_ROOT

__all__ = [
    'DEBUG', 'DEFAULT_TEMPLATE_MODE', 'ERROR_401_TEMPLATE',
    'ERROR_404_TEMPLATE', 'ERROR_500_TEMPLATE', 'SITE_ADMINS',
    'SITE_MAIN_TEMPLATE', 'STATIC'
    ]

# ------------------------------------------------------------------------------
# general
# ------------------------------------------------------------------------------

DEFAULT_TEMPLATE_MODE = 'mako'
LIVE_DEPLOYMENT = False
SITE_HTTP_URL = None
STATIC_PATH = '/.static/'
STATIC_HOSTS = None

execfile(join_path(APP_ROOT, 'siteinfo.py'))

# ------------------------------------------------------------------------------
# you shouldn't have to change these ...
# ------------------------------------------------------------------------------

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
# load the real site konfig
# ------------------------------------------------------------------------------

execfile(join_path(APP_ROOT, 'siteconfig.py'))
