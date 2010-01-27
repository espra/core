# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# ------------------------------------------------------------------------------
# the base site info
# ------------------------------------------------------------------------------

SITE_DOMAIN = 'ampifyit.appspot.com'
SITE_HTTP_URL = 'http://ampifyit.appspot.com'

# the following line needs to be left exactly as it is!!
# %(include_base_config)s

# ------------------------------------------------------------------------------
# define your config values from here on as you please
# ------------------------------------------------------------------------------

SITE_ADMINS = frozenset([
    'adminuser@gmail.com'
    ])

TAMPER_PROOF_KEY = "Place your long, randomly generated sekret key here."

TAMPER_PROOF_DEFAULT_DURATION = timedelta(minutes=20)

REMOTE_TOKEN = "Your sekret key for Remote API calls goes here."

# ------------------------------------------------------------------------------
# import non-public values to override predefined dummy values
# ------------------------------------------------------------------------------

try:
    from secret import *
except ImportError:
    pass
