# No Copyright (-) 2004-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Utility tools to render structured content in git as HTML articles."""

import sys

from commands import getoutput
from datetime import datetime
from os import getcwd, sep as SEP, stat, environ, walk
from os.path import abspath, basename, dirname, join as join_path, isfile, isdir
from os.path import split as split_path, splitext, isabs
from optparse import OptionParser
from pickle import load as load_pickle, dump as dump_pickle
from re import compile

from genshi.template import MarkupTemplate, TemplateLoader
from pygments import highlight
from pygments.formatters import HtmlFormatter
from pygments.lexers import PythonLexer

from rst import render_rst

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

LINE = '-' * 78
HOME = getcwd()

MORE_LINE = '\n.. more\n'
SEPARATORS = ('-', ':')

SYNTAX_FORMATTER = HtmlFormatter(cssclass='syntax', lineseparator='<br/>')

INDEX_FILES = (
    ('index.html', 2, 'xhtml'),
    ('feed.rss', 0, 'xml'),
    ('archive.html', 1, 'xhtml'),
    )

# ------------------------------------------------------------------------------
# utility funktion
# ------------------------------------------------------------------------------

match_non_whitespace_regex = compile(r'\S').match
docstring_regex = compile(r'[r]?"""[^"\\]*(?:(?:\\.|"(?!""))[^"\\]*)*"""')

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

    if match_non_whitespace_regex(last_line):
        raise ValueError("Last line contains non-whitespace!")

    indent = len(last_line)
    return '\n'.join(line[indent:] for line in lines).strip()

def get_git_info(filename, path):

    environ['TZ'] = 'UTC'
    git_info = getoutput('git log --pretty=raw -- "%s"' % filename)

    info = {}
    info['__path__'] = path
    info['__url__'] = ''

    if (not git_info) or git_info.startswith('fatal:'):
        info['__updated__'] = datetime.utcfromtimestamp(
            stat(filename).st_mtime
            )
        return info

    info['__git__'] = 1

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

# ------------------------------------------------------------------------------
# our main skript funktion
# ------------------------------------------------------------------------------

def main(argv, genfiles=None):

    op = OptionParser()

    op.add_option('-c', dest='package', default='',
                  help="Generate documentation for the Python package (optional)")

    op.add_option('-d', dest='data_file', default='',
                  help="Set the path for a persistent data file (optional)")

    op.add_option('-e', dest='output_encoding', default='utf-8',
                  help="Set the output encoding (default: utf-8)")

    op.add_option('-f', dest='format', default='html',
                  help="Set the output format (default: html)")

    op.add_option('-i', dest='input_encoding', default='utf-8',
                  help="Set the input encoding (default: utf-8)")

    op.add_option('-o', dest='output_path', default=HOME,
                  help="Set the output directory for files (default: $PWD)")

    op.add_option('-p', dest='pattern', default='',
                  help="Generate index files for the path pattern (optional)")

    op.add_option('-r', dest='root_path', default='',
                  help="Set the path to the root working directory (optional)")

    op.add_option('-t', dest='template', default='',
                  help="Set the path to a template file (optional)")

    op.add_option('--quiet', dest='quiet', default=False, action='store_true',
                  help="Flag to suppress output")

    op.add_option('--stdout', dest='stdout', default=False, action='store_true',
                  help="Flag to redirect to stdout instead of to a file")

    try:
        options, args = op.parse_args(argv)
    except SystemExit:
        return

    output_path = options.output_path.rstrip('/')

    if not isdir(output_path):
        raise IOError("%r is not a valid directory!" % output_path)

    root_path = options.root_path

    siteinfo = join_path(output_path, '.siteinfo')
    if isfile(siteinfo):
        env = {}
        execfile(siteinfo, env)
        siteinfo = env['INFO']
    else:
        siteinfo = {
            'site_url': '',
            'site_nick': '',
            'site_description': '',
            'site_title': ''
            }

    stdout = sys.stdout if options.stdout else None
    verbose = False if stdout else (not options.quiet)

    format = options.format

    if format not in ('html', 'tex'):
        raise ValueError("Unknown format: %s" % format)

    if (format == 'tex') or (not options.template):
        template = False
    elif not isfile(options.template):
        raise IOError("%r is not a valid template!" % options.template)
    else:
        template_path = abspath(options.template)
        template_root = dirname(template_path)
        template_loader = TemplateLoader([template_root])
        template_file = open(template_path, 'rb')
        template = MarkupTemplate(
            template_file.read(), loader=template_loader, encoding='utf-8'
            )
        template_file.close()

    data_file = options.data_file

    if data_file:
        if isfile(data_file):
            data_file_obj = open(data_file, 'rb')
            data_dict = load_pickle(data_file_obj)
            data_file_obj.close()
        else:
            data_dict = {}

    input_encoding = options.input_encoding
    output_encoding = options.output_encoding

    if genfiles:

        files = genfiles

    elif options.package:

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

                info['__outdir__'] = output_path
                info['__name__'] = 'package.' + module
                info['__type__'] = 'py'
                info['__title__'] = module
                info['__source__'] = highlight(module_source, PythonLexer(), SYNTAX_FORMATTER)
                add_file((docstring, '', info))

    else:

        files = []
        add_file = files.append

        for filename in args:

            if not isfile(filename):
                raise IOError("%r doesn't seem to be a valid file!" % filename)

            if root_path and isabs(filename) and filename.startswith(root_path):
                path = filename[len(root_path)+1:]
            else:
                path = filename

            info = get_git_info(filename, path)

            source_file = open(filename, 'rb')
            source = source_file.read()
            source_file.close()

            if MORE_LINE in source:
                source_lead = source.split(MORE_LINE)[0]
                source = source.replace(MORE_LINE, '')
            else:
                source_lead = ''

            filebase, filetype = splitext(basename(filename))
            info['__outdir__'] = output_path
            info['__name__'] = filebase.lower()
            info['__type__'] = 'txt'
            info['__title__'] = filebase.replace('-', ' ')
            add_file((source, source_lead, info))

    for source, source_lead, info in files:

        if verbose:
            print
            print LINE
            print 'Converting: [%s] %s in [%s]' % (
                info['__type__'], info['__path__'], split_path(output_path)[1]
                )
            print LINE
            print

        if template:
            output, props = render_rst(
                source, format, input_encoding, True
                )
            # output = output.encode(output_encoding)
            info['__text__'] = output.encode(output_encoding)
            info.update(props)
            if source_lead:
                info['__lead__'] = render_rst(
                    source_lead, format, input_encoding, True
                    )[0].encode(output_encoding)
            output = template.generate(
                content=output,
                info=info,
                **siteinfo
                ).render('xhtml', encoding=output_encoding)
        else:
            output, props = render_rst(
                source, format, input_encoding, True, as_whole=True
                )
            info.update(props)
            output = output.encode(output_encoding)
            info['__text__'] = output
            if source_lead:
                info['__lead__'] = render_rst(
                    source_lead, format, input_encoding, True, as_whole=True
                    )[0].encode(output_encoding)

        if data_file:
            data_dict[info['__path__']] = info

        if stdout:
            print output
        else:
            output_filename = join_path(
                output_path, '%s.%s' % (info['__name__'], format)
                )
            output_file = open(output_filename, 'wb')
            output_file.write(output)
            output_file.close()
            if verbose:
                print 'Done!'

    if data_file:
        data_file_obj = open(data_file, 'wb')
        dump_pickle(data_dict, data_file_obj)
        data_file_obj.close()

    if options.pattern:

        pattern = options.pattern

        items = [
            item
            for item in data_dict.itervalues()
            if item['__outdir__'] == pattern
            ]

        # index.js/json

        import json

        index_js_template = join_path(output_path, 'index.js.template')

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
            index_js = open(join_path(output_path, 'index.js'), 'wb')
            index_js.write(index_js_template % index_json)
            index_js.close()

        for name, mode, format in INDEX_FILES:

            pname = name.split('.', 1)[0]
            template_file = None

            if siteinfo['site_nick']:
                template_path = join_path(
                    template_root, '%s.%s.genshi' % (pname, siteinfo['site_nick'])
                    )
                if isfile(template_path):
                    template_file = open(template_path, 'rb')

            if not template_file:
                template_path = join_path(template_root, '%s.genshi' % pname)

            template_file = open(template_path, 'rb')
            page_template = MarkupTemplate(
                template_file.read(), loader=template_loader, encoding='utf-8'
                )
            template_file.close()

            poutput = page_template.generate(
                items=items[:],
                root_path=output_path,
                **siteinfo
                ).render(format)

            poutput = unicode(poutput, output_encoding)

            if mode:
                output = template.generate(
                    alternative_content=poutput,
                    **siteinfo
                    ).render(format)
            else:
                output = poutput

            # @/@ wtf is this needed???
            if isinstance(output, unicode):
                output = output.encode(output_encoding)

            output_file = open(join_path(output_path, name), 'wb')
            output_file.write(output)
            output_file.close()

# ------------------------------------------------------------------------------
# run farmer!
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main(sys.argv[1:])
