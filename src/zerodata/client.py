
# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Ampify ZeroDataStore test client."""

from urllib2 import HTTPSHandler, build_opener, install_opener, urlopen
from os.path import dirname, join as join_path, realpath

# ------------------------------------------------------------------------------
# extend sys.path
# ------------------------------------------------------------------------------

ZERODATA_ROOT = dirname(realpath(__file__))
AMPIFY_ROOT = dirname(dirname(ZERODATA_ROOT))

sys.path.insert(0, join_path(AMPIFY_ROOT, 'environ', 'startup'))

import rconsole


class ZeroDataClient(object)
    """Provides a client for ZeroDataStore"""

    def __init__(self):
        # Create an OpenerDirector with support for SSL...
        opener = build_opener(HTTPSHandler) 
        # ...and install it globally so it can be used with urlopen. 
        install_opener(opener) 
        urlopen(’http://www.example.com/login.html’)


