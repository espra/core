# No Copyright (-) 2009-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""A handler that exports various App Engine services over HTTP."""

import logging
import os

from urllib import unquote
from wsgiref.handlers import CGIHandler

from google.appengine.ext.webapp import WSGIApplication
from google.appengine.ext.remote_api.handler import (
    RemoteDatastoreStub, SERVICE_PB_MAP, ApiCallHandler
    )

from config import REMOTE_KEY
from pyutil.crypto import validate_tamper_proof_string

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

SSL_ENABLED_FLAGS = frozenset(['yes', 'on', '1'])

if os.environ.get('SERVER_SOFTWARE', '').startswith('Google'):
    RUNNING_ON_GOOGLE_SERVERS = True
else:
    RUNNING_ON_GOOGLE_SERVERS = False

# ------------------------------------------------------------------------------
# request handler
# ------------------------------------------------------------------------------

class RemoteAPIHandler(ApiCallHandler):
    """Remote API handler."""

    def get(self):
        """Handle GET requests with a 404."""
        self.response.set_status(404)

    def post(self):
        """Handle POST requests by executing the API call."""

        verifier = unquote(self.request.path.rsplit('/', 1)[1])

        if not validate_tamper_proof_string('remote', verifier, REMOTE_KEY):
            logging.info("Unauthorised Remote Access Attempt: %r", verifier)
            self.response.set_status(401)
            return

        # we skip the SSL check for local dev instances
        if (RUNNING_ON_GOOGLE_SERVERS and
            os.environ.get('HTTPS') not in SSL_ENABLED_FLAGS
            ):
            logging.info("Insecure Remote Access Attempt")
            self.response.set_status(401)
            return

        ApiCallHandler.post(self)

    def CheckIsAdmin(self):
        """Dummy for parent class since we don't depend on this for security."""
        return True

# ------------------------------------------------------------------------------
# main funktion
# ------------------------------------------------------------------------------

def main():
    application = WSGIApplication([('.*', RemoteAPIHandler)])
    CGIHandler().run(application)

if __name__ == '__main__':
    main()
