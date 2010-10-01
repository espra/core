# No Copyright (-) 2004-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""
========
Yatiblog
========

Yatiblog provides a set of utility tools to generate HTML articles/blogs/sites
from source files.

"""

import atexit
import sys

from cStringIO import StringIO
from datetime import datetime
from os import chdir, getcwd, environ, listdir, mkdir, remove as rm, stat, walk
from os.path import (
    abspath, basename, dirname, exists, join as join_path, isfile, isdir,
    realpath, splitext
    )

from optparse import OptionParser
from pickle import load as load_pickle, dump as dump_pickle
from re import compile
from tokenize import generate_tokens, COMMENT, STRING, INDENT, NEWLINE, NL

from genshi.template import MarkupTemplate, NewTextTemplate as TextTemplate
from pygments import highlight
from pygments.lexers import get_lexer_by_name
from yaml import safe_load as load_yaml

from pyutil.env import run_command
from pyutil.rst import render_rst, SYNTAX_FORMATTER
from pyutil.scm import SCMConfig

# ------------------------------------------------------------------------------
# Some Constants
# ------------------------------------------------------------------------------

LINE = '-' * 78
MORE_LINE = '\n.. more\n'

# ------------------------------------------------------------------------------
# Utility Functions
# ------------------------------------------------------------------------------

docstring_regex = compile(r'[ru]?"""[^"\\]*(?:(?:\\.|"(?!""))[^"\\]*)*"""')
find_include_refs = compile(r'\.\. include::\s*(.*)\s*\n').findall
match_non_whitespace = compile(r'\S').match
match_yaml_frontmatter = compile('^---\s*\n((?:.|\n)+?)\n---\s*\n').match
replace_yaml_frontmatter = compile('^---\s*\n(?:(?:.|\n)+?)\n---\s*\n').sub

def strip_leading_indent(text):
    """
    Strip common leading indentation from a piece of ``text``.

      >>> indented_text = \"""
      ...
      ...     first indented line
      ...     second indented line
      ...
      ...     \"""

      >>> print strip_leading_indent(indented_text)
      first indented line
      second indented line

    """

    lines = text.split('\n')

    if len(lines) == 1:
        return lines[0].lstrip()

    last_line = lines.pop()

    if match_non_whitespace(last_line):
        raise ValueError("Last line contains non-whitespace!")

    indent = len(last_line)
    return '\n'.join(line[indent:] for line in lines).strip()

def replace_python_docstrings(source, non_code=(COMMENT, NEWLINE, NL)):
    """Replace Python docstrings as # comment lines."""

    comments = []; out = comments.append
    prev = code = None

    for token in generate_tokens(StringIO(source).readline):
        type = token[0]
        handle = None
        if type == STRING:
            if code:
                if prev and prev[0] == INDENT:
                    handle = 1
            else:
                handle = 1
        if type not in non_code:
            code = 1
        if handle:
            docstring = token[1]
            if docstring.endswith('"""') or docstring.endswith("'''"):
                n = 3
            elif docstring.endswith('"') or docstring.endswith("'"):
                n = 1
            if docstring.startswith('r') or docstring.startswith('u'):
                docstring = docstring[n+1:-n]
            else:
                docstring = docstring[n:-n]
            out((
                token[2], token[3], (''.join(
                '\n# ' + line
                for line in strip_leading_indent(docstring).strip().splitlines()
                ) + '\n# <yatiblog.comment>\n').strip()))
        prev = token

    source_lines = source.splitlines()
    result = []; out = result.append
    prev_row, prev_col = 1, 0

    for (start_row, start_col), (end_row, end_col), comment in comments:
        if prev_row == start_row:
            block = source_lines[prev_row-1][start_col:end_col]
        else:
            start = source_lines[prev_row-1][prev_col:]
            end = source_lines[start_row-1][:start_col]
            if prev_row == (start_row - 1):
                block = '\n'.join([start, end])
            else:
                block = '\n'.join([start] + source_lines[prev_row:start_row-1] + [end])
        out(block)
        out(comment)
        prev_row, prev_col = end_row, end_col

    out('\n')
    out('\n'.join(source_lines[prev_row:]))
    return ''.join(result)

def get_git_info(filename):
    """Extract info from the Git repository."""

    environ['TZ'] = 'UTC'
    git_info = run_command(['git', 'log', '--pretty=raw', '--', filename])

    info = {'__git__': False}

    if (not git_info) or git_info.startswith('fatal:'):
        info['__updated__'] = datetime.utcfromtimestamp(
            stat(filename).st_mtime
            )
        return info

    info['__git__'] = True

    for line in git_info.splitlines():
        if line.startswith('author'):
            email, timestamp, tz = line.split()[-3:]
            email = email.lstrip('<').rstrip('>')
            if '(' in email:
                email = email.split('(')[0].strip()
            info['__by__'] = email
            info['__updated__'] = datetime.utcfromtimestamp(float(timestamp))
            break

    return info

def listfiles(path):
    """Return a list of all non-hidden files in a directory."""
    return [
        filename
        for filename in listdir(path)
        if (not filename.startswith('.'))
        ]

def load_layout(name, path, layouts, deps=None):
    """Load the given layout template."""

    template_path = join_path(path, '_layouts', name + '.genshi')
    template_file = open(template_path, 'rb')
    content = template_file.read()
    template_file.close()

    env = {}
    front_matter = match_yaml_frontmatter(content)

    if front_matter:
        env = load_yaml(front_matter.group(1))
        layout = env.pop('layout', None)
        if layout:
            if layout not in layouts:
                load_layout(layout, path, layouts)
            deps = layouts[layout]['__deps__']
            if deps:
                deps = [layout] + deps
            else:
                deps = [layout]
        content = replace_yaml_frontmatter('', content)

    if env.get('text_template'):
        try:
            template = TextTemplate(content, encoding='utf-8')
        except Exception:
            print "Error parsing template:", name
            raise
    else:
        try:
            template = MarkupTemplate(content, encoding='utf-8')
        except Exception:
            print "Error parsing template:", name
            raise

    layouts[name] = {
        '__deps__': deps,
        '__env__': env,
        '__mtime__': stat(template_path).st_mtime,
        '__name__': name,
        '__path__': template_path,
        '__template__': template,
        }

# Define the mappings for the supported programming languages.
PROGLANGS = {
    '.coffee': ['coffeescript', '#', None],
    '.el': ['scheme', ';;', None],
    '.go': ['go', '//', None],
    '.js': ['javascript', '//', None],
    '.py': ['python', '#', replace_python_docstrings],
    '.pyx': ['cython', '#', None],
    '.rb': ['ruby', '#', None],
    '.sh': ['sh', '#', None]
    }

for lang_settings in PROGLANGS.values():
    comment_symbol = lang_settings[1]
    lang_settings.extend([
        compile(r'^\s*' + comment_symbol + r'\s?'),
        '\n' + comment_symbol + ' YATIBLOG-DIVIDER\n',
        compile('<span class="c1?">'+comment_symbol+r' YATIBLOG-DIVIDER<\/span>'),
        '\n\n.. break:: YATIBLOG-DIVIDER\n\n',
        compile('<hr class="YATIBLOG-DIVIDER" />')
        ])

del comment_symbol, lang_settings

SHEBANGS = [
    ('bash', '.sh'),
    ('node', '.js'),
    ('python', '.py'),
    ('ruby', '.rb')
    ]

# ------------------------------------------------------------------------------
# Our Main Script Function
# ------------------------------------------------------------------------------

def main(argv=None):

    argv = argv or sys.argv[1:]
    op = OptionParser(
        usage="Usage: %prog [options] [path/to/source/directory]"
        )

    op.add_option('-d', dest='data_file', default='.articlestore',
                  help="Set the path for a data file (default: .articlestore)")

    op.add_option('-o', dest='output_directory', default='website',
                  help="Set the output directory for files (default: website)")

    op.add_option('-p', dest='package', default='',
                  help="Generate documentation for a Python package (optional)")

    op.add_option('--clean', dest='clean', default=False, action='store_true',
                  help="Flag to remove all generated output files")

    op.add_option('--force', dest='force', default=False, action='store_true',
                  help="Flag to force regeneration of all files")

    op.add_option('--quiet', dest='quiet', default=False, action='store_true',
                  help="Flag to suppress output")

    try:
        options, args = op.parse_args(argv)
    except SystemExit:
        return

    # Normalise various options and load from the config file.
    if args:
        source_directory = args[0]
        source_directory_specified = True
    else:
        source_directory = getcwd()
        source_directory_specified = False

    source_directory = abspath(source_directory)
    chdir(source_directory)

    if not isdir(source_directory):
        raise IOError("%r is not a directory!" % source_directory)

    config_file = join_path(source_directory, 'yatiblog.conf')

    if isfile(config_file):
        config_file_obj = open(config_file, 'rb')
        config_data = config_file_obj.read()
        config_file_obj.close()
        config = load_yaml(config_data)
    elif not source_directory_specified:
        raise IOError("Couldn't find: %s" % config_file)
    else:
        config = {}

    index_pages = config.pop('index_pages', [])
    if not isinstance(index_pages, list):
        raise ValueError("The 'index_pages' config value is not a list!")

    index_pages = dict(
        (index_page.keys()[0], index_page.values()[0])
        for index_page in index_pages
        )

    output_directory = join_path(source_directory, options.output_directory.rstrip('/'))
    if not isdir(output_directory):
        if not exists(output_directory):
            mkdir(output_directory)
        else:
            raise IOError("%r is not a directory!" % output_directory)

    code_pages = config.pop('code_pages', {})

    if code_pages:

        code_layout = code_pages['layout']
        code_paths = code_pages['paths']
        code_files = {}

        git_root = realpath(SCMConfig().root)

        for output_filename, input_pattern in code_paths.items():

            ignore_pattern = None
            if isinstance(input_pattern, dict):
                definition = input_pattern
                input_pattern = definition['pattern']
                if 'ignore' in definition:
                    ignore_pattern = definition['ignore']

            files = run_command(['git', 'ls-files', input_pattern], cwd=git_root)
            files = filter(None, files.splitlines())

            if ignore_pattern is not None:
                ignore_files = run_command(
                    ['git', 'ls-files', ignore_pattern], cwd=git_root
                    )
                for file in ignore_files.splitlines():
                    if file in files:
                        files.remove(file)

            if '%' in output_filename:
                output_pattern = True
            else:
                output_pattern = False

            for file in files:
                directory = basename(dirname(file))
                filename, ext = splitext(basename(file))
                if output_pattern:
                    dest = output_filename % {
                        'dir':directory, 'filename':filename, 'ext':ext
                        }
                else:
                    dest = output_filename
                code_files[
                    join_path(output_directory, dest + '.html')
                    ] = [file, join_path(git_root, file)]

    else:
        code_files = {}
        code_layout = None

    verbose = not options.quiet

    # See if there's a persistent data file to read from.
    data_file = join_path(source_directory, options.data_file)
    if isfile(data_file):
        data_file_obj = open(data_file, 'rb')
        data_dict = load_pickle(data_file_obj)
        data_file_obj.close()
    else:
        data_dict = {}

    # Persist the data file to disk.
    def persist_data_file():
        if data_file:
            data_file_obj = open(data_file, 'wb')
            dump_pickle(data_dict, data_file_obj)
            data_file_obj.close()

    atexit.register(persist_data_file)

    # Figure out what the generated files would be.
    source_files = [
        file for file in listfiles(source_directory) if file.endswith('.txt')
        ]

    generated_files = [
        join_path(output_directory, splitext(file)[0] + '.html')
        for file in source_files
        ]

    index_files = [join_path(output_directory, index) for index in index_pages]

    # Handle --clean support.
    if options.clean:
        for file in generated_files + index_files + [data_file] + code_files.keys():
            if isfile(file):
                if verbose:
                    print "Removing: %s" % file
                rm(file)
        data_dict.clear()
        sys.exit()

    # Figure out layout dependencies for the source .txt files.
    layouts = {}
    sources = {}

    def init_rst_source(source_file, destname=None):

        source_path = join_path(source_directory, source_file)
        source_file_obj = open(source_path, 'rb')
        content = source_file_obj.read()
        source_file_obj.close()

        if not content.startswith('---'):
            return

        filebase, filetype = splitext(source_file)
        filebase = filebase.lower()

        env = load_yaml(match_yaml_frontmatter(content).group(1))
        layout = env.pop('layout')

        if layout not in layouts:
            load_layout(layout, source_directory, layouts)

        content = replace_yaml_frontmatter('', content)

        if MORE_LINE in content:
            lead = content.split(MORE_LINE)[0]
            content = content.replace(MORE_LINE, '')
        else:
            lead = content

        if destname:
            destname = join_path(output_directory, destname)
        else:
            destname = join_path(output_directory, filebase + '.html')

        sources[source_file] = {
            '__content__': content,
            '__deps__': find_include_refs(content),
            '__env__': env,
            '__genfile__': destname,
            '__id__': source_file,
            '__layout__': layout,
            '__lead__': lead,
            '__mtime__': stat(source_path).st_mtime,
            '__name__': basename(destname), # filebase,
            '__outdir__': output_directory,
            '__path__': source_path,
            '__rst__': True,
            '__type__': 'text',
            '__filetype__': filetype
            }

    for source_file in source_files:
        init_rst_source(source_file)

    # And likewise for any source code files.
    def init_rst_source_code(relative_source_path, source_path, destname):

        source_file_obj = open(source_path, 'rb')
        content = source_file_obj.read()
        source_file_obj.close()

        filebase, filetype = splitext(basename(source_path))
        filebase = filebase.lower()

        if not filetype:
            if content.startswith('#!'):
                content = content.split('\n', 1)
                if len(content) == 2:
                    shebang, content = content
                else:
                    shebang = content[0]
                    content = ''
                for interp, ext in SHEBANGS:
                    if interp in shebang:
                        filetype = ext
                        break
            if not filetype:
                raise ValueError("Unknown file type: %s" % source_path)

        sources[source_path] = {
            '__content__': content,
            '__deps__': [],
            '__env__': {'title': filebase},
            '__genfile__': destname,
            '__gitpath__': relative_source_path,
            '__id__': source_path,
            '__layout__': code_layout,
            '__lead__': '',
            '__mtime__': stat(source_path).st_mtime,
            '__name__': basename(destname), # filebase,
            '__outdir__': output_directory,
            '__path__': source_path,
            '__rst__': True,
            '__type__': 'code',
            '__filetype__': filetype
            }

    if code_layout and code_layout not in layouts:
        load_layout(code_layout, source_directory, layouts)

    for destname, (relative_source_path, source_path) in code_files.items():
        init_rst_source_code(relative_source_path, source_path, destname)

    # And likewise for the ``index_pages``.
    render_last = set()

    for index_page, index_source in index_pages.items():
        layout, filetype = splitext(index_source)
        if filetype == '.genshi':
            if layout not in layouts:
                load_layout(layout, source_directory, layouts)
            source_path = join_path(source_directory, '_layouts', index_source)
            sources[index_source] = {
                '__content__': '',
                '__deps__': [],
                '__env__': {},
                '__genfile__': join_path(output_directory, index_page),
                '__id__': index_source,
                '__layout__': layout,
                '__lead__': '',
                '__mtime__': stat(source_path).st_mtime,
                '__name__': basename(index_page),
                '__outdir__': output_directory,
                '__path__': source_path,
                '__rst__': False,
                '__type__': 'index',
                '__filetype__': 'genshi'
                }
        else:
            init_rst_source(index_source, index_page)
        render_last.add(index_source)

    # Update the envs for all the source files.
    for source in sources:
        info = sources[source]
        layout = info['__layout__']
        layout_info = layouts[layout]
        if layout_info['__deps__']:
            for dep_layout in reversed(layout_info['__deps__']):
                info.update(layouts[dep_layout]['__env__'])
        info.update(layouts[layout]['__env__'])
        info.update(get_git_info(info['__path__']))
        info.update(info.pop('__env__'))

    # Figure out which files to regenerate.
    if not options.force:

        no_regen = set()
        for source in sources:

            info = sources[source]
            try:
                gen_mtime = stat(info['__genfile__']).st_mtime
            except:
                continue

            dirty = False
            if gen_mtime < info['__mtime__']:
                dirty = True

            layout = info['__layout__']
            layout_info = layouts[layout]
            if layout_info['__deps__']:
                layout_chain = [layout] + layout_info['__deps__']
            else:
                layout_chain = [layout]

            for layout in layout_chain:
                if gen_mtime < layouts[layout]['__mtime__']:
                    dirty = True
                    break

            for dep in info['__deps__']:
                dep_mtime = stat(join_path(source_directory, dep)).st_mtime
                if gen_mtime < dep_mtime:
                    dirty = True
                    break

            if not dirty:
                no_regen.add(source)

        for source in no_regen:
            if source in render_last:
                continue
            del sources[source]

        remaining = set(sources.keys())
        if remaining == render_last:
            for source in remaining.intersection(no_regen):
                del sources[source]

    # Regenerate!
    items = sorted(sources.items(), key=lambda x: x[1]['__rst__'] == False)

    for source, source_info in items:

        info = config.copy()
        info.update(source_info)

        if verbose:
            print
            print LINE
            print 'Converting: [%s] %s' % (info['__type__'], info['__path__'])
            print LINE
            print

        if info['__type__'] == 'code':

            content = info['__content__']
            conf = PROGLANGS[info['__filetype__']]
            if conf[2]:
                content = conf[2](content)
            comment_matcher = conf[3]

            lines = content.split('\n')
            include_section = None

            if lines and lines[0].startswith('#!'):
                lines.pop(0)

            sections = []; new_section = sections.append
            docs_text = []; docs_out = docs_text.append
            code_text = []; code_out = code_text.append

            for line in lines:
                if comment_matcher.match(line):
                    line = comment_matcher.sub('', line)
                    if line == '<yatiblog.comment>':
                        include_section = 1
                    else:
                        docs_out(line)
                else:
                    if not line.strip():
                        if docs_text and not include_section:
                            last_line = docs_text[-1].strip()
                            if last_line:
                                last_line_char = last_line[0]
                                for char in last_line:
                                    if char != last_line_char:
                                        break
                                else:
                                    include_section = 1
                    else:
                        if docs_text:
                            include_section = 1
                    if docs_text:
                        if include_section:
                            new_section({
                                'docs_text': '\n'.join(docs_text) + '\n',
                                'code_text': '\n'.join(code_text)
                                })
                            docs_text[:] = []
                            code_text[:] = []
                            include_section = None
                        else:
                            docs_text[:] = []
                        code_out(line)
                    else:
                        code_out(line)

            new_section({'docs_text': '', 'code_text': '\n'.join(code_text)})

            docs = conf[6].join(part['docs_text'] for part in sections)
            code = conf[4].join(part['code_text'] for part in sections)

            docs_html, props = render_rst(docs, with_props=1)
            if ('title' in props) and props['title']:
                info['title'] = props['title']

            code = code.replace('\t', '    ')
            code_html = highlight(code, get_lexer_by_name(conf[0]), SYNTAX_FORMATTER)

            docs_split = conf[7].split(docs_html)
            code_split = conf[5].split(code_html)
            output = info['__output__'] = []
            out = output.append

            if docs_split and docs_split[0]:
                diff = 0
                docs_split.insert(0, u'')
            else:
                diff = 1

            last = len(docs_split) - 2
            for i in range(last + 1):
                code = code_split[i+diff].split(u'<br/>')
                while (code and code[0] == ''):
                    code.pop(0)
                while (code and code[-1] == ''):
                    code.pop()
                code = u'<br />'.join(code)
                if code:
                    if i == last:
                        code = u'<div class="syntax"><pre>' + code
                    else:
                        code = u'<div class="syntax"><pre>' + code + "</pre></div>"
                out((docs_split[i], code))

        elif info['__rst__']:
            with_props = info.get('with_props', False)
            if with_props:
                output, props = render_rst(info['__content__'], with_props=1)
                if ('title' in props) and props['title']:
                    info['title'] = props['title']
                info['__output__'] = output
            else:
                output = info['__output__'] = render_rst(info['__content__'])

            if info['__lead__'] == info['__content__']:
                info['__lead_output__'] = info['__output__']
            else:
                info['__lead_output__'] = render_rst(info['__lead__'])
        else:
            output = ''

        layout = info['__layout__']
        layout_info = layouts[layout]

        if layout_info['__deps__']:
            layout_chain = [layout] + layout_info['__deps__']
        else:
            layout_chain = [layout]

        for layout in layout_chain:
            template = layouts[layout]['__template__']
            output = template.generate(
                content=output,
                yatidb=data_dict,
                **info
                ).render('xhtml', encoding=None)

        if isinstance(output, unicode):
            output = output.encode('utf-8')

        data_dict[info['__name__']] = info

        output_file = open(info['__genfile__'], 'wb')
        output_file.write(output)
        output_file.close()

        if verbose:
            print 'Done!'

    sys.exit()

# PDF_COMMAND = ['prince', '--input=html', '--output=pdf'] # --no-compress
# PDF_CSS = join_path(WEBSITE_ROOT, 'static', 'css', 'print.css')
# $(documentation_pdf_files): documentation/pdf/%.pdf: documentation/article/%.html $(template) $(pdf_css)
# 	@echo "---> generating" "$@"
# 	    $(prince) $$n --style=$(pdf_css) --output=$@; \

# ------------------------------------------------------------------------------
# Run Farmer!
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
