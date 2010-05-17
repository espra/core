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
BIN = os.path.join(LOCAL, 'bin')

DOWNLOAD_ROOT = "http://cloud.github.com/downloads/tav/ampify/"

JAR_FILES = {
    'closure.jar': 'closure-2010-03-24.jar',
    'yuicompressor.jar': 'yuicompressor-2.4.2.jar'
    }

top = '.'
out = 'build'

JS_WRAP_START = "(function(){\n"
JS_WRAP_END = "\n})();\n\n"

JS_MIN = (
    "${JAVA} -jar ${AMPIFY_BIN}/%s --js ${SRC} --js_output_file ${TGT}"
    % JAR_FILES['closure.jar']
    )

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

def get_target(task):
    return task.outputs[0].bldpath(task.env)

def write_dummy_target(task, target=None):
    if not target:
        target = task.outputs[0].bldpath(task.env)
    target = open(target, 'wb')
    target.write('1')
    target.close()

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
            os.mkdir(BIN)

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
    ctx.check_tool('python')

    ctx.find_program('coffee', var='COFFEE', mandatory=True)
    ctx.find_program('git', var='GIT', mandatory=True)
    ctx.find_program('java', var='JAVA', mandatory=True)
    ctx.find_program('ruby', var='RUBY', mandatory=True)
    ctx.find_program('sass', var='SASS', mandatory=True)
    ctx.find_program('touch', var='TOUCH', mandatory=True)

    ctx.env['AMPIFY_ROOT'] = ROOT
    ctx.env['AMPIFY_BIN'] = BIN
    ctx.env['ZERO_STATIC'] = join(ROOT, 'src', 'zero', 'espra', 'static')
    ctx.env['ZERO_COFFEE_OUTPUT'] = join(ROOT, 'src', 'zero', 'js')
    ctx.env['ZERO_SASS_OUTPUT'] = join(ROOT, 'src', 'zero', 'css')

def build(ctx):
    """build ampify"""

    if Options.options.zero:
        build_zero(ctx)
        return

def build_zero(ctx):

    from os.path import exists, join
    from shutil import copy
    from stat import S_ISDIR, ST_MODE, ST_MTIME
    from urllib import urlopen

    def check_submodule(task):
        target = get_target(task)
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
        if not exists(join(BIN, 'redis')):
            do(['make'], cwd=directory)
            copy(join(directory, 'redis-server'), join(BIN, 'redis'))
        write_dummy_target(task)

    ctx(source='check.redis',
        target='redis',
        rule=compile_redis,
        after='check.redis',
        name='redis')

    def compile_nodejs(task):
        directory = join(ROOT, 'third_party', 'nodejs')
        if not exists(join(BIN, 'node')):
            do(['./configure', '--prefix', LOCAL], cwd=directory)
            do(['make', 'install'], cwd=directory)
        write_dummy_target(task)

    ctx(source='check.nodejs',
        target='node',
        rule=compile_nodejs,
        after='check.nodejs',
        name='nodejs')

    TaskGen.declare_chain(
        name='coffeescript',
        rule='cat ${SRC} | ${COFFEE} --no-wrap --compile --stdio > ${TGT}',
        ext_in='.coffee',
        ext_out='.js',
        reentrant=False,
        after='nodejs'
        )

    coffeescript_files = [
        'third_party/coffee-script/examples/underscore.coffee'
        ] + ctx.path.ant_glob('src/zero/js/*.coffee').split()

    for path in coffeescript_files:
        dest_path = '%s.js' % path.rsplit('.', 1)[0]
        ctx(source=path)
        ctx.install_files('${ZERO_COFFEE_OUTPUT}', dest_path)

    TaskGen.declare_chain(
        name='sass',
        rule='${SASS} ${SRC} > ${TGT}',
        ext_in='.sass',
        ext_out='.css',
        reentrant=False
        )

    for path in ctx.path.ant_glob('src/zero/css/*.sass').split():
        dest_path = '%s.css' % path.rsplit('.', 1)[0]
        ctx(source=path)
        ctx.install_files('${ZERO_SASS_OUTPUT}', dest_path)

    def download_file(filename):

        def download(task):
            dest_path = join(BIN, filename)
            if not exists(dest_path):
                Logs.warn("Downloading %s" % filename)
                source = urlopen(DOWNLOAD_ROOT + filename)
                data = source.read()
                source.close()
                dest = open(dest_path, 'wb')
                dest.write(data)
                dest.close()
            write_dummy_target(task)

        return download

    for name, target in JAR_FILES.iteritems():
        ctx(target=target, rule=download_file(target), name=name)

    css_minify = (
        "${JAVA} -jar %s --charset utf-8 ${SRC} -o ${TGT}"
        % join(BIN, JAR_FILES['yuicompressor.jar'])
        )

    ctx(target='src/zero/espra/static/espra.min.css',
        source='src/zero/css/espra.css',
        rule=css_minify,
        after=['yuicompressor.jar', 'sass'],
        name="css.minify")

    ctx.install_files('${ZERO_STATIC}', 'src/zero/espra/static/espra.min.css')

    def pyutil_install(task):
        if not ctx.is_install > 0:
            return
        stdout, retval = do(
            [sys.executable, 'setup.py'],
            retcode=True,
            cwd=join(ROOT, 'src'),
            redirect_stdout=True
            )
        if stdout != 'running build_ext\n':
            pass
        return retval

    ctx(rule=pyutil_install,
        name='pyutil install',
        always=True)

    wrap_start = object()
    wrap_end = object()

    def concat_js(name, segments, target, dest, wrap=(wrap_start, wrap_end)):

        sources = [
            segment for segment in segments
            if segment not in wrap
            ]

        def concat(task):

            target = get_target(task)
            output = []; out = output.append
            wrapped = count = 0

            for segment in segments:
                if segment == wrap_start:
                    out(JS_WRAP_START)
                    wrapped = True
                elif segment == wrap_end:
                    out(JS_WRAP_END)
                    wrapped = False
                else:
                    source_path = task.inputs[count].srcpath(task.env)
                    jsfile = open(source_path, 'rb')
                    source = jsfile.readlines()
                    jsfile.close()
                    if wrapped:
                        for line in source:
                            out('  ' + line)
                    else:
                        for line in source:
                            out(line)
                    count += 1

            dest = open(target, 'wb')
            dest.write(''.join(output))
            dest.close()

        ctx(source=sources,
            rule=concat,
            target=target,
            after=['coffeescript'],
            name=name)

        ctx.install_files(dest, target)

        minified_target = '%s.min.js' % target.rsplit('.', 1)[0]

        ctx(source=target,
            target=minified_target,
            after=[name, 'closure.jar'],
            rule=JS_MIN,
            name="js.minify"
            )

        ctx.install_files(dest, minified_target)

    zerojs_segments = [
        wrap_start,
        'third_party/coffee-script/examples/underscore.js',
        wrap_end,
        wrap_start,
        'src/zero/js/naaga.js',
        'src/zero/js/espra.js',
        wrap_end
        ]

    concat_js(
        'zero.js', zerojs_segments, 'src/zero/espra/static/zero.js',
        '${ZERO_STATIC}'
        )

def docs(ctx):
    """generate ampify docs"""

    do(['yatiblog'], cwd=os.path.join(ctx.curdir, 'doc'))

def distclean(ctx):
    """remove the build and local directories"""

    from errno import ENOENT
    from os.path import exists, join
    from shutil import rmtree

    extension_modules = [
        join(ROOT, 'src', 'pyutil', 'pylzf.so')
        ]

    for ext in extension_modules:
        if exists(ext):
            os.remove(ext)

    Scripting.distclean(ctx)

    for name, path in [
        ('environ/local', LOCAL),
        ('src/build', join(ROOT, 'src', 'build'))
        ]:
        try:
            rmtree(path)
        except IOError:
            pass
        except OSError, e:
            if e.errno != ENOENT:
                Logs.warn("Couldn't remove the %s directory." % name)

    redis = join(ROOT, 'third_party', 'redis')
    do(['make', 'clean'], cwd=redis)

    nodejs = join(ROOT, 'third_party', 'nodejs')
    do(['make', 'distclean'], cwd=nodejs, redirect_stderr=True)

def clean(ctx):
    """remove the generated files"""

    Scripting.clean(ctx)

def uninstall(ctx):

    Scripting.uninstall(ctx)

# ------------------------------------------------------------------------------
# suppress some of the default waf commands
# ------------------------------------------------------------------------------

def dist(ctx):
    pass

def distcheck(ctx):
    pass

# ------------------------------------------------------------------------------
# future commands
# ------------------------------------------------------------------------------

def benchmark(ctx):
    Scripting.commands.extend(['build', 'install'])

def test(ctx):
    Scripting.commands.extend(['build', 'install'])
