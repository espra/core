# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

import os

import Scripting
import Options

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

srcdir = '.'
blddir = 'build'

APPNAME = 'ampify'
VERSION = '0.0.1'

# ------------------------------------------------------------------------------
# the core functions
# ------------------------------------------------------------------------------

def set_options(ctx):
    ctx.add_option(
        '--all', action='store', default=False, help='Run the full test suite'
        )

def check_ampenv_setup(ctx):
    if not os.environ.get('AMPIFY_ROOT'):
        ctx.fatal(
            "You haven't sourced ampenv.sh! To fix, run the following in a "
            "bash shell: \n\n"
            "    $ source %s" %
            os.path.join(ctx.curdir, 'environ', 'startup', 'ampenv.sh')
            )

def configure(ctx):
    """configure the ampify installation"""

    check_ampenv_setup(ctx)

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
    ctx.check_tool('python')

    ctx.find_program('git', var='GIT')
    ctx.find_program('jekyll', var='JEKYLL')
    ctx.find_program('touch', var='TOUCH', mandatory=True)

    prefix = Options.options.prefix
    min_python_version = (2, 5)

def build(ctx):
    """build ampify"""

    ctx(target='latest', rule='touch ${TGT}')
