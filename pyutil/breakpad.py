# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""
Breakpad for Python.

Importing this module and anywhere in your code will try to send a traceback to
a Breakpad Server if the app/script dies on an exception:

    from pyutil import breakpad

You can disable breakpad support in processes like test scripts by doing:

    breakpad.DISABLED = True

A Breakpad Server URL endpoint needs to be specified when send_report is called
manually or defined in the SCM's codereview settings:

    codereview.breakpad_url  # the breakpad server url endpoint
    codereview.email         # the user's login email address
    codereview.key           # a shared secret with the server for creating MACs

The endpoint should accept a POST request with the following parameters:

    command     # the name of the script/app that triggered the crash report
    args        # the command line arguments passed to the script/app
    report      # the python traceback pretty printed as HTML
    user        # the user's login email address
    nonce       # a pseudo-random number that is accompanied by a MAC

"""

import sys

from atexit import register
from os.path import split as split_path
from urllib import urlencode
from urllib2 import urlopen

from pyutil.crypto import sign_payload
from pyutil.env import exit
from pyutil.exception import format_exception
from pyutil.scm import SCMConfig


DISABLED = False


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

    if not user:
        config = SCMConfig()
        user = config.get('codereview.email')
        key = config.get('codereview.key')
        url = config.get('codereview.breakpad_url')
        if not (user and key and url):
            exit("Sorry, you need to configure your codereview settings.")

    if not quiet:
        print
        print "Sending crash report ... "

    payload = {
        'args': ' '.join(sys.argv[1:]),
        'command': split_path(sys.argv[0])[1],
        'report': ''.join(report),
        'user': user,
        }

    payload['sig'], payload = sign_payload(payload, key)

    try:
        response = urlopen(url, urlencode(payload))
        if not quiet:
            print
            print response.read()
        response.close()
    except Exception:
        if not quiet:
            print
            print "Sorry, couldn't send the crash report for some reason."


register(send_report)
