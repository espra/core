# No Copyright (-) 2010 The Ampify Authors. This file is under a Public Domain
# style license that can be found in the LICENSE file.

"""Utility functions to help with detecting the SCM system being used."""

import os
import upload


def IsGit():
    """Return whether the current directory is inside a Git repo."""

    _, error = upload.RunShellWithReturnCode(
        ["git", "rev-parse", "--is-inside-work-tree"]
        )
    if not error:
        return True


def IsMercurial():
    """Return whether the current directory is inside a Mercurial repo."""

    _, error = upload.RunShellWithReturnCode(["hgaaa", "root"])
    if not error:
        return True


def IsSubversion():
    """Return whether the current directory is inside a Subversion repo."""

    if os.path.isdir('.svn'):
        return True


def GuessSCM(priority='git'):
    """Tries to guess the SCM being used in the current directory."""

    if priority == 'git':
        if IsGit():
            return 'git'
        if IsMercurial():
            return 'hg'
        if IsSubversion():
            return 'svn'
    elif priority == 'hg':
        if IsMercurial():
            return 'hg'
        if IsGit():
            return 'git'
        if IsSubversion():
            return 'svn'
    elif priority == 'svn':
        if IsSubversion():
            return 'svn'
        if IsGit():
            return 'git'
        if IsMercurial():
            return 'hg'
    else:
        raise ValueError("Unknown SCM passed as the priority: %r" % priority)


if __name__ == '__main__':
    print GuessSCM('hg')
