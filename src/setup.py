#! /usr/bin/env python

# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import ampify
import sys

from distutils.core import Extension, setup

# ------------------------------------------------------------------------------
# the extensions
# ------------------------------------------------------------------------------

extensions = [
    Extension(
        "ampify.lzf",
        ["ampify/lzf.c", "ampify/lzf/lzf_c.c", "ampify/lzf/lzf_d.c"],
        include_dirs=["ampify/lzf"],
        )
    ]

if sys.platform == 'darwin':
    extensions.append(
        Extension("ampify.darwinsandbox", ["ampify/darwinsandbox.c"])
        )

# ------------------------------------------------------------------------------
# run setup
# ------------------------------------------------------------------------------

if not sys.argv[1:]:
    sys.argv.extend(['build_ext', '-i'])

setup(
    name="ampify",
    version=ampify.__release__,
    description="Ampify: A decentralised social platform",
    ext_modules=extensions,
    )
