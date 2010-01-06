# Released into the Public Domain by tav <tav@espians.com>

import sys

from os.path import dirname, join as join_path

# ------------------------------------------------------------------------------
# extend sys.path to include the ``third_party`` lib direktory
# ------------------------------------------------------------------------------

PKG_ROOT = dirname(__file__)
APP_ROOT = dirname(PKG_ROOT)

THIRD_PARTY_LIBRARY_PATH = join_path(APP_ROOT, 'third_party')

def extend_sys_path():
    """Insert the ``third_party`` libraries directory into ``sys.path``."""

    if THIRD_PARTY_LIBRARY_PATH not in sys.path:
        sys.path.insert(0, THIRD_PARTY_LIBRARY_PATH)

extend_sys_path() # @/@ for some reason calling this once is not enough ?!?
