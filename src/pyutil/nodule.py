# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""
========================
Python Nodule for Ampify
========================

"""

import sys

from pyutil.io import DEVNULL

NODULE_API_VERSION = 0

# A global ``SERVICE_REGISTRY`` object holds the mappings between service names
# and the representative functions.
SERVICE_REGISTRY = {}

# A super simple decorator is defined to make it easy it "register" services in
# the ``SERVICE_REGISTRY``. This function will be passed along when evaluating
# other scripts that might have been specified.
def register(name):
    def __wrapper(func):
        SERVICE_REGISTRY[name] = func
        return func
    return __wrapper

@register('test')
def test():
    return

# ------------------------------------------------------------------------------
# Main Runner
# ------------------------------------------------------------------------------

def main():
    stdin, stdout, stderr = sys.stdin, sys.stdout, sys.stderr
    sys.stdin, sys.stdout, sys.stderr = DEVNULL, DEVNULL, DEVNULL
    while 1:
        stdin.read()
    print "Hello"
