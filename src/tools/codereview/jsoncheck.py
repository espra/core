#! /usr/bin/env python

# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Simple validator for JSON files."""


import sys

from pyutil.env import read_file
from demjson import decode, JSONDecodeError


def validate(content, source_id="<source>"):
    """Return whether the content is valid JSON."""

    try:
        decode(content, strict=True)
    except JSONDecodeError, error:
        print "\nInvalid JSON source: %s" % source_id
        print "\n\t%s\n" % error.pretty_description()
        return False

    return True


if __name__ == '__main__':

    if len(sys.argv) != 2:
        print "Usage: jsoncheck.py path/to/file.json"
        sys.exit(2)

    filename = sys.argv[1]
    content = read_file(filename)

    sys.exit(int(not validate(content, filename)))
