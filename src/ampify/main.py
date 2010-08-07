# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import ampify
import ampify.require

import sys

from doctest import ELLIPSIS, testmod
from optparse import OptionParser
from os import chdir, environ, makedirs, symlink
from os.path import dirname, exists, join, realpath, split
from urllib import urlopen

from optcomplete import autocomplete, DirCompleter, ListCompleter
from optcomplete import make_autocompleter, parse_options
from pyutil.env import run_command, CommandNotFound

from ampify import settings
from ampify.build import ERROR, PROGRESS, SUCCESS, MAKE
from ampify.build import do, error, exit, log, lock, unlock
from ampify.build import decode_json, load_role, install_packages

# ------------------------------------------------------------------------------
# Constants
# ------------------------------------------------------------------------------

AMPIFY_ROOT = dirname(dirname(dirname(realpath(__file__))))
AMPIFY_ROOT_PARENT = dirname(AMPIFY_ROOT)

ERRMSG_GIT_NAME_DETECTION = (
    "ERROR: Couldn't detect the instance name from the Git URL.\n          "
    "Please provide an instance name parameter. Thanks!"
    )

# ------------------------------------------------------------------------------
# Utility Functions
# ------------------------------------------------------------------------------

def relative_path(source, destination):
    pass

# This function normalises instance names -- just in case someone accidentally
# passed in a path name.
def normalise_instance_name(instance_name):
    instance_name = split(realpath(instance_name))[-1]
    instance_root = join(AMPIFY_ROOT_PARENT, instance_name)
    return instance_name, instance_root

# ------------------------------------------------------------------------------
# Main Runner
# ------------------------------------------------------------------------------

def main(argv=None, show_help=False):

    argv = argv or sys.argv[1:]

    # Set the script name to ``amp`` so that OptionParser error messages don't
    # display a meaningless ``main.py`` to end users.
    sys.argv[0] = 'amp'

    usage = ("""Usage: amp <command> [options]
    \nCommands:
    \n%s
    version  show the version number and exit
    \nSee `amp help <command>` for more info on a specific command.""" %
    '\n'.join("    %-8s %s" % (cmd, COMMANDS[cmd].help) for cmd in sorted(COMMANDS))
    )

    autocomplete(
        OptionParser(add_help_option=False),
        ListCompleter(AUTOCOMPLETE_COMMANDS.keys()),
        subcommands=AUTOCOMPLETE_COMMANDS
        )

    if not argv:
        show_help = True
    else:
        command = argv[0]
        argv = argv[1:]
        if command in ['-h', '--help']:
            show_help = True
        elif command == 'help':
            if argv:
                command = argv[0]
                argv = ['--help']
            else:
                show_help = True
        if command in ['-v', '--version', 'version']:
            print('amp version %s' % ampify.__release__)
            sys.exit()

    if show_help:
        print(usage)
        sys.exit(1)

    if command in COMMANDS:
        return COMMANDS[command](argv)

    # We support git-command like behaviour. That is, if there's an external
    # binary named ``amp-foo`` available on the ``$PATH``, then running ``amp
    # foo`` will automatically delegate to it.
    try:
        output, retcode = run_command(
            ['amp-%s' % command] + argv, retcode=True, redirect_stdout=False,
            redirect_stderr=False
            )
    except CommandNotFound:
        exit("ERROR: Unknown command %r" % command)

    if retcode:
        sys.exit(retcode)

# ------------------------------------------------------------------------------
# Build Command
# ------------------------------------------------------------------------------

def build(argv=None, completer=None):

    op = OptionParser(usage="Usage: amp build [options]", add_help_option=False)

    op.add_option('--role', dest='role', default='default',
                  help="specify a non-default role to build")

    options, args = parse_options(op, argv, completer)

    load_role(options.role)
    install_packages()

# ------------------------------------------------------------------------------
# Check Command
# ------------------------------------------------------------------------------

def check(argv=None, completer=None):

    op = OptionParser(usage="Usage: amp check", add_help_option=False)
    options, args = parse_options(op, argv, completer)

    log("Checking the current revision id for your code.", PROGRESS)
    revision_id = do(
        'git', 'show', '--pretty=oneline', '--summary', redirect_stdout=True
        ).split()[0]

    log("Checking the latest commits on GitHub.", PROGRESS)
    commit_info = urlopen(
        'http://github.com/api/v2/json/commits/list/tav/ampify/master'
        ).read()

    latest_revision_id = decode_json(commit_info)['commits'][0]['id']

    if revision_id != latest_revision_id:
        exit("A new version is available. Please run `git pull`.")

    log("Your checkout is up-to-date.", SUCCESS)

# ------------------------------------------------------------------------------
# Deploy Command
# ------------------------------------------------------------------------------

def deploy(argv=None, completer=None):

    op = OptionParser(
        usage="Usage: amp deploy <instance-name> [options]",
        add_help_option=False
        )

    op.add_option('--test', dest='test', action='store_true', default=False,
                  help="run tests before completing the switch")

    options, args = parse_options(op, argv, completer, True)

# ------------------------------------------------------------------------------
# Hub Command
# ------------------------------------------------------------------------------

def hub(argv=None, completer=None):

    op = OptionParser(
        usage="Usage: amp hub [register|update] [options]",
        add_help_option=False
        )

    options, args = parse_options(op, argv, completer, True)

# ------------------------------------------------------------------------------
# Init Command
# ------------------------------------------------------------------------------

def init(argv=None, completer=None):

    op = OptionParser(
        usage="Usage: amp init <instance-name> [options]",
        add_help_option=False
        )

    op.add_option('--clobber', dest='clobber', action='store_true',
                  help="clobber any existing files/directories if they exist")

    op.add_option('--from', dest='git_repo', default='',
                  help="initialise by cloning the given git repository")

    options, args = parse_options(op, argv, completer)

    git_repo = options.git_repo
    if git_repo:
        if args:
            instance_name = args[0]
        else:
            instance_name = git_repo.split('/')[-1].rsplit('.', 1)[0]
            if not instance_name:
                exit(ERRMSG_GIT_NAME_DETECTION)
    else:
        if args:
            instance_name = args[0]
        else:
            op.print_help()
            sys.exit(1)

    instance_name, instance_root = normalise_instance_name(instance_name)
    clobber = options.clobber

    if exists(instance_root):
        if not clobber:
            exit(
                "ERROR: A directory already exists at %s\n          "
                "Use the --clobber parameter to overwrite the directory"
                % instance_root
                )
        chdir(instance_root)
        diff = do('git', 'diff', '--cached', '--name-only', redirect_stdout=1)
        if diff.strip():
            error(
                "ERROR: You have a dirty working tree at %s\n          "
                "Please either commit your changes or move your files.\n"
                % instance_root
                )
            error("  These are the problematic files:")
            for filename in diff.strip().splitlines():
                log("    %s" % filename, ERROR)
            print
        first_run = 0
    else:
        create = query("Create a instance at %s" % instance_root).lower()
        if not create.startswith('y'):
            sys.exit(2)
        makedirs(instance_root)
        print
        print("Created %s" % instance_root)
        print
        chdir(instance_root)
        do('git', 'init')
        print
        readme = open('README.md', 'wb')
        readme.close()
        do('git', 'add', 'README.md')
        do('git', 'commit', '-m', "Initialised the instance [amp].")
        first_run = 1

    diff = do('git', 'diff', '--cached', '--name-only', redirect_stdout=1)
    if diff.strip():
        do('git', 'commit', '-m', "Updated instance [amp].")

    print(DEBUG)

# ------------------------------------------------------------------------------
# Run Command
# ------------------------------------------------------------------------------

def run(argv=None, completer=None):
    
    op = OptionParser(
        usage="Usage: amp run <instance-name> [options] [stop|quit|restart]",
        add_help_option=False
        )

    op.add_option('-d', '--debug', dest='debug', action='store_true',
                  help="enable debug mode")

    op.add_option("--file", dest="filename",
                  help="input file to read data from")

    if completer:
        return op, DirCompleter(AMPIFY_ROOT_PARENT)

    options, args = parse_options(op, argv, completer, True)

    if options.debug:
        settings.debug = True

    instance_name, instance_root = normalise_instance_name(args[0])

# ------------------------------------------------------------------------------
# Test Command
# ------------------------------------------------------------------------------

PYTHON_TEST_MODULES = [
    'pyutil.exception',
    'pyutil.rst'
    ]

def test(argv=None, completer=None, run_all=False):

    op = OptionParser(usage="Usage: amp test [options]", add_help_option=False)

    op.add_option('-a', '--all', dest='all', action='store_true',
                  help="run the comprehensive test suite")

    op.add_option('-v', '--verbose', dest='verbose', action='store_true',
                  default=False, help="enable verbose mode")

    testers = ['python', 'go']
    if completer:
        return op, ListCompleter(testers)

    options, args = parse_options(op, argv, completer)
    if not args:
        args = testers

    args = set(args)
    if options.all:
        run_all = True

    if 'python' in args:
        py_tests(PYTHON_TEST_MODULES, verbose=options.verbose)

    if 'go' in args:
        go_tests()

def go_tests():
    go_root = join(AMPIFY_ROOT, 'src', 'amp')
    chdir(go_root)
    run_command([MAKE, 'nuke'])
    _, retval = run_command(
        [MAKE, 'install', 'test'], retcode=True, redirect_stderr=False,
        redirect_stdout=False
        )
    if retval:
        sys.exit(retval)

def py_tests(modules, verbose):
    failed = 0
    for module in modules:
        module = __import__(module, fromlist=[''])
        fail, _ = testmod(module, optionflags=ELLIPSIS, verbose=verbose)
        failed += fail
    if failed:
        sys.exit(1)

# ------------------------------------------------------------------------------
# Help Strings
# ------------------------------------------------------------------------------

# These, along with other strings, should perhaps be internationalised at a
# later date.
build.help = "download and build the ampify zero dependencies"
check.help = "check if your checkout is up-to-date"
deploy.help = "deploy an instance to remote hosts"
hub.help = "interact with amphub"
init.help = "initialise a new amp instance"
run.help = "run the components for an amp instance"
test.help = "run the ampify zero test suite"

# ------------------------------------------------------------------------------
# Command Mapping
# ------------------------------------------------------------------------------

COMMANDS = {
    'build': build,
    'check': check,
    'deploy': deploy,
    'hub': hub,
    'init': init,
    'run': run,
    'test': test
    }

# ------------------------------------------------------------------------------
# Command Autocompletion
# ------------------------------------------------------------------------------

AUTOCOMPLETE_COMMANDS = COMMANDS.copy()

AUTOCOMPLETE_COMMANDS['help'] = lambda completer: (
    OptionParser(add_help_option=False),
    ListCompleter(COMMANDS.keys())
    )

AUTOCOMPLETE_COMMANDS['version'] = lambda completer: (
    OptionParser(add_help_option=False),
    DirCompleter(AMPIFY_ROOT_PARENT)
    )

for command in AUTOCOMPLETE_COMMANDS.values():
    command.autocomplete = make_autocompleter(command)

# ------------------------------------------------------------------------------
# Script Runner
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
