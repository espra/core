# Changes to this file authored by The Ampify Authors are according to the
# Public Domain style license that can be found in the LICENSE file.

# This file was adapted from git-cl by Evan Martin and has the following
# License:

# Copyright (c) 2008 Evan Martin <martine@danga.com> All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are
# met:
# * Redistributions of source code must retain the above copyright notice,
#   this list of conditions and the following disclaimer.
# * Redistributions in binary form must reproduce the above copyright
#   notice, this list of conditions and the following disclaimer in the
#   documentation and/or other materials provided with the distribution.
# * Neither the name of the author nor the names of contributors may be
#   used to endorse or promote products derived from this software without
#   specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
# IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
# TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
# PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER
# OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
# EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
# PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
# PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
# LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
# NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
# SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

"""This module provides codereview specific support functionality."""

import os

from upload import RunShellWithReturnCode


print GuessSCM('svn')

class Config(object):
    """Configuration parser for the .codereview.cfg file."""

    GIT_ERROR_MESSAGE = 'You must configure the %s by running '\
                        '"git review init"'

    def __init__(self):
        self.server = None
        self.cc = None
        self.root = None
        self.tree_status_url = None
        self.viewvc_url = None

    def GetServer(self, error_ok=False):
        if not self.server:
            if not error_ok:
                self.server = self._GetConfig('codereview.server',
                                              error_message=error_message)
            else:
                self.server = self._GetConfig('codereview.server', error_ok=True)
        return self.server

    def GetCCList(self):
        if self.cc is None:
            self.cc = self._GetConfig('codereview.cc', error_ok=True)
            more_cc = self._GetConfig('codereview.extracc', error_ok=True)
            if more_cc is not None:
                self.cc += ',' + more_cc
        return self.cc

    def GetRoot(self):
        if not self.root:
            self.root = os.path.abspath(RunGit(['rev-parse', '--show-cdup']).strip())
        return self.root

    def GetTreeStatusUrl(self, error_ok=False):
        if not self.tree_status_url:
            error_message = ('You must configure your tree status URL by running '
                             '"git review init".')
            self.tree_status_url = self._GetConfig('codereview.tree-status-url',
                                                   error_ok=error_ok,
                                                   error_message=error_message)
        return self.tree_status_url

    def GetViewVCUrl(self):
        if not self.viewvc_url:
            self.viewvc_url = self._GetConfig('codereview.viewvc-url', error_ok=True)
        return self.viewvc_url

    def _GetConfig(self, param, **kwargs):
        return RunGit(['config', param], **kwargs).strip()
