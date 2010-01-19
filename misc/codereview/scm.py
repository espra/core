# Copyright (c) 2006-2009 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""SCM-specific utility classes."""

import glob
import os
import re
import shutil
import subprocess
import sys
import tempfile
import xml.dom.minidom

import gclient_utils

def ValidateEmail(email):
 return (re.match(r"^[a-zA-Z0-9._%-+]+@[a-zA-Z0-9._%-]+.[a-zA-Z]{2,6}$", email)
         is not None)


def GetCasedPath(path):
  """Elcheapos way to get the real path case on Windows."""
  if sys.platform.startswith('win') and os.path.exists(path):
    # Reconstruct the path.
    path = os.path.abspath(path)
    paths = path.split('\\')
    for i in range(len(paths)):
      if i == 0:
        # Skip drive letter.
        continue
      subpath = '\\'.join(paths[:i+1])
      prev = len('\\'.join(paths[:i]))
      # glob.glob will return the cased path for the last item only. This is why
      # we are calling it in a loop. Extract the data we want and put it back
      # into the list.
      paths[i] = glob.glob(subpath + '*')[0][prev+1:len(subpath)]
    path = '\\'.join(paths)
  return path


class GIT(object):
  COMMAND = "git"

  @staticmethod
  def Capture(args, in_directory=None, print_error=True, error_ok=False):
    """Runs git, capturing output sent to stdout as a string.

    Args:
      args: A sequence of command line parameters to be passed to git.
      in_directory: The directory where git is to be run.

    Returns:
      The output sent to stdout as a string.
    """
    c = [GIT.COMMAND]
    c.extend(args)
    try:
      return gclient_utils.CheckCall(c, in_directory, print_error)
    except gclient_utils.CheckCallError:
      if error_ok:
        return ''
      raise

  @staticmethod
  def CaptureStatus(files, upstream_branch='origin'):
    """Returns git status.

    @files can be a string (one file) or a list of files.

    Returns an array of (status, file) tuples."""
    command = ["diff", "--name-status", "-r", "%s.." % upstream_branch]
    if not files:
      pass
    elif isinstance(files, basestring):
      command.append(files)
    else:
      command.extend(files)

    status = GIT.Capture(command).rstrip()
    results = []
    if status:
      for statusline in status.split('\n'):
        m = re.match('^(\w)\t(.+)$', statusline)
        if not m:
          raise Exception("status currently unsupported: %s" % statusline)
        results.append(('%s      ' % m.group(1), m.group(2)))
    return results

  @staticmethod
  def RunAndFilterOutput(args,
                         in_directory,
                         print_messages,
                         print_stdout,
                         filter):
    """Runs a command, optionally outputting to stdout.

    stdout is passed line-by-line to the given filter function. If
    print_stdout is true, it is also printed to sys.stdout as in Run.

    Args:
      args: A sequence of command line parameters to be passed.
      in_directory: The directory where git is to be run.
      print_messages: Whether to print status messages to stdout about
        which commands are being run.
      print_stdout: Whether to forward program's output to stdout.
      filter: A function taking one argument (a string) which will be
        passed each line (with the ending newline character removed) of
        program's output for filtering.

    Raises:
      gclient_utils.Error: An error occurred while running the command.
    """
    command = [GIT.COMMAND]
    command.extend(args)
    gclient_utils.SubprocessCallAndFilter(command,
                                          in_directory,
                                          print_messages,
                                          print_stdout,
                                          filter=filter)

  @staticmethod
  def GetEmail(repo_root):
    """Retrieves the user email address if known."""
    return GIT.Capture(['config', 'user.email'],
                       repo_root, error_ok=True).strip()

  @staticmethod
  def ShortBranchName(branch):
    """Converts a name like 'refs/heads/foo' to just 'foo'."""
    return branch.replace('refs/heads/', '')

  @staticmethod
  def GetBranchRef(cwd):
    """Returns the full branch reference, e.g. 'refs/heads/master'."""
    return GIT.Capture(['symbolic-ref', 'HEAD'], cwd).strip()

  @staticmethod
  def GetBranch(cwd):
    """Returns the short branch name, e.g. 'master'."""
    return GIT.ShortBranchName(GIT.GetBranchRef(cwd))

  @staticmethod
  def FetchUpstreamTuple(cwd):
    """Returns a tuple containg remote and remote ref,
       e.g. 'origin', 'refs/heads/master'
    """
    remote = '.'
    branch = GIT.GetBranch(cwd)
    upstream_branch = None
    upstream_branch = GIT.Capture(
        ['config', 'branch.%s.merge' % branch], error_ok=True).strip()
    if upstream_branch:
      remote = GIT.Capture(
          ['config', 'branch.%s.remote' % branch],
          error_ok=True).strip()
    else:
      # Fall back on origin/master if it exits.
      GIT.Capture(['branch', '-r']).split().count('origin/master')
      remote = 'origin'
      upstream_branch = 'refs/heads/master'
    return remote, upstream_branch

  @staticmethod
  def GetUpstream(cwd):
    """Gets the current branch's upstream branch."""
    remote, upstream_branch = GIT.FetchUpstreamTuple(cwd)
    if remote is not '.':
      upstream_branch = upstream_branch.replace('heads', 'remotes/' + remote)
    return upstream_branch

  @staticmethod
  def GenerateDiff(cwd, branch=None, branch_head='HEAD', full_move=False,
                   files=None):
    """Diffs against the upstream branch or optionally another branch.

    full_move means that move or copy operations should completely recreate the
    files, usually in the prospect to apply the patch for a try job."""
    if not branch:
      branch = GIT.GetUpstream(cwd)
    command = ['diff-tree', '-p', '--no-prefix', branch, branch_head]
    if not full_move:
      command.append('-C')
    # TODO(maruel): --binary support.
    if files:
      command.append('--')
      command.extend(files)
    diff = GIT.Capture(command, cwd).splitlines(True)
    for i in range(len(diff)):
      # In the case of added files, replace /dev/null with the path to the
      # file being added.
      if diff[i].startswith('--- /dev/null'):
        diff[i] = '--- %s' % diff[i+1][4:]
    return ''.join(diff)

  @staticmethod
  def GetDifferentFiles(cwd, branch=None, branch_head='HEAD'):
    """Returns the list of modified files between two branches."""
    if not branch:
      branch = GIT.GetUpstream(cwd)
    command = ['diff', '--name-only', branch, branch_head]
    return GIT.Capture(command, cwd).splitlines(False)

  @staticmethod
  def GetPatchName(cwd):
    """Constructs a name for this patch."""
    short_sha = GIT.Capture(['rev-parse', '--short=4', 'HEAD'], cwd).strip()
    return "%s-%s" % (GIT.GetBranch(cwd), short_sha)

  @staticmethod
  def GetCheckoutRoot(path):
    """Returns the top level directory of a git checkout as an absolute path.
    """
    root = GIT.Capture(['rev-parse', '--show-cdup'], path).strip()
    return os.path.abspath(os.path.join(path, root))
