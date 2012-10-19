# Public Domain (-) 2004-2012 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

"""
The redpill bootstraps you into the world of Ampify.

                        _         _ _ _ 
           _ __ ___  __| |  _ __ (_) | |
          | '__/ _ \/ _` | | '_ \| | | |
          | | |  __/ (_| | | |_) | | | |
          |_|  \___|\__,_| | .__/|_|_|_|
                           |_|          
"""

import os
import sys
import subprocess
import tarfile
import traceback

from errno import EACCES, ENOENT
from glob import glob
from hashlib import sha1, sha256
from optparse import OptionParser
from os import chdir, getcwd, environ, execve, listdir, makedirs, remove, stat
from os.path import exists, isabs, isdir, isfile, islink, join
from shutil import copy, copytree, rmtree
from stat import ST_MTIME
from thread import start_new_thread
from time import sleep
from urllib import urlopen

try:
    from simplejson import loads as decode_json
except ImportError:
    from json import loads as decode_json

try:
    from tavutil.optcomplete import autocomplete, ListCompleter
    from tavutil.optcomplete import make_autocompleter, parse_options
except ImportError:
    autocomplete = None

try:
    from yaml import safe_load as decode_yaml
except ImportError:
    decode_yaml = None

try:
    from multiprocessing import cpu_count
except ImportError:
    cpu_count = lambda: 1

# ------------------------------------------------------------------------------
# Metadata
# ------------------------------------------------------------------------------

__version__ = (0, 1, 0)
__release__ = '.'.join(map(str, __version__))

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

# -----------------------------------------------------------------------------
# Fill in for tavutil.optcomplete.parse_options
# -----------------------------------------------------------------------------

if not autocomplete:
    class CompletionResult(Exception):
        def __init__(self, result):
            self.result = result

    def parse_options(parser, argv, completer=None, exit_if_no_args=False):
        if completer:
            raise CompletionResult(parser)
        if (argv == ['--help']) or (argv == ['-h']):
            parser.print_help()
            sys.exit(1)
        options, args = parser.parse_args(argv)
        if exit_if_no_args and not args:
            parser.print_help()
            sys.exit(1)
        return options, args

# -----------------------------------------------------------------------------
# Fill in for yaml.loads
# -----------------------------------------------------------------------------

if not decode_yaml:
    def decode_yaml(s):
        data = {}
        for line in s.splitlines():
            line = line.strip()
            if line.startswith('#'):
                continue
            if line.endswith(':'):
                data[line[:-1]] = xs = []
                continue
            if line.startswith('-'):
                xs.append(line[1:].strip())
        return data

# -----------------------------------------------------------------------------
# Command Execution
# -----------------------------------------------------------------------------

def exit_cmd(message, error_code=1):
    """Write an error message to stderr and exit with the given error_code."""
    sys.stderr.write(message + '\n')
    sys.exit(error_code)

class CommandNotFound(Exception):
    """Exception raised when a command line app could not be found."""

def run_command(
    args, retcode=False, reterror=False, exit_on_error=False, error_message="",
    log=None, redirect_stdout=True, redirect_stderr=True, cwd=None,
    shell=sys.platform.startswith('win'), env=None, universal_newlines=True
    ):
    """Execute the command with the given options."""

    log_message = "%s cwd=%s" % (' '.join(args), cwd or getcwd())
    if log:
        if hasattr(log, '__call__'):
            log(log_message)
        else:
            sys.stderr.write("Running command: " + log_message + '\n')

    if redirect_stdout:
        stdout = subprocess.PIPE
    else:
        stdout = None

    if redirect_stderr:
        stderr = subprocess.PIPE
    else:
        stderr = None

    try:
        process = subprocess.Popen(
            args, stdout=stdout, stderr=stderr, shell=shell, cwd=cwd, env=env,
            universal_newlines=universal_newlines
            )
        out, err = process.communicate()
    except OSError:
        error = sys.exc_info()[1]
        if error.errno == 2:
            if not error.filename:
                if exit_on_error:
                    exit_cmd("Couldn't find the %r command!" % args[0])
                raise CommandNotFound(args[0])
            raise error
        if exit_on_error:
            exit_cmd("Error running: %s\n\n%s" % (log_message, error_message))
        raise

    if process.returncode and exit_on_error:
        if stderr:
            exit_extra = error_message or err
        else:
            exit_extra = error_message or out
        if exit_extra:
            exit_cmd("Error running: %s\n\n%s" % (log_message, exit_extra))
        else:
            exit_cmd("Error running: %s" % log_message)

    if retcode:
        if reterror:
            return out, err, process.returncode
        return out, process.returncode

    if reterror:
        return out, err
    return out

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

def mkdir(path, sudo=False):
    if not isdir(path):
        try:
            makedirs(path)
        except OSError, e:
            if (e.errno == EACCES) and sudo:
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

LOCKS = {}

def lock(path):
    LOCKS[path] = lock = open(path, 'w')
    try:
        from fcntl import flock, LOCK_EX, LOCK_NB
    except ImportError:
        exit("ERROR: Locking is not supported on this platform.")
    try:
        flock(lock.fileno(), LOCK_EX | LOCK_NB)
    except Exception:
        exit("ERROR: Another redpill process is already running.")

def unlock(path):
    if path in LOCKS:
        LOCKS[path].close()
        del LOCKS[path]

# Collate the set of resources within the given ``directory``.
def gather_local_filelisting(directory, gathered=None):
    if gathered is None:
        if not isdir(directory):
            return set()
        gathered = set()
    for item in listdir(directory):
        path = join(directory, item)
        if isdir(path):
            gathered.add(path + '/')
            gather_local_filelisting(path, gathered)
        else:
            gathered.add(path)
    return gathered

# Strip the given ``prefix`` from the elements in the given ``listing`` set.
def strip_prefix(listing, prefix):
    new = set()
    lead = len(prefix) + 1
    for item in listing:
        new.add(item[lead:])
    return new

def get_listing():
    return strip_prefix(gather_local_filelisting(LOCAL), LOCAL)

def cleanup_partial_install(current_filelisting):
    new_filelisting = get_listing()
    diff = new_filelisting.difference(current_filelisting)
    for file in diff:
        file = join(LOCAL, file)
        remove(file)

# ------------------------------------------------------------------------------
# Constants
# ------------------------------------------------------------------------------

ROOT = environ['AMPIFY_ROOT']
LOCAL = environ['AMPIFY_LOCAL']

ENVIRON = join(ROOT, 'environ')
SHARE = join(LOCAL, 'share')

BIN = join(LOCAL, 'bin')
INCLUDE = join(LOCAL, 'include')
INFO = join(SHARE, 'info')
LIB = join(LOCAL, 'lib')
MAN = join(SHARE, 'man')
RECEIPTS = join(ENVIRON, 'receipts')
SRC = join(ROOT, 'src')
THIRD_PARTY = join(ROOT, 'third_party')
TMP = join(LOCAL, 'tmp')
VAR = join(LOCAL, 'var')

DISTFILES_URL_BASE = environ.get(
    'REDPILL_DISTFILES_URL_BASE',
    "http://cloud.github.com/downloads/tav/ampify/distfile."
    )

BUILD_WORKING_DIRECTORY = '/tmp/redpill-%s' % sha1(ROOT).hexdigest()[:8]
BUILD_LOCK = BUILD_WORKING_DIRECTORY + '.lock'

BUILD_RECIPES = [path for path in environ.get(
    'REDPILL_BUILD_RECIPES', join(ENVIRON, 'buildrecipes')
    ).split(':') if isfile(path)]

ROLES_PATH = [path for path in environ.get(
    'REDPILL_ROLES_PATH', join(ENVIRON, 'roles')
    ).split(':') if isdir(path)]

CURRENT_DIRECTORY = getcwd()

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

if sys.maxint == 2**63 - 1:
    ARCH = 'amd64'
else:
    ARCH = '386'

CPPFLAGS = "-I%s" % INCLUDE
LDFLAGS = "-L%s" % LIB

RECIPES = {}
BUILTINS = locals()

BASE_PACKAGES = set()
PACKAGES = {}
RECIPES_INITIALISED = []

DEBUG = False

# ------------------------------------------------------------------------------
# Distfiles Downloader
# ------------------------------------------------------------------------------

DOWNLOAD_QUEUE = []
DOWNLOAD_ERROR = []

class DownloadError(Exception):
    def __init__(self, msg):
        self.msg = msg

# Download the given distfile and ensure it has a matching digest. We try to
# capture and exit on all errors to avoid them being silently ignored in a
# separate thread.
def _download_distfile(distfile, url, hash, dest):
    try:
        try:
            distfile_obj = urlopen(url)
            distfile_source = distfile_obj.read()
        except Exception:
            raise DownloadError("Failed to download %s" % distfile)
        if sha256(distfile_source).hexdigest() != hash:
            raise DownloadError("Got an invalid hash digest for %s" % distfile)
        try:
            distfile_file = open(dest, 'wb')
            distfile_file.write(distfile_source)
            distfile_file.close()
            distfile_obj.close()
        except Exception:
            raise DownloadError("Writing %s" % distfile)
        DOWNLOAD_QUEUE.pop()
    except DownloadError, errmsg:
        DOWNLOAD_QUEUE.pop()
        DOWNLOAD_ERROR.append(errmsg)

# Check if there's an existing valid download. If not, fire off a fresh download
# in a separate thread if the ``fork`` parameter has been set.
def download_distfile(distfile, url, hash, fork=False):
    dest = join(BUILD_WORKING_DIRECTORY, distfile)
    if isfile(dest):
        log("Verifying existing %s" % distfile, PROGRESS)
        distfile_file = open(dest, 'rb')
        distfile_source = distfile_file.read()
        distfile_file.close()
        if sha256(distfile_source).hexdigest() == hash:
            return
        remove(dest)
    log("Downloading %s" % distfile, PROGRESS)
    DOWNLOAD_QUEUE.append(distfile)
    if fork:
        start_new_thread(_download_distfile, (distfile, url, hash, dest))
    else:
        _download_distfile(distfile, url, hash, dest)

# ------------------------------------------------------------------------------
# Instance Roles
# ------------------------------------------------------------------------------

ROLES = {}

def load_role(role):

    init_build_recipes()
    if role in ROLES:
        return ROLES[role]

    for path in ROLES_PATH:
        role_file = join(path, role) + '.yaml'
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

    try:
        role_data = decode_yaml(role_data)
    except Exception:
        exit("ERROR: Couldn't decode the JSON input: %s" % role_file.name)

    packages = set(role_data['packages'])
    for package in packages:
        install_package(package)

    if 'requires' in role_data:
        packages.update(load_role(role_data['requires']))

    if role == 'base':
        for package in packages:
            BASE_PACKAGES.update([package])
            BASE_PACKAGES.update(get_dependencies(package))

    return ROLES.setdefault(role, packages)

def get_dependencies(package, gathered=None):
    if gathered is None:
        gathered = set()
    else:
        gathered.add(package)
    recipe = RECIPES[package][PACKAGES[package][0]]
    for dep in recipe.get('requires', []):
        get_dependencies(dep, gathered)
    return gathered

# ------------------------------------------------------------------------------
# Version Checkers
# ------------------------------------------------------------------------------

def ensure_gcc_version(version=(4, 0)):
    gcc = environ.get('CC', 'gcc')
    try:
        ver = do(
            gcc, '-dumpversion', redirect_stdout=True, reterror=True
            )
        ver = tuple(map(int, ver[0].strip().split('.')))
        if ver < version:
            raise RuntimeError("Invalid version")
    except Exception:
        exit('ERROR: GCC %s+ not found!' % '.'.join(map(str, version)))

def ensure_git_version(version=(1, 7)):
    try:
        ver = do(
            'git', '--version', redirect_stdout=True,
            redirect_stderr=True, reterror=True
            )
        ver = ver[0].splitlines()[0].split()[2]
        ver = tuple(int(part) for part in ver.split('.'))
        if ver < version:
            raise RuntimeError("Invalid version")
    except Exception:
        exit('ERROR: Git %s+ not found!' % '.'.join(map(str, version)))

def ensure_java_version(version=(1, 6), title='Java 6+ runtime'):
    try:
        ver = do(
            'java', '-version', redirect_stdout=True,
            redirect_stderr=True, reterror=True
            )
        ver = ver[1].splitlines()[0].split()[-1][1:-1]
        if not ver >= '.'.join(map(str, version)):
            raise RuntimeError("Invalid version")
    except Exception:
        exit('ERROR: %s not found!' % title)

def ensure_node_version(version=(0, 8, 2)):
    try:
        ver = do(
            'node', '-v', redirect_stdout=True,
            redirect_stderr=True, reterror=True
            )
        ver = tuple(map(int, ver[0][1:].strip().split('.')))
        if ver < version:
            raise RuntimeError("Invalid version")
    except Exception:
        exit('ERROR: Node.js %s+ not found!' % '.'.join(map(str, version)))

def ensure_ruby_version(version=(1, 8, 7)):
    try:
        ver = do(
            'ruby', '-v', redirect_stdout=True,
            redirect_stderr=True, reterror=True
            )
        ver = tuple(map(int, ver[0].strip().split()[1].strip().split('.')))
        if ver < version:
            raise RuntimeError("Invalid version")
    except Exception:
        exit('ERROR: Ruby %s+ not found!' % '.'.join(map(str, version)))

# ------------------------------------------------------------------------------
# Build Recipes Initialiser
# ------------------------------------------------------------------------------

def init_build_recipes():
    if RECIPES_INITIALISED:
        return
    # Try getting a lock to avoid concurrent builds.
    lock(BUILD_LOCK)
    mkdir(RECEIPTS)
    for recipe in BUILD_RECIPES:
        execfile(recipe, BUILTINS)
    for package in list(RECIPES):
        recipes = RECIPES[package]
        versions = []
        data = {}
        for recipe in recipes:
            recipe_type = recipe.get('type')
            if recipe_type == 'git':
                path = join(ROOT, recipe['path'])
                version = run_command(
                    ['git', 'rev-parse', 'HEAD'], cwd=path, exit_on_error=True
                    ).strip()
            elif 'depends' in recipe:
                contents = {}
                latest = 0
                for pattern in recipe['depends']:
                    for file in glob(pattern):
                        dep_file = open(file, 'rb')
                        contents[file] = dep_file.read()
                        dep_file.close()
                        dep_mtime = stat(file)[ST_MTIME]
                        if dep_mtime > latest:
                            latest = dep_mtime
                generate = 0
                for pattern in recipe['outputs']:
                    files = glob(pattern)
                    if not files:
                        generate = 1
                        break
                    for file in files:
                        if not isfile(file):
                            generate = 1
                            break
                        if stat(file)[ST_MTIME] <= latest:
                            generate = 1
                            break
                    if generate:
                        break
                if generate:
                    for file in listdir(RECEIPTS):
                        if file.startswith(package + '-'):
                            remove(join(RECEIPTS, file))
                version = sha1(''.join([
                    '%s\x00%s' % (f, contents[f])
                    for f in sorted(contents)
                    ])).hexdigest()
            else:
                version = recipe['version']
            versions.append(version)
            data[version] = recipe
        RECIPES[package] = data
        PACKAGES[package] = versions
    RECIPES_INITIALISED.append(1)

# ------------------------------------------------------------------------------
# Env Manipulation
# ------------------------------------------------------------------------------

def get_redpill_env(environ):
    new = {}
    for key in environ:
        if key.startswith('AMPIFY') or key.startswith('REDPILL'):
            new[key] = environ[key]
        elif key.startswith('JAVA') or key.endswith('JAVA'):
            new[key] = environ[key]
    for var in [
        'PATH', 'LD_LIBRARY_PATH',
        'DYLD_FALLBACK_LIBRARY_PATH',
        'MANPATH', 'PYTHONPATH'
        ]:
        pre_var = 'PRE_AMPENV_' + var
        if pre_var in environ:
            new[var] = environ[pre_var]
    return new

# ------------------------------------------------------------------------------
# Build Types
# ------------------------------------------------------------------------------

BASE_BUILD = {
    'after': None,
    'before': None,
    'commands': None,
    'distfile': "%(name)s-%(version)s.tar.bz2",
    'distfile_url_base': DISTFILES_URL_BASE,
    'env': None,
    }

def default_build_commands(package, info):
    commands = []; add = commands.append
    if info['config_command']:
        add([info['config_command']] + info['config_flags'])
    if info['separate_make_install']:
        add([MAKE])
    add([MAKE] + info['make_flags'])
    return commands

DEFAULT_BUILD = BASE_BUILD.copy()
DEFAULT_BUILD.update({
    'commands': default_build_commands,
    'config_command': './configure',
    'config_flags': ['--prefix=%s' % LOCAL],
    'make_flags': ['install'],
    'separate_make_install': False
    })

PYTHON_BUILD = BASE_BUILD.copy()
PYTHON_BUILD.update({
    'commands': [
        [sys.executable, 'setup.py', 'build_ext', '-i']
        ]
    })

def resource_build_commands(package, info):
    source = info['source'] or join(BUILD_WORKING_DIRECTORY, package)
    destination = info['destination'] or join(SHARE, package)
    return [['cp', '-R', source, destination]]

RESOURCE_BUILD = BASE_BUILD.copy()
RESOURCE_BUILD.update({
    'commands': resource_build_commands,
    'source': None,
    'destination': None,
    })

def jar_install(package, info):
    filename = info['distfile'] % {'name': package, 'version': info['version']}
    return [lambda: copy(filename, join(BIN, filename))]

JAR_BUILD = BASE_BUILD.copy()
JAR_BUILD.update({
    'distfile': '%(name)s-%(version)s.jar',
    'commands': jar_install
    })

GIT_BUILD = BASE_BUILD.copy()
GIT_BUILD.update({
    'distfile': ''
    })

MAKELIKE_BUILD = BASE_BUILD.copy()
MAKELIKE_BUILD.update({
    'distfile': ''
    })

BUILD_TYPES = {
    'default': DEFAULT_BUILD,
    'git': GIT_BUILD,
    'jar': JAR_BUILD,
    'makelike': MAKELIKE_BUILD,
    'python': PYTHON_BUILD,
    'resource': RESOURCE_BUILD
    }

# ------------------------------------------------------------------------------
# Core Installer Functionality
# ------------------------------------------------------------------------------

TO_INSTALL = {}
TO_UNINSTALL = {}

def get_installed_packages(called=[], cache={}):
    if called:
        return cache
    called.append(1)
    cache.update(dict(f.split('-', 1) for f in listdir(RECEIPTS)))
    return cache

def get_installed_dependencies(
    package, gathered=None, raw_types=['git', 'makelike']
    ):
    if gathered is None:
        gathered = set()
    else:
        gathered.add(package)
    installed_version = get_installed_packages()[package]
    recipes = RECIPES[package]
    if recipes.values()[0].get('type') in raw_types:
        recipe = recipes.values()[0]
    else:
        recipe = recipes[installed_version]
    for dep in recipe.get('requires', []):
        get_installed_dependencies(dep, gathered)
    return gathered

def get_installed_data():
    installed = get_installed_packages()
    inverse_dependencies = {}
    for package in installed:
        if package in PACKAGES:
            for dep in get_installed_dependencies(package):
                if dep not in inverse_dependencies:
                    inverse_dependencies[dep] = set()
                inverse_dependencies[dep].add(package)
    return installed, inverse_dependencies

# Check and load the build recipe for the given package name and add it to the
# ``TO_INSTALL`` set.
def install_package(package):
    if package not in RECIPES:
        exit(
            "ERROR: Couldn't find a build recipe for the %s package."
            % package
            )
    version = PACKAGES[package][0]
    TO_INSTALL[package] = version
    for dependency in RECIPES[package][version].get('requires', []):
        install_package(dependency)

# Handle the actual installation/uninstallation of appropriate packages.
def install_packages(types=BUILD_TYPES):

    ensure_gcc_version()
    ensure_git_version()

    for directory in [
        BUILD_WORKING_DIRECTORY, LOCAL, BIN, SHARE, TMP
        ]:
        mkdir(directory)

    cleanup_install()

    # We assume the invariant that all packages only have one version installed.
    installed, inverse_dependencies = get_installed_data()
    uninstall = set()

    for package in TO_INSTALL:
        if package in installed:
            existing_version = installed[package]
            if TO_INSTALL[package] != existing_version:
                uninstall.add(package)
                for inv_dep in inverse_dependencies.get(package, []):
                    uninstall.add(package)

    # If a base package needs to be uninstalled, just nuke environ/local and
    # rebuild everything from scratch.
    for package in BASE_PACKAGES:
        if package in uninstall:
            rmdir(LOCAL)
            rmdir(RECEIPTS)
            unlock(BUILD_LOCK)
            execve(join(ENVIRON, 'redpill'), sys.argv, get_redpill_env(environ))

    if uninstall:
        for package in uninstall:
            uninstall_package(package)
        uninstall_packages()

    to_install = set(TO_INSTALL) - set(installed)

    inverse_dependencies = {}
    for package in to_install:
        for dep in get_dependencies(package):
            if dep in to_install:
                if dep not in inverse_dependencies:
                    inverse_dependencies[dep] = set()
                inverse_dependencies[dep].add(package)

    to_install_list = []
    for package in to_install:
        index = len(to_install_list)
        for dep in inverse_dependencies.get(package, []):
            try:
                dep_index = to_install_list.index(dep)
            except:
                continue
            else:
                if dep_index < index:
                    index = dep_index
        to_install_list.insert(index, package)

    current_filelisting = get_listing()
    install_data = []
    install_items = len(to_install_list) - 1

    for idx, package in enumerate(to_install_list):

        version = TO_INSTALL[package]
        recipe = RECIPES[package][version]
        build_type = recipe.get('type', 'default')
        info = types[build_type].copy()
        info.update(recipe)

        distfile = info['distfile'] % {'name': package, 'version': version}
        if distfile:
            url = info['distfile_url_base'] + distfile
        else:
            url = ''

        install_data.append((idx, package, version, info, distfile, url))

    if install_data:
        _, _, _, info, distfile, url = install_data[0]
        if distfile:
            download_distfile(distfile, url, info['hash'])

    for idx, package, version, info, distfile, url in install_data:

        current_queue = DOWNLOAD_QUEUE[:]
        if current_queue:
            if idx:
                log("Waiting for %s to download" % current_queue[0], PROGRESS)
            while DOWNLOAD_QUEUE:
                sleep(0.5)

        if DOWNLOAD_ERROR:
            exit("ERROR: %s" % DOWNLOAD_ERROR[0].msg)

        if idx < install_items:
            _, _, _, infoN, distfileN, urlN = install_data[idx+1]
            if distfileN:
                download_distfile(distfileN, urlN, infoN['hash'], fork=True)

        log("Installing %s %s" % (package, version))

        chdir(BUILD_WORKING_DIRECTORY)
        if distfile and distfile.endswith('.tar.bz2'):
            if isdir(package):
                log("Removing previously unpacked %s distfile" % package,
                    PROGRESS)
                rmdir(package)
            tar = tarfile.open(distfile, 'r:bz2')
            log("Unpacking %s" % distfile, PROGRESS)
            tar.extractall()
            tar.close()
            chdir(package)
        elif info.get('type') == 'git':
            chdir(join(ROOT, info['path']))
            if info.get('clean'):
                do('git', 'clean', '-fdx')

        if info['before']:
            info['before']()

        env = environ.copy()
        if 'MAKE' in env:
            del env['MAKE']
        if 'MAKELEVEL' in env:
            del env['MAKELEVEL']
        if info['env']:
            env.update(info['env'])

        commands = info['commands']
        if isinstance(commands, basestring):
            commands = [commands]
        elif hasattr(commands, '__call__'):
            try:
                commands = commands(package, info)
            except Exception:
                log("ERROR: Error calling build command for %s %s" %
                    (package, version))
                traceback.print_exc()
                sys.exit(1)

        if not isinstance(commands, (tuple,list)):
            exit("ERROR: Invalid build commands for %s %s: %r" %
                 (package, version, commands))

        try:
            for command in commands:
                if hasattr(command, '__call__'):
                    command()
                else:
                    log("Running: %s" % ' '.join(command), PROGRESS)
                    cmd_env = {'CPPFLAGS': CPPFLAGS, 'LDFLAGS': LDFLAGS}
                    cmd_env.update(env)
                    kwargs = dict(env=cmd_env)
                    do(*command, **kwargs)
        except Exception:
            error("ERROR: Building %s %s failed" % (package, version))
            traceback.print_exc()
            sys.exit(1)
        except SystemExit:
            cleanup_partial_install(current_filelisting)
            exit("ERROR: Building %s %s failed" % (package, version))

        if info['after']:
            info['after']()

        log("Successfully Installed %s %s" % (package, version), SUCCESS)

        new_filelisting = get_listing()
        receipt_data = new_filelisting.difference(current_filelisting)
        current_filelisting = new_filelisting

        receipt = open(join(RECEIPTS, '%s-%s' % (package, version)), 'wb')
        receipt.write('\n'.join(sorted(receipt_data)))
        receipt.close()

        chdir(BUILD_WORKING_DIRECTORY)
        if distfile.endswith('.tar.bz2'):
            rmdir(join(package))

    chdir(CURRENT_DIRECTORY)

# A utility function to uninstall a single package.
def uninstall_package(package):
    installed = get_installed_packages()
    if package in installed:
        TO_UNINSTALL[package] = installed[package]

# Handle the actual uninstallation of the various packages.
def uninstall_packages():
    installed = get_installed_packages()
    for name, version in TO_UNINSTALL.iteritems():
        log("Uninstalling %s %s" % (name, version))
        installed_version = '%s-%s' % (name, version)
        receipt_path = join(RECEIPTS, installed_version)
        receipt = open(receipt_path, 'rb')
        directories = set()
        for path in receipt:
            if isabs(path):
                exit("ERROR: Got an absolute path in receipt %s" % receipt_path)
            path = path.strip()
            if not path:
                continue
            path = join(LOCAL, path)
            if not islink(path):
                if not exists(path):
                    continue
            if isdir(path):
                directories.add(path)
            else:
                print "Removing:", path
                remove(path)
        for path in reversed(sorted(directories)):
            if not listdir(path):
                print "Removing Directory:", path
                rmtree(path)
        receipt.close()
        remove(receipt_path)
        del installed[name]

def cleanup_install():
    current = get_listing()
    expected = set()
    for f in listdir(RECEIPTS):
        f = open(join(RECEIPTS, f), 'rb')
        expected.update(f.read().splitlines())
        f.close()
    diff = current.difference(expected)
    for path in diff:
        if isabs(path):
            continue
        path = join(LOCAL, path)
        if not isdir(path):
            print "Removing:", path
            remove(path)

# ------------------------------------------------------------------------------
# Virgin Build Handler
# ------------------------------------------------------------------------------

VIRGIN_BUILD = False # not exists(join(LOCAL, 'bin', 'bsdiff'))

def build_base_and_reload():
    load_role('base')
    install_packages()
    unlock(BUILD_LOCK)
    execve(join(ENVIRON, 'redpill'), sys.argv, get_redpill_env(environ))

# ------------------------------------------------------------------------------
# Main Runner
# ------------------------------------------------------------------------------

def main(argv=None, show_help=False):

    argv = argv or sys.argv[1:]

    # Set the script name to ``redpill`` so that OptionParser error messages
    # don't display a potentially confusing ``redpill.py`` to end users.
    sys.argv[0] = 'redpill'

    major_listing = '\n'.join(
        "    %-10s %s"
        % (cmd, MAJOR_COMMANDS[cmd].__doc__) for cmd in sorted(MAJOR_COMMANDS)
        )

    mini_listing = '\n'.join(
        "    %-10s %s"
        % (cmd, MINI_COMMANDS[cmd].__doc__) for cmd in sorted(MINI_COMMANDS)
        )

    usage = ("""%s\nUsage: redpill <command> [options]
    \nCommands:
    \n%s\n\n%s
    \nSee `redpill help <command>` for more info on a specific command.""" %
    (__doc__, major_listing, mini_listing))

    if autocomplete:
        autocomplete(
            OptionParser(add_help_option=False),
            ListCompleter(AUTOCOMPLETE_COMMANDS.keys()),
            subcommands=AUTOCOMPLETE_COMMANDS
            )
    elif 'OPTPARSE_AUTO_COMPLETE' in environ:
        sys.exit(1)

    if not argv:
        show_help = True
    else:
        command = argv[0]
        argv = argv[1:]
        if command in ['help', '-h', '--help']:
            if argv:
                command = argv[0]
                argv = ['--help']
                if command in MINI_COMMANDS:
                    help = MINI_COMMANDS[command].__doc__
                    print "Usage: redpill %s\n\n    %s\n" % (command, help)
                    sys.exit()
            else:
                show_help = True
        elif command in ['-v', '--version']:
            version()
            sys.exit()
        elif command in MINI_COMMANDS:
            MINI_COMMANDS[command]()
            sys.exit()

    if show_help:
        print usage
        sys.exit()

    if command in MAJOR_COMMANDS:
        return MAJOR_COMMANDS[command](argv)

    # We support git-command like behaviour. That is, if there's an external
    # binary named ``redpill-foo`` available on the ``$PATH``, then running ``redpill
    # foo`` will automatically delegate to it.
    try:
        output, retcode = run_command(
            ['redpill-%s' % command] + argv, retcode=True, redirect_stdout=False,
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
    """download and build the dependencies"""

    usage = "Usage: redpill build [options]\n\n    %s" % build.__doc__

    op = OptionParser(usage=usage, add_help_option=False)

    op.add_option('--role', dest='role', default=get_role(),
                  help="specify the role to build [%s]" % get_role())

    options, args = parse_options(op, argv, completer)

    if VIRGIN_BUILD and options.role != 'base':
        build_base_and_reload()

    load_role(options.role)
    install_packages()

# ------------------------------------------------------------------------------
# Check Command
# ------------------------------------------------------------------------------

def check():
    """check if this checkout is up-to-date"""

    log("Checking the current revision id for your code.", PROGRESS)
    revision_id = do(
        'git', 'show', '--pretty=oneline', '--summary', redirect_stdout=True
        ).split()[0]

    log("Checking the latest commits on GitHub.", PROGRESS)
    commit_info = urlopen(environ.get(
        'REDPILL_CHECK_URL',
        'https://api.github.com/repos/tav/ampify/branches/master'
        )).read()

    latest_revision_id = decode_json(commit_info)['commit']['sha']

    if revision_id != latest_revision_id:
        exit("A new version is available. Please run `git update`.")

    log("Your checkout is up-to-date.", SUCCESS)

# ------------------------------------------------------------------------------
# Info Utilities
# ------------------------------------------------------------------------------

def get_role():
    role = environ.get('REDPILL_ROLE', 'default')
    if PLATFORM == 'freebsd' and role == 'default':
        return 'freebsd'
    return role

def get_build_info():
    roles = set()
    for path in ROLES_PATH:
        roles.update([f[:-5] for f in listdir(path) if f.endswith('.yaml')])
    roles = list(sorted(roles))
    for role in roles:
        load_role(role)
    stream = []; write = stream.append
    write('\t\t')
    write(':'.join(roles))
    write('\n')
    for package in sorted(TO_INSTALL):
        write(package)
        write('\t\t')
        write(TO_INSTALL[package])
        write('\n')
    while stream[-1] == '\n':
        stream.pop()
    return ''.join(stream)

def get_installed_info():
    stream = []; write = stream.append
    installed = get_installed_packages()
    for package in sorted(installed):
        write(package)
        write('\t\t')
        write(installed[package])
        write('\n')
    while stream and stream[-1] == '\n':
        stream.pop()
    return ''.join(stream)

# ------------------------------------------------------------------------------
# Info Command
# ------------------------------------------------------------------------------

def info(argv=None, completer=None):
    """show metadata relating to the installs"""

    usage = "Usage: redpill info [options]\n\n    %s" % info.__doc__
    op = OptionParser(usage=usage, add_help_option=False)

    op.add_option(
        '--hash', action='store_true',
        help="show the sha256 hash of the output"
        )

    op.add_option(
        '--installed', action='store_true',
        help="output the list of installed packages/versions"
        )

    op.add_option(
        '--role', action='store_true',
        help="output the default redpill role"
        )

    if completer:
        return op

    options, args = parse_options(op, argv, completer)

    lock(BUILD_LOCK)
    if options.role:
        output = get_role()
    elif options.installed:
        output = get_installed_info()
    else:
        output = get_build_info()

    if options.hash:
        output = sha256(output).hexdigest()

    print output
    unlock(BUILD_LOCK)

# ------------------------------------------------------------------------------
# Install Command
# ------------------------------------------------------------------------------

def install(argv=None, completer=None):
    """install specific build packages"""

    usage = (
        "Usage: redpill install [packages]\n\n    %s"
        % install.__doc__
        )

    op = OptionParser(usage=usage, add_help_option=False)

    init_build_recipes()
    if completer:
        installed_packages = get_installed_packages()
        potentials = [pkg for pkg in RECIPES if pkg not in installed_packages]
        return op, ListCompleter(potentials)

    options, args = parse_options(op, argv, completer, True)
    for package in args:
        install_package(package)

    install_packages()

# ------------------------------------------------------------------------------
# Nuke Command
# ------------------------------------------------------------------------------

def nuke(argv=None, completer=None):
    """nuke the local install"""

    usage = "Usage: redpill nuke [options]\n\n    %s" % nuke.__doc__
    op = OptionParser(usage=usage, add_help_option=False)

    if completer:
        return op

    options, args = parse_options(op, argv, completer)

    lock(BUILD_LOCK)
    rmdir(LOCAL)
    rmdir(RECEIPTS)
    unlock(BUILD_LOCK)

# ------------------------------------------------------------------------------
# Uninstall Command
# ------------------------------------------------------------------------------

def uninstall(argv=None, completer=None):
    """uninstall specific build packages"""

    usage = (
        "Usage: redpill uninstall [packages]\n\n    %s"
        % uninstall.__doc__
        )

    op = OptionParser(usage=usage, add_help_option=False)

    init_build_recipes()
    if completer:
        return op, ListCompleter(get_installed_packages())

    options, args = parse_options(op, argv, completer, True)
    for package in args:
        uninstall_package(package)

    uninstall_packages()

# ------------------------------------------------------------------------------
# Version Command
# ------------------------------------------------------------------------------

def version():
    """show the version number and exit"""

    print 'redpill %s' % __release__

# ------------------------------------------------------------------------------
# Command Mapping
# ------------------------------------------------------------------------------

MAJOR_COMMANDS = {
    'build': build,
    'info': info,
    'install': install,
    'nuke': nuke,
    'uninstall': uninstall
    }

MINI_COMMANDS = {
    'check': check,
    'version': version
    }

# ------------------------------------------------------------------------------
# Command Autocompletion
# ------------------------------------------------------------------------------

if autocomplete:

    AUTOCOMPLETE_COMMANDS = MAJOR_COMMANDS.copy()

    AUTOCOMPLETE_COMMANDS['help'] = lambda completer: (
        OptionParser(add_help_option=False),
        ListCompleter(MAJOR_COMMANDS.keys() + MINI_COMMANDS.keys())
        )

    no_autocomplete = lambda completer: (
        OptionParser(add_help_option=False), None
        )

    for command in MINI_COMMANDS:
        AUTOCOMPLETE_COMMANDS[command] = no_autocomplete

    for command in AUTOCOMPLETE_COMMANDS.values():
        command.autocomplete = make_autocompleter(command)

# ------------------------------------------------------------------------------
# Script Runner
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
