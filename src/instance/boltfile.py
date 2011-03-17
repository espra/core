# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

from bolt.api import *

# ------------------------------------------------------------------------------
# Tasks
# ------------------------------------------------------------------------------

@task('frontend')
def shell():
    """start a shell for the given context [default: frontend]"""
    with settings(warn_only=True):
        env().shell()
