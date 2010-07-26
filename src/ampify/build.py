# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import sys

from os import getcwd, environ, execve, makedirs
from os.path import dirname, exists, expanduser, isdir, isfile, join, realpath

try:
    from multiprocessing import cpu_count
except ImportError:
    cpu_count = lambda: 1

try:
    from json import loads as decode_json
except ImportError:
    decode_json = None

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
    print (
        "ERROR: Sorry, the %r operating system is not supported yet."
        % sys.platform
        )
    sys.exit(1)

NUMBER_OF_CPUS = cpu_count()

# ------------------------------------------------------------------------------
# Utility Functions
# ------------------------------------------------------------------------------

def mkdir(path):
    if not isdir(path):
        makedirs(path)
        return 1

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
        print "ERROR: Couldn't find a data file for the %r role." % role
        sys.exit(1)

    role_file = open(role_file, 'rb')
    role_data = role_file.read()
    role_file.close()

    role_data = decode_json(role_data)
    packages = set(role_data['packages'])

    for package in packages:
        install_package(package)

    if 'depends' in role_data:
        packages.update(load_role(role_data['depends']))

    return ROLES.setdefault(role, packages)

    try:
        role_info_file = open(join_path(ROLES_DIRECTORY, '%s.role' % role), 'rb')
    except IOError, error:
        print_message("%s: %s" % (error[1], error.filename), ERROR)
        sys.exit(1)

def init_build_recipes(debug=False):
    if RECIPES_INITIALISED:
        return
    for recipe in BUILD_RECIPES:
        execfile(recipe, BUILTINS)
    print RECIPES.keys()
    RECIPES_INITIALISED.append(1)

def install_package(package, debug=False):
    if package not in RECIPES:
        print (
            "ERROR: Couldn't find a build recipe for the %s package."
            % package
            )
        sys.exit(1)

def install_packages(debug=False):
    mkdir(BUILD_WORKING_DIRECTORY)
    print 'installing...'

def build_base_and_reload(debug=False):
    load_role('base')
    install_packages()
    execve(join(ENVIRON, 'amp'), sys.argv, {})
