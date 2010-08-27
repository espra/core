# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""App configuration for the Tent App."""

import os

from datetime import timedelta
from time import time

try:
    from updated import APPLICATION_TIMESTAMP
except:
    APPLICATION_TIMESTAMP = time()

__all__ = [
    'APPLICATION_TIMESTAMP', 'COOKIE_DOMAIN_HTTP', 'COOKIE_DOMAIN_HTTPS',
    'DEBUG', 'LIVE_HOST', 'REMOTE_KEY', 'SITE_ADMINS', 'STATIC_HTTP_HOSTS',
    'STATIC_HTTPS_HOSTS', 'STATIC_PATH',
    'TAMPER_PROOF_DEFAULT_DURATION', 'TAMPER_PROOF_KEY', 'TENT_HTTP_HOST',
    'TENT_HTTPS_HOST'
    ]

# ------------------------------------------------------------------------------
# Core Settings
# ------------------------------------------------------------------------------

COOKIE_DOMAIN_HTTP = '.espra.com'
COOKIE_DOMAIN_HTTPS = 'espra.appspot.com'

DEBUG = False
#DEBUG = 1

LIVE_HOST = 'https://tentlive.espra.com'

STATIC_HTTP_HOSTS = [
    'http://static1.espra.com',
    'http://static2.espra.com',
    'http://static3.espra.com'
    ]

STATIC_HTTPS_HOSTS = [
    'https://static1.espra.appspot.com',
    'https://static2.espra.appspot.com',
    'https://static3.espra.appspot.com'
    ]

STATIC_PATH = '/static/'

TENT_HTTP_HOST = 'http://tent.espra.com'
TENT_HTTPS_HOST = 'https://espra.appspot.com'

# ------------------------------------------------------------------------------
# Secret Settings
# ------------------------------------------------------------------------------

SITE_ADMINS = frozenset([
    'admin@googlemail.com'
    ])

TAMPER_PROOF_KEY = "key"

TAMPER_PROOF_DEFAULT_DURATION = timedelta(minutes=20)

#CLEANUP_BATCH_SIZE = 100
#EXPIRATION_WINDOW = timedelta(seconds=60*60*1) # 1 hour

REMOTE_KEY = "secret"

from secret import *
