# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""
Pyutil -- A collection of utility modules for Python.

                                _       _   
                               ( )_  _ (_ ) 
           _ _    _   _  _   _ | ,_)(_) | | 
          ( '_`\ ( ) ( )( ) ( )| |  | | | | 
          | (_) )| (_) || (_) || |_ | | | | 
          | ,__/'`\__, |`\___/'`\__)(_)(___)
          | |    ( )_| |                    
          (_)    `\___/'                    

"""

import sys

from os.path import dirname, join as join_path, realpath

# ------------------------------------------------------------------------------
# extend sys.path to include the ``third_party`` lib direktory
# ------------------------------------------------------------------------------

THIRD_PARTY_LIBRARY_PATH = join_path(
    dirname(dirname(realpath(__file__))), 'third_party', 'pylibs'
    )

if THIRD_PARTY_LIBRARY_PATH not in sys.path:
    sys.path.insert(0, THIRD_PARTY_LIBRARY_PATH)
