# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

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

def extend_sys_path():

    THIRD_PARTY_LIBS_PATH = join_path(
        dirname(dirname(realpath(__file__))), 'third_party', 'pylibs'
        )

    if THIRD_PARTY_LIBS_PATH not in sys.path:
        sys.path.insert(0, THIRD_PARTY_LIBS_PATH)

    return THIRD_PARTY_LIBS_PATH

extend_sys_path()
