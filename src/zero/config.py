# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

try:
    from env import *
except ImportError:
    pass

if 'REMOTE_KEY' not in globals():
    REMOTE_KEY = "this should be replaced by a secure key"
