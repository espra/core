# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""App configuration for the Tent App."""

import os

from datetime import timedelta

__all__ = [
    'DEBUG', 'REMOTE_KEY'
    ]

# ------------------------------------------------------------------------------
# Core Settings
# ------------------------------------------------------------------------------

DEBUG = False

SITE_DOMAIN = 'espra.com'
SITE_HTTP_URL = 'http://www.espra.com'
SITE_HTTPS_URL = 'https://espra.appspot.com'

# ------------------------------------------------------------------------------
# Secret Settings
# ------------------------------------------------------------------------------

SITE_ADMINS = frozenset([
    'admin@googlemail.com'
    ])

TAMPER_PROOF_KEY = "key"

TAMPER_PROOF_DEFAULT_DURATION = timedelta(minutes=20)

REMOTE_KEY = "secret"

try:
    from secret import *
except:
    pass
