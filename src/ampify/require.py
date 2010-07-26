# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import sys

# Exit if the Python version is not a 2.x more recent than 2.5. For this to be
# useful, the code in main.py shouldn't trigger a syntax error under Python 3.
if (not hasattr(sys, 'version_info')) or sys.version_info <= (2, 5):
    print("ERROR: Sorry, you do not have a compatible Python version.")
    print("ERROR: It needs to be a 2.x version more recent than 2.5.")
    sys.exit(1)

if sys.version_info >= (3, 0):
    print("ERROR: Sorry, you do not have a compatible Python version.")
    print("ERROR: Python 3 is not supported. You need a recent 2.x version.")
    sys.exit(1)

