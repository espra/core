# No Copyright (-) 2004-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Utility tools to generate HTML articles/blogs/sites from source files."""

import sys

from commands import getoutput
from datetime import datetime
from os import chdir, getcwd, environ, listdir, mkdir, remove as rm, stat, walk
from os.path import abspath, exists, join as join_path, isfile, isdir, splitext
from optparse import OptionParser
from pickle import load as load_pickle, dump as dump_pickle
from re import compile

from genshi.template import MarkupTemplate
from pygments import highlight
from pygments.formatters import HtmlFormatter
from pygments.lexers import PythonLexer
from yaml import safe_load as load_yaml

from rst import render_rst

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

LINE = '-' * 78
MORE_LINE = '\n.. more\n'
SYNTAX_FORMATTER = HtmlFormatter(cssclass='syntax', lineseparator='<br/>')

# ------------------------------------------------------------------------------
# utility funktion
# ------------------------------------------------------------------------------

docstring_regex = compile(r'[r]?"""[^"\\]*(?:(?:\\.|"(?!""))[^"\\]*)*"""')
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

def get_git_info(filename):

    environ['TZ'] = 'UTC'
    git_info = getoutput('git log --pretty=raw -- "%s"' % filename)

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

    template = MarkupTemplate(content, encoding='utf-8')

    layouts[name] = {
        '__deps__': deps,
        '__env__': env,
        '__mtime__': stat(template_path).st_mtime,
        '__name__': name,
        '__path__': template_path,
        '__template__': template,
        }

# ------------------------------------------------------------------------------
# our main skript funktion
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

    # normalise various options and load from the config file

    if args:
        source_directory = args[0]
    else:
        source_directory = getcwd()

    source_directory = abspath(source_directory)
    chdir(source_directory)

    if not isdir(source_directory):
        raise IOError("%r is not a directory!" % source_directory)

    config_file = join_path(source_directory, '_config.yml')
    if not isfile(config_file):
        raise IOError("Couldn't find: %s" % config_file)

    config_file_obj = open(config_file, 'rb')
    config_data = config_file_obj.read()
    config_file_obj.close()
    config = load_yaml(config_data)

    index_pages = config.pop('index_pages')
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

    verbose = not options.quiet

    # see if there's a persistent data file to read from

    data_file = join_path(source_directory, options.data_file)
    if isfile(data_file):
        data_file_obj = open(data_file, 'rb')
        data_dict = load_pickle(data_file_obj)
        data_file_obj.close()
    else:
        data_dict = {}

    # figure out what the generated files would be

    source_files = [
        file for file in listfiles(source_directory) if file.endswith('.txt')
        ]

    generated_files = [
        join_path(output_directory, splitext(file)[0] + '.html')
        for file in source_files
        ]

    index_files = [join_path(output_directory, index) for index in index_pages]

    # handle --clean

    if options.clean:
        for file in generated_files + index_files + [data_file]:
            if isfile(file):
                if verbose:
                    print "Removing: %s" % file
                rm(file)
        sys.exit()

    # figure out layout dependencies for the source .txt files

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
            '__name__': filebase,
            '__outdir__': output_directory,
            '__path__': source_path,
            '__rst__': True,
            '__type__': filetype
            }

    for source_file in source_files:
        init_rst_source(source_file)

    # and likewise for the index_pages

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
                '__name__': index_page,
                '__outdir__': output_directory,
                '__path__': source_path,
                '__rst__': False,
                '__type__': 'index'
                }
        else:
            init_rst_source(index_source, index_page)
        render_last.add(index_source)

    # update the envs for all the source files

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

    # figure out which files to regenerate

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

    # regenerate!

    for source, source_info in sorted(sources.items(), key=lambda x: x[1]['__rst__'] == False):

        info = config.copy()
        info.update(source_info)

        if verbose:
            print
            print LINE
            print 'Converting: [%s] %s' % (info['__type__'], info['__path__'])
            print LINE
            print

        if info['__rst__']:
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

    # persist the data file to disk

    if data_file:
        data_file_obj = open(data_file, 'wb')
        dump_pickle(data_dict, data_file_obj)
        data_file_obj.close()

    sys.exit()

    # @/@ site config

    # @/@ need to fix up this old segment of the code to the latest approach

    if options.package:

        package_root = options.package
        files = []
        add_file = files.append
        package = None
        for part in reversed(package_root.split(SEP)):
            if part:
                package = part
                break
        if package is None:
            raise ValueError("Couldn't find the package name from %r" % package_root)

        for dirpath, dirnames, filenames in walk(package_root):
            for filename in filenames:

                if not filename.endswith('.py'):
                    continue

                filename = join_path(dirpath, filename)
                module = package + filename[len(package_root):]
                if module.endswith('__init__.py'):
                    module = module[:-12]
                else:
                    module = module[:-3]

                module = '.'.join(module.split(SEP))
                module_file = open(filename, 'rb')
                module_source = module_file.read()
                module_file.close()

                docstring = docstring_regex.search(module_source)

                if docstring:
                    docstring = docstring.group(0)
                    if docstring.startswith('r'):
                        docstring = docstring[4:-3]
                    else:
                        docstring = docstring[3:-3]

                if docstring and docstring.strip().startswith('=='):
                    docstring = strip_leading_indent(docstring)
                    module_source = docstring_regex.sub('', module_source, 1)
                else:
                    docstring = ''

                info = {}

                if root_path and isabs(filename) and filename.startswith(root_path):
                    info['__path__'] = filename[len(root_path)+1:]
                else:
                    info['__path__'] = filename

                info['__updated__'] = datetime.utcfromtimestamp(
                    stat(filename).st_mtime
                    )

                info['__outdir__'] = output_directory
                info['__name__'] = 'package.' + module
                info['__type__'] = 'py'
                info['__title__'] = module
                info['__source__'] = highlight(module_source, PythonLexer(), SYNTAX_FORMATTER)
                add_file((docstring, '', info))

    # @/@ fix up the old index.js/json generator

    try:
        import json
    except ImportError:
        import simplejson as json

    index_js_template = join_path(output_directory, 'index.js.template')

    if isfile(index_js_template):

        index_json = json.dumps([
            [_art['__name__'], _art['title'].encode('utf-8')]
            for _art in sorted(
                [item for item in items if item.get('x-created') and
                 item.get('x-type', 'blog') == 'blog'],
                key=lambda i: i['x-created']
                )
            ])

        index_js_template = open(index_js_template, 'rb').read()
        index_js = open(join_path(output_directory, 'index.js'), 'wb')
        index_js.write(index_js_template % index_json)
        index_js.close()

# PDF_COMMAND = ['prince', '--input=html', '--output=pdf'] # --no-compress
# PDF_CSS = join_path(WEBSITE_ROOT, 'static', 'css', 'print.css')
# $(documentation_pdf_files): documentation/pdf/%.pdf: documentation/article/%.html $(template) $(pdf_css)
# 	@echo "---> generating" "$@"
# 	    $(prince) $$n --style=$(pdf_css) --output=$@; \

# ------------------------------------------------------------------------------
# run farmer!
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
