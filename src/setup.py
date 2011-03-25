#! /usr/bin/env python

# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

import sys

from distutils.core import Extension, setup
from Cython.Distutils import build_ext

# ------------------------------------------------------------------------------
# Extensions
# ------------------------------------------------------------------------------

extensions = [
    Extension(
        "ampify._lzf",
        ["ampify/_lzf.pyx", "ampify/liblzf/lzf_c.c", "ampify/liblzf/lzf_d.c"],
        include_dirs=["ampify/liblzf"],
        )
    ]

if sys.platform == 'darwin':
    extensions.append(
        Extension("ampify.darwinsandbox", ["ampify/darwinsandbox.pyx"])
        )

# ------------------------------------------------------------------------------
# Run Setup
# ------------------------------------------------------------------------------

if not sys.argv[1:]:
    sys.argv.extend(['build_ext', '-i'])

setup(
    name="ampify",
    version="git",
    description="Ampify: A decentralised social platform",
    cmdclass=dict(build_ext=build_ext),
    ext_modules=extensions,
    )
