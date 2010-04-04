#! /usr/bin/env python

# Changes to this file by The Ampify Authors are according to the
# Public Domain license that can be found in the root LICENSE file.

# This file was adapted from depot_tools/watchlists.py in the Chromium
# repository and has the following License:

# Copyright (c) 2009 The Chromium Authors. All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are
# met:
#
#    * Redistributions of source code must retain the above copyright
# notice, this list of conditions and the following disclaimer.
#    * Redistributions in binary form must reproduce the above
# copyright notice, this list of conditions and the following disclaimer
# in the documentation and/or other materials provided with the
# distribution.
#    * Neither the name of Google Inc. nor the names of its
# contributors may be used to endorse or promote products derived from
# this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
# "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
# LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
# A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
# OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
# SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
# LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
# DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
# (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

"""
Watchlist support.

Watchlist is a mechanism that allows a developer (a "watcher") to watch over
portions of code that they are interested in. A watcher will be cc-ed to
changes that modify that portion of code, thereby giving them an opportunity
to review it before the change is committed.

Refer: http://dev.chromium.org/developers/contributing-code/watchlists

When invoked directly from the base of a repository, this script lists out
the watchers for files given on the command line. This is useful to verify
changes to a watchlist file.

"""

import logging
import os
import re
import sys


class Watchlists(object):
  """Manage Watchlists.

  This class provides mechanism to load watchlists for a repo and identify
  watchers.
  Usage:
    wl = Watchlists("/path/to/repo/root")
    watchers = wl.GetWatchersForPaths(["/path/to/file1",
                                       "/path/to/file2",])
  """

  _filename = ".watchlist"
  _repo_root = None
  _defns = {}       # Definitions
  _watchlists = {}  # name to email mapping

  def __init__(self, repo_root, filename=None):
    self._repo_root = repo_root
    if filename:
      self._filename = filename
    self._LoadWatchlistRules()

  def _GetRulesFilePath(self):
    """Returns path to WATCHLISTS file."""
    return os.path.join(self._repo_root, self._filename)

  def _HasWatchlistsFile(self):
    """Determine if watchlists are available for this repo."""
    return os.path.exists(self._GetRulesFilePath())

  def _ContentsOfWatchlistsFile(self):
    """Read the WATCHLISTS file and return its contents."""
    try:
      watchlists_file = open(self._GetRulesFilePath())
      contents = watchlists_file.read()
      watchlists_file.close()
      return contents
    except IOError, e:
      logging.error("Cannot read %s: %s" % (self._GetRulesFilePath(), e))
      return ''

  def _LoadWatchlistRules(self):
    """Load watchlists from WATCHLISTS file. Does nothing if not present."""
    if not self._HasWatchlistsFile():
      return

    contents = self._ContentsOfWatchlistsFile()
    watchlists_data = None
    try:
      watchlists_data = eval(contents, {'__builtins__': None}, None)
    except SyntaxError, e:
      logging.error("Cannot parse %s. %s" % (self._GetRulesFilePath(), e))
      return

    defns = watchlists_data.get("WATCHLIST_DEFINITIONS")
    if not defns:
      logging.error("WATCHLIST_DEFINITIONS not defined in %s" %
                    self._GetRulesFilePath())
      return
    watchlists = watchlists_data.get("WATCHLISTS")
    if not watchlists:
      logging.error("WATCHLISTS not defined in %s" % self._GetRulesFilePath())
      return
    self._defns = defns
    self._watchlists = watchlists

    # Verify that all watchlist names are defined
    for name in watchlists:
      if name not in defns:
        logging.error("%s not defined in %s" % (name, self._GetRulesFilePath()))

  def GetWatchersForPaths(self, paths):
    """Fetch the list of watchers for |paths|

    Args:
      paths: [path1, path2, ...]

    Returns:
      [u1@chromium.org, u2@gmail.com, ...]
    """
    watchers = set()  # A set, to avoid duplicates
    for path in paths:
      path = path.replace(os.sep, '/')
      for name, rule in self._defns.iteritems():
        if name not in self._watchlists: continue
        rex_str = rule.get('filepath')
        if not rex_str: continue
        if re.search(rex_str, path):
          map(watchers.add, self._watchlists[name])
    return list(watchers)


def main(argv):
  # Confirm that watchlists can be parsed and spew out the watchers
  if len(argv) < 2:
    print "Usage (from the base of repo):"
    print "  %s [file-1] [file-2] ...." % argv[0]
    return 1
  wl = Watchlists(os.getcwd())
  watchers = wl.GetWatchersForPaths(argv[1:])
  print watchers


if __name__ == '__main__':
  main(sys.argv)
