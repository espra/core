# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

import os
import sys

from os.path import dirname, join as join_path, realpath

AMPIFY_ROOT = dirname(dirname(realpath(__file__)))

for path in [
    join_path(AMPIFY_ROOT, 'environ'),
    join_path(AMPIFY_ROOT, 'third_party', 'pylibs'),
    join_path(AMPIFY_ROOT, 'third_party', 'yatiblog'),
    join_path(AMPIFY_ROOT, 'third_party', 'tavutil'),
    join_path(AMPIFY_ROOT, 'src', 'python'),
    ]:
    if path not in sys.path:
        sys.path.insert(0, path)

if not hasattr(sys, 'skip_pylibs_check'):
    try:
        import tavutil.optcomplete
    except ImportError:
        print
        print "You need to checkout the third_party/pylibs submodule."
        print
        print "Run the following command from inside %s" % AMPIFY_ROOT
        print
        print "    git submodule update --init third_party/pylibs"
        print
        sys.exit(1)

os.environ.setdefault('AMPIFY_ROOT', AMPIFY_ROOT)
