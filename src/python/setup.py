#! /usr/bin/env python

# Public Domain (-) 2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

import ampify

from setuptools import setup

# ------------------------------------------------------------------------------
# Run Setup
# ------------------------------------------------------------------------------

setup(
    name="ampify",
    author="tav",
    author_email="tav@espians.com",
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: Public Domain",
        "Operating System :: OS Independent",
        "Programming Language :: Python",
        "Topic :: System :: Networking"
        ],
    description="Support package for writing Ampify Nodules in Python",
    keywords=["ampify"],
    license="Public Domain",
    long_description=ampify.__doc__.strip(),
    packages=["ampify"],
    url="http://ampify.it",
    version=ampify.__release__,
    zip_safe=True
    )
