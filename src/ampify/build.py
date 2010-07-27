# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import os
import sys

from errno import EACCES, ENOENT
from os import getcwd, environ, execve, makedirs
from os.path import dirname, exists, expanduser, isdir, isfile, join, realpath
from shutil import rmtree
from thread import start_new_thread

try:
    from multiprocessing import cpu_count
except ImportError:
    cpu_count = lambda: 1

try:
    from json import loads as decode_json
except ImportError:
    decode_json = None

from pyutil.env import run_command, CommandNotFound

# ------------------------------------------------------------------------------
# Print Functions
# ------------------------------------------------------------------------------

if os.name == 'posix' and not environ.get('AMPIFY_PLAIN'):
    ACTION = '\x1b[34;01m>> '
    INSTRUCTION = '\x1b[31;01m!! '
    ERROR = '\x1b[31;01m!! '
    NORMAL = '\x1b[0m'
    PROGRESS = '\x1b[30;01m## '
    SUCCESS = '\x1b[32;01m** '
    TERMTITLE = '\x1b]2;%s\x07'
else:
    INSTRUCTION = ACTION = '>> '
    ERROR = '!! '
    NORMAL = ''
    PROGRESS = '## '
    SUCCESS = '** '
    TERMTITLE = ''

# Pretty print the given ``message`` in nice colours.
def log(message, type=ACTION):
    print type + message + NORMAL

def error(message):
    print ERROR + message + NORMAL
    print ''

def exit(message):
    print ERROR + message + NORMAL
    sys.exit(1)

# ------------------------------------------------------------------------------
# Platform Detection
# ------------------------------------------------------------------------------

# Only certain UNIX-like platforms are currently supported. Most of the code is
# easily portable to other UNIX platforms. Unfortunately porting to Windows will
# be problematic as POSIX APIs are used a fair bit -- especially by dependencies
# like redis.
if sys.platform.startswith('darwin'):
    PLATFORM = 'darwin'
elif sys.platform.startswith('linux'):
    PLATFORM = 'linux'
elif sys.platform.startswith('freebsd'):
    PLATFORM = 'freebsd'
else:
    exit(
        "ERROR: Sorry, the %r operating system is not supported yet."
        % sys.platform
        )

NUMBER_OF_CPUS = cpu_count()

# ------------------------------------------------------------------------------
# Utility Functions
# ------------------------------------------------------------------------------

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

def sudo(*command, **kwargs):
    response = query(
        "\tsudo %s\n\nDo you want to run the above command" % ' '.join(command),
        )
    if response.lower().startswith('y'):
        return do('sudo', *command, **kwargs)

def mkdir(path):
    if not isdir(path):
        try:
            makedirs(path)
        except OSError, e:
            if e.errno == EACCES:
                error("ERROR: Permission denied to create %s" % path)
                done = sudo('mkdir', '-p', path, retcode=True)
                if not done:
                    raise e
                return 1
            raise
        return 1

def rmdir(path, name=None):
    try:
        rmtree(path)
    except IOError:
        pass
    except OSError, e:
        if e.errno != ENOENT:
            if not name:
                name = path
            exit("ERROR: Couldn't remove the %s directory." % name)

# ------------------------------------------------------------------------------
# Constants
# ------------------------------------------------------------------------------

ROOT = dirname(dirname(dirname(realpath(__file__))))
ENVIRON = join(ROOT, 'environ')
LOCAL = join(ENVIRON, 'local')
BIN = join(LOCAL, 'bin')
INCLUDE = join(LOCAL, 'include')
INFO = join(LOCAL, 'share', 'info')
LIB = join(LOCAL, 'lib')
RECEIPTS = join(LOCAL, 'share', 'installed')
TMP = join(LOCAL, 'tmp')
VAR = join(LOCAL, 'var')

DISTFILES_URL = environ.get(
    'AMPIFY_DISTFILES_URL',
    "http://cloud.github.com/downloads/tav/ampify/distfile."
    )

# The HTTPS url -- which would be using S3 and not Amazon cloudfront would be:
# https://github.s3.amazonaws.com/downloads/tav/ampify/distfile.

BUILD_WORKING_DIRECTORY = join(ROOT, '.build')
BUILD_RECIPES = [path for path in environ.get(
    'AMPIFY_BUILD_RECIPES', join(ENVIRON, 'buildrecipes')
    ).split(':') if isfile(path)]

ROLES_PATH = [path for path in environ.get(
    'AMPIFY_ROLES_PATH', join(ENVIRON, 'roles')
    ).split(':') if isdir(path)]

CURRENT_DIRECTORY = getcwd()
HOME = expanduser('~')

if PLATFORM == 'darwin':
    LIB_EXTENSION = '.dylib'
elif PLATFORM == 'windows':
    LIB_EXTENSION = '.dll'
else:
    LIB_EXTENSION = '.so'

if PLATFORM == 'freebsd':
    MAKE = 'gmake'
else:
    MAKE = 'make'

RECIPES = {}
BUILTINS = locals()

RECIPES_INITIALISED = []
VIRGIN_BUILD = not exists(join(LOCAL, 'bin', 'python'))

# DISTFILES_SERVER = (
#     "http://cloud.github.com/downloads/tav/plexnet/distfile."
#     "%(name)s-%(version)s.tar.bz2"
#     )

# ------------------------------------------------------------------------------
#
# ------------------------------------------------------------------------------

# ------------------------------------------------------------------------------
# JSON Support
# ------------------------------------------------------------------------------

# We implement super-basic JSON decoding support for older Pythons which don't
# have a JSON module.
if not decode_json:

    import re

    JSON_ENV = {
        '__builtins__': None,
        'null': None,
        'true': True,
        'false': False
        }

    replace_string = re.compile(r'("(\\.|[^"\\])*")').sub
    match_json = re.compile(r'^[,:{}\[\]0-9.\-+Eaeflnr-u \n\r\t]+$').match

    def decode_json(json):
        if match_json(replace_string('', json)):
            json = replace_string(lambda m: 'u' + m.group(1), json)
            return eval(json.strip(' \t\r\n'), JSON_ENV, {})
        raise ValueError("Couldn't decode JSON input.")

# ------------------------------------------------------------------------------
# Distfiles Downloader
# ------------------------------------------------------------------------------

def download_distfile(distfile):
    pass

# ------------------------------------------------------------------------------
# Instance Roles
# ------------------------------------------------------------------------------

ROLES = {}

def load_role(role, debug=False):

    init_build_recipes(debug)
    if role in ROLES:
        return ROLES[role]

    if VIRGIN_BUILD and role != 'base':
        build_base_and_reload(debug)

    for path in ROLES_PATH:
        role_file = join(path, role) + '.json'
        if isfile(role_file):
            break
    else:
        exit("ERROR: Couldn't find a data file for the %r role." % role)

    try:
        role_file = open(role_file, 'rb')
    except IOError, error:
        exit("ERROR: %s: %s" % (error[1], error.filename))

    role_data = role_file.read()
    role_file.close()

    role_data = decode_json(role_data)
    packages = set(role_data['packages'])

    for package in packages:
        install_package(package)

    if 'requires' in role_data:
        packages.update(load_role(role_data['requires']))

    return ROLES.setdefault(role, packages)

# ------------------------------------------------------------------------------
# Build Recipes Initialiser
# ------------------------------------------------------------------------------

def init_build_recipes(debug=False):
    if RECIPES_INITIALISED:
        return
    for recipe in BUILD_RECIPES:
        execfile(recipe, BUILTINS)
    RECIPES_INITIALISED.append(1)

def install_package(package, debug=False):
    print "Install", package
    return

    # git --version
    if package not in RECIPES:
        exit(
            "ERROR: Couldn't find a build recipe for the %s package."
            % package
            )

def install_packages(debug=False):
    mkdir(BUILD_WORKING_DIRECTORY)
    mkdir('/opt/ampify')
    rmdir('/opt/ampify')
    log('installing...')

def build_base_and_reload(debug=False):
    load_role('base')
    install_packages()
    execve(join(ENVIRON, 'amp'), sys.argv, {})
