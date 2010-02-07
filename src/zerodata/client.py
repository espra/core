
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

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------
API_OPERATIONS = ['get', 'delete', 'invalidate', 'put', 'query']
PRODUCTION_ENDPOINT = "http://ampify.appspot.com"
TEST_ENDPOINT = "http://localhost"


class ZeroDataClient(object)
    """Provides a client for ZeroDataStore"""

    def __init__(self):
        self._endpoint = TEST_ENDPOINT
        # Create an OpenerDirector with support for SSL and other stuff...
        opener = self.build_opener(True) 
        # ...and install it globally so it can be used with urlopen. 
        install_opener(opener) 
        try:
            openurl = urlopen(self._endpoint)
        except URLError:
            raise RuntimeError("API endpoint not available")

    def build_opener(self, debug=False):
        """Create handlers with the appropriate debug level.  
        We intentionally create new ones because the OpenerDirector 
        class in urllib2 is smart enough to replace its internal 
        versions with ours if we pass them into the 
        urllib2.build_opener method.  This is much easier than 
        trying to introspect into the OpenerDirector to find the 
        existing handlers.
        Based on http://code.activestate.com/recipes/440574/#c1
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

        return opener

    def call(self, api_operation, api_request):
        try:
            API_OPERATIONS.index(api_operation)
        except ValueError:
            raise RuntimeError("Invalid API operation.")
        url = "%s/%s" % (self._endpoint, api_operation)
        req = http_request(url)
        json_request = json_encode(api_request)
        try:
            openurl = urlopen(req, json_request)
        except URLError:
            raise RuntimeError("API request failed.")
        return json_decode(openurl.read())

