# No Copyright (-) 2010 The Ampify Authors. This file is under a Public Domain
# style license that can be found in the LICENSE file.

"""
Breakpad for Python.

Importing this module and anywhere in your code will try to send a traceback to
a Breakpad Server if the app/script dies on an exception:

    from pyutil import breakpad

You can disable breakpad support in processes like test scripts by doing:

    breakpad.DISABLED = True

A Breakpad Server URL endpoint needs to be specified by the user at runtime
(i.e. when the process stops on an exception) or can be defined by your app
using:

    breakpad.SERVER_URL = 'http://your-server.com/breakpad/endpoint/path'

The endpoint should accept a POST request with the following parameters:

    command     # the name of the script/app that triggered the crash report
    args        # the command line arguments passed to the script/app
    report      # the python traceback pretty printed as HTML
    user        # the user's email address as defined in .breakpad.cfg
    nonce       # a pseudo-random number that is accompanied by a MAC

"""

import sys

from atexit import register
from base64 import b64encode
from os import urandom
from os.path import split as split_path
from urllib import urlencode
from urllib2 import urlopen

from pyutil.crypto import create_tamper_proof_string
from pyutil.env import exit
from pyutil.exception import format_exception
from pyutil.scm import SCMConfig


DISABLED = False
SERVER_URL = None


def send_report(
    report=None, url=None, user=None, key=None, prompt=True, force=False,
    quiet=False
    ):
    """Send a crash report to a Breakpad Server URL endpoint."""

    if DISABLED and not force:
        return

    if not report:
        if not hasattr(sys, 'last_traceback'):
            return
        report = format_exception(
            sys.last_type, sys.last_value, sys.last_traceback, as_html=True
            )

    if prompt:
        try:
            confirm = raw_input('Do you want to send a crash report [Y/n]? ')
        except EOFError:
            return
        if confirm.lower() in ['n', 'no']:
            return

    url = url or SERVER_URL
    if not url:
        try:
            url = raw_input('What Breakpad Server URL do you want to send to? ')
        except EOFError:
            return
        if not url:
            exit("Invalid Breakpad Server URL -- Not sending crash report.")

    if not user:
        config = SCMConfig()
        user = config.get('codereview.email')
        key = config.get('codereview.key')
        if not (user and key):
            exit("Sorry, you need to configure your codereview settings.")

    if not quiet:
        print "Sending crash report ... "

    nonce = create_tamper_proof_string(
        'nonce', b64encode(urandom(144), '-_'), key
        )

    try:
        params = {
            'args': ' '.join(sys.argv[1:]),
            'command': split_path(sys.argv[0])[1],
            'nonce': nonce,
            'report': ''.join(report),
            'user': user,
            }
        response = urlopen(url, urlencode(params))
        if not quiet:
            print response.read()
        response.close()
    except Exception:
        if not quiet:
            print "Sorry, couldn't send the crash report for some reason."


register(send_report)
