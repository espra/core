# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import sys

from os.path import dirname, join as join_path, realpath

AMPIFY_ROOT = dirname(dirname(dirname(realpath(__file__))))
SRC_PATH = join_path(AMPIFY_ROOT, 'src')
THIRD_PARTY_LIBS_PATH = join_path(AMPIFY_ROOT, 'third_party', 'pylibs')

if SRC_PATH not in sys.path:
    sys.path.insert(0, SRC_PATH)

if THIRD_PARTY_LIBS_PATH not in sys.path:
    sys.path.insert(0, THIRD_PARTY_LIBS_PATH)
