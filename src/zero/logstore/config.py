# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""App configuration for Ampify Zero Logstore."""

import os

__all__ = [
    'API_KEY', 'DEBUG', 'REMOTE_KEY'
    ]

# ------------------------------------------------------------------------------
# the settings
# ------------------------------------------------------------------------------

DEBUG = False

REMOTE_KEY = ''

API_KEY = 'bar'

# ------------------------------------------------------------------------------
# override sample config values with the real "secret" ones
# ------------------------------------------------------------------------------

try:
    from secret import *
except:
    pass
