# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""
==============================
Runner Scripts for Ampify Zero
==============================

"""

import sys

from optparse import OptionParser
from os import chdir, makedirs, symlink
from os.path import dirname, exists, join, realpath, split

from pyutil.env import run_command

# ------------------------------------------------------------------------------
# Utility Functions
# ------------------------------------------------------------------------------

def mkdir(path):
    if not exists(path):
        makedirs(path)
        return 1

def do(*cmd, **kwargs):
    if 'redirect_stdout' not in kwargs:
        kwargs['redirect_stdout'] = False
    if 'redirect_stderr' not in kwargs:
        kwargs['redirect_stderr'] = False
    if 'exit_on_error' not in kwargs:
        kwargs['exit_on_error'] = True
    return run_command(cmd, **kwargs)

def query(question, options='Y/n', default='Y', alter=1):
    if alter:
        if options:
            question = "%s? [%s] " % (question, options)
        else:
            question = "%s? " % question
    print
    response = raw_input(question)
    if not response:
        return default
    return response

def relative_path(source, destination):
    pass

# ------------------------------------------------------------------------------
# ampinit
# ------------------------------------------------------------------------------

def ampinit(argv=None):

    argv = argv or sys.argv[1:]

    op = OptionParser(usage="Usage: %prog <instance-name> [options]")

    op.add_option('--clobber', dest='clobber', action='store_true',
                  help="clobber any existing files/directories if they exist")

    options, args = op.parse_args(argv)

    if not args:
        op.print_help()
        sys.exit(1)

    clobber = options.clobber
    instance_name = split(realpath(args[0]))[-1]

    ampify_root = dirname(dirname(realpath(__file__)))
    zero_root = join(ampify_root, 'src', 'zero')
    instance_root = join(dirname(ampify_root), instance_name)

    if exists(instance_root):
        if not clobber:
            print (
                "ERROR: A directory already exists at %s\n"
                "ERROR: Use the --clobber parameter to overwrite the directory"
                % instance_root
                )
            sys.exit(1)
        chdir(instance_root)
        diff = do('git', 'diff', '--cached', '--name-only', redirect_stdout=1)
        if diff.strip():
            print (
                "ERROR: You have a dirty working tree at %s\n"
                "ERROR: Please either commit your changes or move your files.\n"
                % instance_root
                )
            print "  These are the problematic files:"
            print
            for filename in diff.strip().splitlines():
                print "    %s" % filename
            print
        first_run = 0
    else:
        create = query("Create a instance at %s" % instance_root).lower()
        if not create.startswith('y'):
            sys.exit(2)
        makedirs(instance_root)
        print
        print "Created", instance_root
        print
        chdir(instance_root)
        do('git', 'init')
        print
        readme = open('README.md', 'wb')
        readme.close()
        do('git', 'add', 'README.md')
        do('git', 'commit', '-m', "Initialised the instance [ampinit].")
        first_run = 1

    diff = do('git', 'diff', '--cached', '--name-only', redirect_stdout=1)
    if diff.strip():
        do('git', 'commit', '-m', "Updated instance [ampinit].")

# ------------------------------------------------------------------------------
# ampzero
# ------------------------------------------------------------------------------

def ampzero(argv=None):

    argv = argv or sys.argv[1:]
    op = OptionParser(usage="Usage: %prog [options] [path/to/app/directory]")

    op.add_option('-d', '--debug', dest='debug', action='store_true',
                  help="enable debug mode")

    options, args = op.parse_args(argv)

    DEBUG = False

    if options.debug:
        DEBUG = True

    print DEBUG

def amprun(argv=None):
    
    argv = argv or sys.argv[1:]
    op = OptionParser(
        usage="Usage: %prog <instance-name> [options] [stop|quit|restart]"
        )

    op.add_option('-d', '--debug', dest='debug', action='store_true',
                  help="enable debug mode")

    options, args = op.parse_args(argv)

def ampdeploy(argv=None):

    argv = argv or sys.argv[1:]
    op = OptionParser(usage="Usage: %prog <instance-name> [options]")

    options, args = op.parse_args(argv)

def amphub(argv=None):

    argv = argv or sys.argv[1:]
    op = OptionParser(usage="Usage: %prog [register|update] [options]")

    options, args = op.parse_args(argv)

    if not args:
        op.print_help()
        sys.exit(1)
