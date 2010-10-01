#! /usr/bin/env python

# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import pyutil
import sys

from distutils.core import Extension, setup
from Cython.Distutils import build_ext

# ------------------------------------------------------------------------------
# the extensions
# ------------------------------------------------------------------------------

extensions = [
    Extension(
        "pyutil.lzf",
        ["pyutil/lzf.pyx", "pyutil/lzf/lzf_c.c", "pyutil/lzf/lzf_d.c"],
        include_dirs=["pyutil/lzf"],
        )
    ]

if sys.platform == 'darwin':
    extensions.append(
        Extension("pyutil.darwinsandbox", ["pyutil/darwinsandbox.pyx"])
        )

# ------------------------------------------------------------------------------
# run setup
# ------------------------------------------------------------------------------

if not sys.argv[1:]:
    sys.argv.extend(['build_ext', '-i'])

setup(
    name="pyutil",
    version="git",
    description="Pyutil: A collection of useful Python modules",
    cmdclass=dict(build_ext=build_ext),
    ext_modules=extensions,
    )
