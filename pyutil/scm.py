# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Utility functions to help with detecting the SCM system being used."""

from os.path import abspath, isdir
from pyutil.env import run_command, CommandNotFound


def is_git():
    """Return whether the current directory is inside a Git repo."""

    try:
        _, error = run_command(
            ["git", "rev-parse", "--is-inside-work-tree"], retcode=True
            )
    except CommandNotFound:
        return

    if not error:
        return True


def is_mercurial():
    """Return whether the current directory is inside a Mercurial repo."""

    try:
        _, error = run_command(["hg", "root"], retcode=True)
    except CommandNotFound:
        return

    if not error:
        return True


def is_subversion():
    """Return whether the current directory is inside a Subversion repo."""

    if isdir('.svn'):
        return True


def guess(priority='git'):
    """Tries to guess the SCM being used in the current directory."""

    if priority == 'git':
        if is_git():
            return 'git'
        if is_mercurial():
            return 'hg'
        if is_subversion():
            return 'svn'
    elif priority == 'hg':
        if is_mercurial():
            return 'hg'
        if is_git():
            return 'git'
        if is_subversion():
            return 'svn'
    elif priority == 'svn':
        if is_subversion():
            return 'svn'
        if is_git():
            return 'git'
        if is_mercurial():
            return 'hg'
    else:
        raise ValueError("Unknown SCM passed as the priority: %r" % priority)


class SCMBase(object):
    """SCM handler base class."""

    def __init__(self, preferred_scm='git'):
        self._preferred_scm = preferred_scm
        self._scm = None
        self._root = None

    @property
    def scm(self):
        if not self._scm:
            scm = guess(self._preferred_scm)
            if scm not in ['git']:
                raise NotImplementedError(
                    "Sorry, support not yet implemented for: %r" % scm
                    )
            self._scm = scm
        return self._scm

    @property
    def root(self):
        if not self._root:
            if self.scm == 'git':
                self._root = abspath(
                    run_command(['git', 'rev-parse', '--show-cdup']).strip()
                    )
        return self._root


class SCMConfig(SCMBase):
    """SCM Configuration handler."""

    def __init__(self, preferred_scm='git'):
        super(SCMConfig, self).__init__(preferred_scm)
        self._config_cache = {}

    def get(self, prop, default=None):
        if prop in self._config_cache:
            return self._config_cache[prop]
        if self.scm == 'git':
            value, error = run_command(['git', 'config', prop], retcode=True)
            if error:
                value = default
            else:
                value = value.strip()
        return self._config_cache.setdefault(prop, value)

    def set(self, prop, value):
        if self.scm == 'git':
            _, error = run_command(
                ['git', 'config', prop, value], retcode=True
                )
            if error:
                raise IOError("Couldn't set: git config %s %s" % (prop, value))
        self._config_cache[prop] = value

    def delete(self, prop, value_regex=None, section=None, all=None):
        if self.scm == 'git':
            if all:
                args = ['--unset-all', prop]
                if value_regex:
                    args.append(value_regex)
            elif section:
                args = ['--remove-section', prop]
            else:
                args = ['--unset', prop]
                if value_regex:
                    args.append(value_regex)
            run_command(['git', 'config'] + args)
            self._config_cache.clear()


if __name__ == '__main__':
    print guess('hg')
