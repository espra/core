#! /usr/bin/env python

# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import sys

from distutils.core import Extension, setup

# ------------------------------------------------------------------------------
# the extensions
# ------------------------------------------------------------------------------

extensions = [
    Extension(
        "pyutil.pylzf",
        ["pyutil/pylzf.c", "pyutil/lzf/lzf_c.c", "pyutil/lzf/lzf_d.c"],
        include_dirs=["pyutil/lzf"],
        ),
    Extension("pyutil.darwinsandbox", ["pyutil/darwinsandbox.c"])
    ]

# ------------------------------------------------------------------------------
# run setup
# ------------------------------------------------------------------------------

if not sys.argv[1:]:
    sys.argv.extend(['build_ext', '-i'])

setup(
    name="pyutil",
    version="git",
    description="A Python utility library",
    ext_modules=extensions,
    )
