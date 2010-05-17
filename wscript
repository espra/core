# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import os
import sys

import Logs
import Options
import Scripting
import TaskGen

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

APPNAME = 'ampify'
VERSION = 'zero'

ROOT = os.path.realpath(os.getcwd())
LOCAL = os.path.join(ROOT, 'environ', 'local')

top = '.'
out = 'build'

# ------------------------------------------------------------------------------
# utility functions
# ------------------------------------------------------------------------------

def do(*args, **kwargs):

    from pyutil.env import run_command

    if 'redirect_stdout' not in kwargs:
        kwargs['redirect_stdout'] = False

    if 'redirect_stderr' not in kwargs:
        kwargs['redirect_stderr'] = False

    return run_command(*args, **kwargs)

# ------------------------------------------------------------------------------
# the core functions
# ------------------------------------------------------------------------------

def set_options(ctx):
    ctx.add_option('--zero', action='store_true', help='build the zero variant')

def configure(ctx):
    """configure the ampify installation"""

    from os.path import exists, join

    py_min = (2, 5, 1)
    py_max = (3, 0)
    py_version = getattr(sys, 'version_info', None)

    if (not py_version) or not (py_min <= py_version < py_max):
        ctx.fatal(
            "You need to have Python version %s+" % '.'.join(map(str, py_min))
            )

    def _check_ampenv_setup(ctx):

        if not os.environ.get('AMPIFY_ROOT'):
            ctx.fatal(
                "You haven't sourced ampenv.sh! To fix, run the following in a "
                "bash shell: \n\n"
                "    $ source %s" % join(ROOT, 'environ', 'ampenv.sh')
                )

        if not exists(LOCAL):
            os.mkdir(LOCAL)
            os.mkdir(join(LOCAL, 'bin'))

    _check_ampenv_setup(ctx)

    ctx.check_tool('gcc')
    if not ctx.env.CC:
        ctx.fatal('C Compiler not found!')

    ctx.check_tool('gxx')
    if not ctx.env.CXX:
        ctx.fatal('C++ Compiler not found!')

    ctx.check_tool('bison')
    ctx.check_tool('flex')
    ctx.check_tool('libtool')
    ctx.check_tool('perl')
    ctx.check_tool('ruby')

    ctx.find_program('coffee', var='COFFEE')
    ctx.find_program('git', var='GIT')
    ctx.find_program('touch', var='TOUCH', mandatory=True)

    ctx.env['ZERO_COFFEE_OUTPUT'] = join(ROOT, 'src', 'zero', 'js')

def build(ctx):
    """build ampify"""

    if Options.options.zero:
        build_zero(ctx)
        return

def build_zero(ctx):

    from os.path import join
    from stat import S_ISDIR, ST_MODE, ST_MTIME

    def check_submodule(task):
        target = task.outputs[0].bldpath(task.env)
        source = os.path.join(ROOT, 'third_party', target.rsplit('.', 1)[1])
        info = os.stat(source)
        if not S_ISDIR(info[ST_MODE]):
            Logs.error("Couldn't find a directory at %s" % source)
            return
        dest = open(target, 'w')
        dest.write(str(info[ST_MTIME]))
        dest.close()

    for submodule in ['keyspace', 'nodejs', 'redis', 'pylibs']:
        ctx(target='check.%s' % submodule,
            rule=check_submodule,
            always=True,
            name='check.%s' % submodule,
            on_results=True)

    def compile_redis(task):
        directory = join(ROOT, 'third_party', 'redis')
        do(['make'], cwd=directory)

    ctx(source='check.redis',
        rule=compile_redis,
        after='check.redis',
        name='redis')

    def compile_nodejs(task):
        directory = join(ROOT, 'third_party', 'nodejs')
        do(['./configure', '--prefix', LOCAL], cwd=directory)
        do(['make', 'install'], cwd=directory)

#     ctx(source='check.nodejs',
#         rule=compile_nodejs,
#         after='check.nodejs',
#         name='node.js')

    def compile_coffeescript(task):
        pass

    TaskGen.declare_chain(
        name='coffeescript',
        rule='cat ${SRC} | ${COFFEE} --no-wrap --compile --stdio > ${TGT}',
        ext_in='.coffee',
        ext_out='.js',
        reentrant=False,
        # install='${ZERO_COFFEE_OUTPUT}',
        )

    for cs in ctx.path.ant_glob('src/zero/js/*.coffee').split():
        ctx(source=cs)
        ctx.install_files('${ZERO_COFFEE_OUTPUT}', cs)

    ctx(source=str(ctx.path.ant_glob('src/zero/js/*.coffee')).split())

    #print ctx.path.ant_glob('src/zero/js/*.coffee')
    #ctx(source=ctx.path.ant_glob('src/zero/js/*.coffee'), rule="echo Hello ${SRC}")

def docs(ctx):
    """generate ampify docs"""

    do(['yatiblog'], cwd=os.path.join(ctx.curdir, 'doc'))

def distclean(ctx):
    """remove the build and local directories"""

    from errno import ENOENT
    from shutil import rmtree

    Scripting.distclean(ctx)

    try:
        rmtree(LOCAL)
    except IOError:
        pass
    except OSError, e:
        if e.errno != ENOENT:
            Logs.warn("Couldn't remove the environ/local directory.")

    nodejs = join(ROOT, 'third_party', 'nodejs')
    do(['make', 'distclean'], cwd=nodejs, redirect_stderr=True)

def clean(ctx, as_cmd=True):
    """remove the build files"""

    from os.path import join

    if as_cmd:
        Scripting.clean(ctx)

    redis = join(ROOT, 'third_party', 'redis')
    do(['make', 'clean'], cwd=redis)

# ------------------------------------------------------------------------------
# suppress some of the default waf commands
# ------------------------------------------------------------------------------

def dist(ctx):
    pass

def distcheck(ctx):
    pass

def uninstall(ctx):
    pass

# ------------------------------------------------------------------------------
# future commands
# ------------------------------------------------------------------------------

def benchmark(ctx):
    Scripting.commands.extend(['build', 'install'])

def test(ctx):
    Scripting.commands.extend(['build', 'install'])
