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

top = '.'
out = 'build'

APPNAME = 'ampify'
VERSION = 'zero'
INSTALL_VERSION = 1

ROOT = os.path.realpath(os.getcwd())
LOCAL = os.path.join(ROOT, 'environ', 'local')
BIN = os.path.join(LOCAL, 'bin')
INCLUDE = os.path.join(LOCAL, 'include')
INFO = os.path.join(LOCAL, 'share', 'info')
LIB = os.path.join(LOCAL, 'lib')
RECEIPTS = os.path.join(LOCAL, 'share', 'installed')
TMP = os.path.join(LOCAL, 'tmp')
VAR = os.path.join(LOCAL, 'var')

DOWNLOAD_ROOT = "http://cloud.github.com/downloads/tav/ampify/"

JAR_FILES = {
    'closure.jar': 'closure-2010-03-24.jar',
    'yuicompressor.jar': 'yuicompressor-2.4.2.jar'
    }

JS_WRAP_START = "\n(function(){\n"
JS_WRAP_END = "\n})();\n\n"

JS_MIN = (
    "${JAVA} -jar ${AMPIFY_BIN}/%s --js ${SRC} --js_output_file ${TGT}"
    % JAR_FILES['closure.jar']
    )

AMPENV_ERROR_MESSAGE = (
    "You haven't sourced ampenv.sh! To fix, run the following in a "
    "bash shell: \n\n"
    "    $ source %s\n\n"
    "You may want to add it to your ~/.bash_profile and ~/.bashrc files.\n"
     % os.path.join(ROOT, 'environ', 'ampenv.sh')
    )

if sys.platform.startswith('freebsd'):
    make = 'gmake'
else:
    make = 'make'

# ------------------------------------------------------------------------------
# utility functions
# ------------------------------------------------------------------------------

def do(*args, **kwargs):

    from pyutil.env import run_command

    if 'redirect_stdout' not in kwargs:
        kwargs['redirect_stdout'] = False

    if 'redirect_stderr' not in kwargs:
        kwargs['redirect_stderr'] = False

    if 'retcode' in kwargs:
        return run_command(*args, **kwargs)

    kwargs['retcode'] = 1
    result = run_command(*args, **kwargs)

    retcode = result[-1]
    if retcode:
        raise ValueError("Error running %s" % ' '.join(args[0]))

    return result[:-1]

def get_target(task):
    return task.outputs[0].bldpath(task.env)

def write_dummy_target(task, target=None):
    if not target:
        target = task.outputs[0].bldpath(task.env)
    target = open(target, 'wb')
    target.write('1')
    target.close()

# ------------------------------------------------------------------------------
# distfiles support
# ------------------------------------------------------------------------------

def default_install(extra=None):
    config = ['./configure', '--prefix', LOCAL]
    if extra:
        config.extend(extra)
    env = os.environ.copy()
    env['CPPFLAGS'] = '-I%s' % INCLUDE
    env['LDFLAGS'] = '-L%s' % LIB
    do(config, env=env)
    do([make], env=env)
    do([make, 'install'], env=env)

def bdb_install():
    if os.name != 'posix':
        Logs.error(
            "Sorry, building of Berkeley DB is only supported on POSIX "
            "platforms for now."
            )
        raise NotImplementedError
    os.chdir('build_unix')
    do(['../dist/configure', '--enable-cxx', '--prefix', LOCAL])
    do([make])
    do([make, 'install'])

def nginx_install():
    join = os.path.join
    nginx = join(LOCAL, 'nginx')
    tmp = join(nginx, 'tmp')
    os.makedirs(tmp)
    config = [
        './configure', '--with-http_stub_status_module',
        '--prefix=%s' % nginx,
        '--sbin-path=%s' % join(BIN, 'nginx'),
        '--conf-path=%s' % join(nginx, 'etc', 'nginx.conf'),
        '--pid-path=%s' % join(nginx, 'var', 'nginx.pid'),
        '--error-log-path=%s' % join(nginx, 'var', 'error.log'),
        '--http-log-path=%s' % join(nginx, 'var', 'access.log'),
        '--http-client-body-temp-path=%s' % join(tmp, 'http_client'),
        '--http-proxy-temp-path=%s' % join(tmp, 'proxy'),
        '--http-fastcgi-temp-path=%s' % join(tmp, 'fastcgi'),
        '--with-cc-opt=-I%s' % INCLUDE, '--with-ld-opt=-L%s' % LIB,
        '--with-ipv6', '--with-http_ssl_module'
        ]
    do(config)
    do([make])
    do([make, 'install'])

def openssl_install():
    do(['./config', 'shared', 'no-idea', 'no-krb5', 'no-mdc2', 'no-rc5', 'zlib',
        '--prefix=%s' % LOCAL, '-L%s' % LIB, '-I%s' % INCLUDE])
    do([make])
    do([make, 'install'])

DISTFILES = {
    'db': ('4.8.26', bdb_install, []),
    'libevent': ('1.4.13', None, []),
    'python': ('2.7rc1', ['--enable-unicode=ucs2', '--enable-ipv6'], [
        'readline.install', 'openssl.install', 'zlib.install'
        ]),
    'nginx': ('0.7.65', nginx_install,
              ['openssl.install', 'pcre.install', 'zlib.install']),
    'openssl': ('0.9.8o', openssl_install, ['zlib.install']),
    'pcre': ('8.02', None, ['zlib.install']),
    'readline': ('6.1', ['--infodir=%s' % INFO], []),
    'zlib': ('1.2.5', None, [])
    }

# ------------------------------------------------------------------------------
# the core functions
# ------------------------------------------------------------------------------

def set_options(ctx):
    ctx.add_option('--zero', action='store_true', help='build the zero variant')
    ctx.add_option('--force', action='store_true', help='override base checks')

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

    if not os.environ.get('AMPIFY_ROOT'):
        ctx.fatal(AMPENV_ERROR_MESSAGE)

    ctx.check_tool('gcc')
    if not ctx.env.CC:
        ctx.fatal('C Compiler not found!')

    ctx.check_tool('gxx')
    if not ctx.env.CXX:
        ctx.fatal('C++ Compiler not found!')

    try:
        gcc_ver = do(
            [ctx.env.CC[0], '-dumpversion'], redirect_stdout=True,
            reterror=True
            )
        gcc_ver = tuple(map(int, gcc_ver[0].strip().split('.')))
        if gcc_ver < (4, 2):
            raise RuntimeError("Invalid GCC version")
    except:
        ctx.fatal('GCC 4.2+ not found!')

    ctx.check_tool('bison')
    ctx.check_tool('flex')
    ctx.check_tool('libtool')
    ctx.check_tool('python')

    ctx.find_program('coffee', var='COFFEE', mandatory=True)
    ctx.find_program('git', var='GIT', mandatory=True)
    ctx.find_program('java', var='JAVA', mandatory=True)

    try:
        java_ver = do(
            [ctx.env.JAVA, '-version'], redirect_stdout=True,
            redirect_stderr=True, reterror=True
            )
        java_ver = java_ver[1].splitlines()[0].split()[-1][1:-1]
        if not java_ver.startswith('1.6'):
            raise RuntimeError("Invalid Java version")
    except:
        ctx.fatal('Java 6 Runtime not found!')

    ctx.find_program(make, var='MAKE', mandatory=True)
    ctx.find_program('perl', var='PERL', mandatory=True)
    ctx.find_program('ruby', var='RUBY', mandatory=True)
    ctx.find_program('sass', var='SASS', mandatory=True)
    ctx.find_program('touch', var='TOUCH', mandatory=True)

    ctx.env['AMPIFY_ROOT'] = ROOT
    ctx.env['AMPIFY_BIN'] = BIN
    ctx.env['INSTALL_VERSION'] = INSTALL_VERSION
    ctx.env['ZERO_STATIC'] = join(ROOT, 'src', 'zero', 'espra', 'www')
    ctx.env['ZERO_COFFEE_OUTPUT'] = join(ROOT, 'src', 'zero', 'espra')
    ctx.env['ZERO_SASS_OUTPUT'] = join(ROOT, 'src', 'zero', 'espra', 'www')

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

    if INSTALL_VERSION != ctx.env.INSTALL_VERSION:
        if not Options.options.force:
            Logs.error("""
            The base setup has changed!!

            Please run ``make distclean`` and start again with ./configure.

            Thanks!
            """)
            sys.exit(1)

    if not os.environ.get('AMPIFY_ROOT'):
        Logs.error(AMPENV_ERROR_MESSAGE)
        sys.exit(1)

    mkdir = os.makedirs

    for directory in [LOCAL, BIN, RECEIPTS, TMP]:
        if not exists(directory):
            mkdir(directory)

    def check_submodule(task):
        target = get_target(task)
        directory = target.rsplit('.', 1)[0].split('/')[1] # @/@ unix only?
        source = os.path.join(ROOT, 'third_party', directory)
        info = os.stat(source)
        if not S_ISDIR(info[ST_MODE]):
            Logs.error("Couldn't find a directory at %s" % source)
            return
        dest = open(target, 'w')
        dest.write(str(info[ST_MTIME]))
        dest.close()

    for submodule in ['keyspace', 'nodejs', 'redis', 'pylibs']:
        ctx(target='%s.check' % submodule,
            rule=check_submodule,
            always=True,
            name='%s.check' % submodule,
            on_results=True)

    def compile_keyspace(task):
        directory = join(ROOT, 'third_party', 'keyspace')
        if not exists(join(BIN, 'keyspaced')):
            env = os.environ.copy()
            env['PREFIX'] = LOCAL
            do([make], cwd=directory, env=env)
            do([make, 'install'], cwd=directory, env=env)
            do([make, 'pythonlib'], cwd=directory, env=env)
            python = join(directory, 'bin', 'python')
            pylibs = join(ROOT, 'third_party', 'pylibs')
            for file in os.listdir(python):
                if file.endswith('.so') or file.endswith('.py'):
                    copy(join(python, file), join(pylibs, file))
        write_dummy_target(task)

    ctx(source='keyspace.check',
        target='keyspace',
        rule=compile_keyspace,
        after=['keyspace.check', 'db.install'],
        name='keyspace')

    def compile_redis(task):
        directory = join(ROOT, 'third_party', 'redis')
        if not exists(join(BIN, 'redis')):
            do([make], cwd=directory)
            copy(join(directory, 'redis-server'), join(BIN, 'redis'))
        write_dummy_target(task)

    ctx(source='redis.check',
        target='redis',
        rule=compile_redis,
        after='redis.check',
        name='redis')

    def compile_nodejs(task):
        directory = join(ROOT, 'third_party', 'nodejs')
        if not exists(join(BIN, 'node')):
            do(['./configure', '--prefix', LOCAL], cwd=directory)
            do([make, 'install'], cwd=directory)
        write_dummy_target(task)

    ctx(source='nodejs.check',
        target='node',
        rule=compile_nodejs,
        after='nodejs.check',
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
        ] + ctx.path.ant_glob('src/zero/espra/*.coffee').split()

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

    for path in ctx.path.ant_glob('src/zero/espra/*.sass').split():
        dest_path = '%s.css' % path.rsplit('.', 1)[0]
        ctx(source=path)
        ctx.install_files('${ZERO_SASS_OUTPUT}', dest_path)

    def download_file(filename, directory=BIN):

        def download(task):
            dest_path = join(directory, filename)
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

    def install_distfile(base, installer):

        import tarfile

        def install(task):
            dest_path = join(RECEIPTS, base)
            if not exists(dest_path):
                cwd = os.getcwd()
                os.chdir(TMP)
                Logs.warn("Unpacking %s.tar.gz" % base)
                tar = tarfile.open("%s.tar.gz" % base, 'r:gz')
                tar.extractall()
                tar.close()
                os.chdir(base)
                _installer = installer
                if not _installer:
                    _installer = default_install
                elif isinstance(_installer, list):
                    _installer = lambda: default_install(installer)
                _installer()
                os.chdir(cwd)
                dest = open(dest_path, 'wb')
                dest.write('1')
                dest.close()
            write_dummy_target(task)

        return install

    for distfile, (version, installer, deps) in DISTFILES.iteritems():

        base = "%s-%s" % (distfile, version)

        ctx(target="%s.tar.gz" % base,
            rule=download_file("%s.tar.gz" % base, TMP),
            name='%s.tar.gz' % distfile)

        ctx(target="%s.install" % base,
            rule=install_distfile(base, installer),
            name="%s.install" % distfile,
            after=['%s.tar.gz' % distfile] + deps)

    css_minify = (
        "${JAVA} -jar %s --charset utf-8 ${SRC} -o ${TGT}"
        % join(BIN, JAR_FILES['yuicompressor.jar'])
        )

    ctx(target='src/zero/espra/www/site.min.css',
        source='src/zero/espra/site.css',
        rule=css_minify,
        after=['yuicompressor.jar', 'sass'],
        name="css.minify")

    ctx.install_files('${ZERO_STATIC}', 'src/zero/espra/www/site.min.css')

    python_exe = join(BIN, 'python')

    def pylibs_install(task):
        if not ctx.is_install > 0:
            return
        stdout, retval = do([python_exe, 'setup.py'],
                            retcode=True,
                            cwd=join(ROOT, 'third_party', 'pylibs'))
        if retval:
            raise ValueError("The pylibs setup.py install failed.")

    ctx(source='pylibs.check',
        rule=pylibs_install,
        after=['pylibs.check', 'libevent.install', 'python.install'],
        name='pylibs.install')

    def pyutil_install(task):
        if not ctx.is_install > 0:
            return
        stdout, retval = do(
            [python_exe, 'setup.py'],
            retcode=True,
            cwd=join(ROOT, 'src'),
            redirect_stdout=True
            )
        if stdout != 'running build_ext\n':
            pass
        return retval

    ctx(rule=pyutil_install,
        name='pyutil.install',
        after=['python.install'],
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
        'src/zero/espra/ampzero.js',
        wrap_end
        ]

    concat_js(
        'ampzero.js', zerojs_segments, 'src/zero/espra/www/ampzero.js',
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
        ('src/build', join(ROOT, 'src', 'build')),
        ('third_party/pylibs/build',
         join(ROOT, 'third_party', 'pylibs', 'build'))
        ]:
        try:
            rmtree(path)
        except IOError:
            pass
        except OSError, e:
            if e.errno != ENOENT:
                Logs.warn("Couldn't remove the %s directory." % name)

    redis = join(ROOT, 'third_party', 'redis')
    do([make, 'clean'], cwd=redis)

    nodejs = join(ROOT, 'third_party', 'nodejs')
    do([make, 'distclean'], cwd=nodejs, redirect_stderr=True)

    keyspace = join(ROOT, 'third_party', 'keyspace')
    do([make, 'clean'], cwd=keyspace)

    pylibs = join(ROOT, 'third_party', 'pylibs')
    for file in os.listdir(pylibs):
        if file.startswith('keyspace') or file.endswith('.so'):
            os.remove(join(pylibs, file))

def clean(ctx):
    """remove the generated files"""

    from errno import ENOENT
    from os.path import join
    from shutil import rmtree

    for name, path in [
        # ('environ/local/tmp', TMP),
        ('environ/local/docs', join(LOCAL, 'docs')),
        ]:
        try:
            rmtree(path)
        except IOError:
            pass
        except OSError, e:
            if e.errno != ENOENT:
                Logs.warn("Couldn't remove the %s directory." % name)

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
