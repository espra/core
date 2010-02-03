# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""App configuration for zerodata."""

__all__ = [
    'DEBUG'
    ]

# ------------------------------------------------------------------------------
# the settings
# ------------------------------------------------------------------------------

DEBUG = True

# ------------------------------------------------------------------------------
# override sample config values with the real "secret" ones
# ------------------------------------------------------------------------------

try:
    from secret import *
except:
    pass
