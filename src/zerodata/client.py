
# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Ampify ZeroDataStore test client."""

from urllib2 import HTTPHandler, HTTPSHandler, ProxyHandler, \
        build_opener, install_opener, urlopen
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
        # Create an OpenerDirector with support for SSL and other stuff...
        opener = self.build_opener(True) 
        # ...and install it globally so it can be used with urlopen. 
        install_opener(opener) 
        urlopen(’https://github.com’)

    def build_opener(debug=False):
        """Create handlers with the appropriate debug level.  
        We intentionally create new ones because the OpenerDirector 
        class in urllib2 is smart enough to replace its internal 
        versions with ours if we pass them into the 
        urllib2.build_opener method.  This is much easier than 
        trying to introspect into the OpenerDirector to find the 
        existing handlers.
        based on http://code.activestate.com/recipes/440574/#c1
        """
        http_handler = HTTPHandler(debuglevel=debug)
        https_handler = HTTPSHandler(debuglevel=debug)
        proxy_handler = ProxyHandler(debuglevel=debug)
        unknown_handler = UnknownHandler(debuglevel=debug)
        http_default_error_handler = HTTPDefaultErrorHandler(debuglevel=debug)
        http_redirect_handler = HTTPRedirectHandler(debuglevel=debug)
        http_error_processor = HTTPErrorProcessor(debuglevel=debug)

        handlers = [http_handler, https_handler, proxy_handler]
        opener = build_opener(handlers)

        # Save the cookie jar with the opener just in case it's needed
        # later on
        opener.cookie_jar = cookie_jar

        return opener

