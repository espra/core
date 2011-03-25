# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

"""Tentapp configuration."""

from time import time

try:
    from env import APPLICATION_TIMESTAMP
except ImportError:
    APPLICATION_TIMESTAMP = time()

DEBUG = True

SECURE_COOKIE_DURATION = 86400
SECURE_COOKIE_KEY = "replace this with your secure key!"

SITE_ADMINS = frozenset([
    'tav'
    ])

SITE_CSS_FILE_BASE = '/static/css/atp'

SSL_ONLY = True

STATIC_HTTP_HOSTS = [
    'http://s1.tentapp.appspot.com',
    'http://s2.tentapp.appspot.com',
    'http://s3.tentapp.appspot.com'
    ]

STATIC_HTTPS_HOSTS = [
    'https://s1.tentapp.appspot.com',
    'https://s2.tentapp.appspot.com',
    'https://s3.tentapp.appspot.com'
    ]

STATIC_PATH = '/static/'

# ------------------------------------------------------------------------------
# Override Sensitive Values
# ------------------------------------------------------------------------------

try:
    from secret import *
except ImportError:
    pass
