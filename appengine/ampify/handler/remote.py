# Released into the Public Domain by tav <tav@espians.com>

"""A handler that exports various App Engine services over HTTP."""

import logging

from os import environ
from urllib import unquote
from wsgiref.handlers import CGIHandler

from google.appengine.ext.webapp import WSGIApplication
from google.appengine.ext.remote_api.handler import (
    RemoteDatastoreStub, SERVICE_PB_MAP, ApiCallHandler
    )

from ampify.core.config import REMOTE_TOKEN, RUNNING_ON_GOOGLE_SERVERS
from ampify.util.crypto import validate_tamper_proof_string

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

SSL_ENABLED_FLAGS = frozenset(['yes', 'on', '1'])

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

        if validate_tamper_proof_string(
            'remote', verifier, timestamped=False, key=REMOTE_TOKEN,
            ) is None:
            logging.error("Unauthorised Remote Access Attempt: %r" % verifier)
            self.response.set_status(500)
            return

        # we skip the SSL check for local dev instances
        if (RUNNING_ON_GOOGLE_SERVERS and
            environ.get('HTTPS') not in SSL_ENABLED_FLAGS
            ):
            logging.error("Insecure Remote Access Attempt")
            self.response.set_status(500)
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
