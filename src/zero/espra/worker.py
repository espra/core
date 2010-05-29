#! /usr/bin/env python

# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Worker script for Ampify Zerodata."""

import sys

from os.path import dirname, join as join_path, realpath

# ------------------------------------------------------------------------------
# extend sys.path
# ------------------------------------------------------------------------------

ZERODATA_ROOT = dirname(realpath(__file__))
AMPIFY_ROOT = dirname(dirname(ZERODATA_ROOT))

sys.path.insert(0, join_path(AMPIFY_ROOT, 'environ'))

import rconsole

# ------------------------------------------------------------------------------
# continue with the imports
# ------------------------------------------------------------------------------

import model

# ------------------------------------------------------------------------------
# setup rconsole
# ------------------------------------------------------------------------------

rconsole.setup([ZERODATA_ROOT, 'production'], ssl=True, shell=False)

# ------------------------------------------------------------------------------
# the main function
# ------------------------------------------------------------------------------

def run(argv=None):
    """The runner function for the worker."""

    argv = argv or sys.argv[1:]

    print "# Running the zerodata worker"

# ------------------------------------------------------------------------------
# self runner
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    run()
