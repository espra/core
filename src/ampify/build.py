# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import os
import re
import sys
import tarfile
import traceback

from errno import EACCES, ENOENT
from hashlib import sha256
from os import chdir, getcwd, environ, execve, listdir, makedirs, remove
from os.path import dirname, exists, expanduser, isabs, isdir, isfile, islink
from os.path import join, realpath
from shutil import copy, rmtree
from thread import start_new_thread
from time import sleep
from urllib import urlopen

try:
    from multiprocessing import cpu_count
except ImportError:
    cpu_count = lambda: 1

try:
    from json import loads as decode_json
except ImportError:
    decode_json = None

from pyutil.env import run_command

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

#     if 'retcode' in kwargs:
#         return run_command(*args, **kwargs)
#     kwargs['retcode'] = 1
#     result = run_command(*args, **kwargs)
#     retcode = result[-1]
#     if retcode and not Options.options.force:
#         raise ValueError("Error running %s" % ' '.join(args[0]))
#     return result[:-1]

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
        exit("ERROR: Another amp process is already running.")

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

# ------------------------------------------------------------------------------
# Constants
# ------------------------------------------------------------------------------

ROOT = dirname(dirname(dirname(realpath(__file__))))
ENVIRON = join(ROOT, 'environ')
LOCAL = join(ENVIRON, 'local')
BIN = join(LOCAL, 'bin')
INCLUDE = join(LOCAL, 'include')
LIB = join(LOCAL, 'lib')
SHARE = join(LOCAL, 'share')
INFO = join(SHARE, 'info')
MAN = join(SHARE, 'man')
RECEIPTS = join(ENVIRON, 'receipts')
TMP = join(LOCAL, 'tmp')
VAR = join(LOCAL, 'var')

DISTFILES_URL_BASE = environ.get(
    'AMPIFY_DISTFILES_URL_BASE',
    "http://cloud.github.com/downloads/tav/ampify/distfile."
    )

DISTFILES_URL_BASE = environ.get(
    'AMPIFY_DISTFILES_URL_BASE',
    "http://github.s3.amazonaws.com/downloads/tav/ampify/distfile."
    )

# Alternatively, could use the HTTPS URL -- note that this would be using S3 and
# not Amazon CloudFront:
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

BASE_PACKAGES = set()
BUILD_LOCK = join(ROOT, '.build-lock')
PACKAGES = {}
RECIPES_INITIALISED = []
VIRGIN_BUILD = not exists(join(LOCAL, 'bin', 'python'))

# ------------------------------------------------------------------------------
# JSON Support
# ------------------------------------------------------------------------------

# We implement super-basic JSON decoding support for older Pythons which don't
# have a JSON module.
if not decode_json:
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
            json = replace_string(lambda m: m.group(1), json)
            return eval(json.strip(' \t\r\n'), JSON_ENV, {})
        raise ValueError("Couldn't decode JSON input.")

# ------------------------------------------------------------------------------
# Distfiles Downloader
# ------------------------------------------------------------------------------
#print '\n'.join(sorted(strip_prefix(gather_local_filelisting(LOCAL), LOCAL)))
#sys.exit()

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

    if VIRGIN_BUILD and role != 'base':
        build_base_and_reload()

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

    try:
        role_data = decode_json(role_data)
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
    try:
        ver = do(
            'gcc', '-dumpversion', redirect_stdout=True, reterror=True
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
        ver = ver[0].splitlines()[0].split()[-1]
        ver = tuple(int(part) for part in ver.split('.'))
        if ver < version:
            raise RuntimeError("Invalid version")
    except Exception:
        exit('ERROR: Git %s+ not found!' % '.'.join(map(str, version)))

def ensure_java_version(version=(1, 6), title='Java 6 runtime'):
    try:
        ver = do(
            'java', '-version', redirect_stdout=True,
            redirect_stderr=True, reterror=True
            )
        ver = ver[1].splitlines()[0].split()[-1][1:-1]
        if not ver.startswith('.'.join(map(str, version))):
            raise RuntimeError("Invalid version")
    except Exception:
        exit('ERROR: %s not found!' % title)

# ------------------------------------------------------------------------------
# Build Recipes Initialiser
# ------------------------------------------------------------------------------

def init_build_recipes():
    if RECIPES_INITIALISED:
        return
    # Try getting a lock to avoid concurrent builds.
    lock(BUILD_LOCK)
    for recipe in BUILD_RECIPES:
        execfile(recipe, BUILTINS)
    for package in list(RECIPES):
        recipes = RECIPES[package]
        versions = []
        data = {}
        for recipe in recipes:
            version = recipe['version']
            versions.append(version)
            data[version] = recipe
        RECIPES[package] = data
        PACKAGES[package] = versions
    RECIPES_INITIALISED.append(1)

# ------------------------------------------------------------------------------
# Env Manipulation
# ------------------------------------------------------------------------------

def get_ampify_env(environ):
    new = {}
    for key in environ:
        if key.startswith('AMPIFY'):
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
    return []

RESOURCE_BUILD = BASE_BUILD.copy()
RESOURCE_BUILD.update({
    'commands': resource_build_commands,
    'source': '',
    'destination': '',
    })

JAVA_BUILD = BASE_BUILD.copy()
JAVA_BUILD.update({
    'distfile': '%(name)s-%(version)s.jar'
    })

BUILD_TYPES = {
    'default': DEFAULT_BUILD,
    'java': JAVA_BUILD,
    'python': PYTHON_BUILD,
    'resources': RESOURCE_BUILD,
    }

# ------------------------------------------------------------------------------
# Core Install Functions
# ------------------------------------------------------------------------------

TO_INSTALL = {}

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

# Uninstall the given list of packages in uninstall.
def uninstall_packages(uninstall, installed):
    for name, version in uninstall.items():
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
            print "PATH:", path
            if not islink(path):
                if not exists(path):
                    continue
            if isdir(path):
                directories.add(path)
            else:
                print "Removing:", path
                os.remove(path)
        for path in reversed(sorted(directories)):
            if not listdir(path):
                print "Removing Direcotry:", path
                rmtree(path)
        receipt.close()
        os.remove(receipt_path)
        del installed[name]

# A utility function to uninstall a single package.
def uninstall_package(package):
    installed = dict(f.rsplit('-', 1) for f in listdir(RECEIPTS))
    if package in installed:
        data = {package: installed[package]}
        uninstall_packages(data, data)

# Handle the actual installation/uninstallation of appropriate packages.
def install_packages(types=BUILD_TYPES):

    ensure_gcc_version()
    ensure_git_version()
    ensure_java_version()

    for directory in [
        BUILD_WORKING_DIRECTORY, LOCAL, BIN, RECEIPTS, SHARE, TMP
        ]:
        mkdir(directory)

    # We assume the invariant that all packages only have one version installed.
    installed = dict(f.rsplit('-', 1) for f in listdir(RECEIPTS))
    uninstall = {}

    def get_installed_dependencies(package, gathered=None):
        if gathered is None:
            gathered = set()
        else:
            gathered.add(package)
        recipe = RECIPES[package][installed[package]]
        for dep in recipe.get('requires', []):
            get_installed_dependencies(dep, gathered)
        return gathered

    inverse_dependencies = {}
    for package in installed:
        if package in PACKAGES:
            for dep in get_installed_dependencies(package):
                if dep not in inverse_dependencies:
                    inverse_dependencies[dep] = set()
                inverse_dependencies[dep].add(package)

    for package in TO_INSTALL:
        if package in installed:
            existing_version = installed[package]
            if TO_INSTALL[package] != existing_version:
                uninstall[package] = existing_version
                for inv_dep in inverse_dependencies.get(package, []):
                    uninstall[package] = installed[package]

    # If a base package needs to be uninstalled, just nuke environ/local and
    # rebuild everything from scratch.
    for package in BASE_PACKAGES:
        if package in uninstall:
            rmdir(LOCAL)
            rmdir(RECEIPTS)
            unlock(BUILD_LOCK)
            execve(join(ENVIRON, 'amp'), sys.argv, get_ampify_env(environ))
    else:
        uninstall_packages(uninstall, installed)

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

    def get_listing():
        return strip_prefix(gather_local_filelisting(LOCAL), LOCAL)

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

        if info['before']:
            info['before']()

        env = environ.copy()
        if info['env']:
            env.update(info['env'])

        commands = info['commands']
        if isinstance(commands, basestring):
            commands = [commands]
        elif callable(commands):
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
                    kwargs = dict(env=env)
                    do(*command, **kwargs)
        except Exception:
            error("ERROR: Building %s %s failed" % (package, version))
            traceback.print_exc()
            sys.exit(1)
        except SystemExit:
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

# ------------------------------------------------------------------------------
# Virgin Build Handler
# ------------------------------------------------------------------------------

def build_base_and_reload():
    load_role('base')
    install_packages()
    unlock(BUILD_LOCK)
#     uninstall_package('openssl')
#     uninstall_package('bzip2')
#     uninstall_package('bsdiff')
#     uninstall_package('readline')
#     uninstall_package('zlib')
#     sys.exit(1)
    execve(join(ENVIRON, 'amp'), sys.argv, get_ampify_env(environ))
