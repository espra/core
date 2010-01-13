# Released into the Public Domain by tav <tav@espians.com>

"""Exception handling support."""

from format_traceback import HTMLExceptionFormatter

# ------------------------------------------------------------------------------
# exseptions
# ------------------------------------------------------------------------------

class Error(Exception):
    """Base Weblite Exception."""

class Redirect(Error):
    """
    Redirection Error.

    This is used to handle both internal and HTTP redirects.
    """

    def __init__(self, uri, method=None, permanent=False):
        self.uri = uri
        self.method = method
        self.permanent = permanent

class UnAuth(Error):
    """Unauthorised."""

class NotFound(Error):
    """404."""

class ServiceNotFound(NotFound):
    """Service 404."""

def format_traceback(type, value, traceback, limit=200):
    """Pretty print a traceback in HTML format."""

    return HTMLExceptionFormatter(limit).formatException(type, value, traceback)
