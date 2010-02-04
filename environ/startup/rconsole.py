#! /usr/bin/env python2.5

# No Copyright (-) 2009-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Usage: rconsole.py <app-directory> [ dev | prod | localhost:8081 ]"""

import code
import os
import sys

from base64 import urlsafe_b64encode
from os.path import dirname, isfile, join as join_path, realpath

MAIN_ROOT = dirname(dirname(dirname(realpath(__file__))))
OUR_SDK_PATH = join_path(MAIN_ROOT, '.appengine_python_sdk')

sys.path.insert(0, OUR_SDK_PATH)
sys.path.insert(0, join_path(OUR_SDK_PATH, 'lib', 'yaml', 'lib'))
sys.path.insert(0, join_path(OUR_SDK_PATH, 'lib', 'webob'))

from google.appengine.api import apiproxy_stub_map
from google.appengine.ext import db
from google.appengine.ext.remote_api.remote_api_stub import (
    GetSourceName, GetUserAgent, RemoteDatastoreStub, RemoteStub
    )

from google.appengine.tools.appengine_rpc import (
    HttpRpcServer, uses_cert_verification
    )

# ------------------------------------------------------------------------------
# die if ssl cert verification isn't enabled
# ------------------------------------------------------------------------------

if not uses_cert_verification:
    raise RuntimeError(
        "The ssl module needs to be installed for certificate verification."
        )

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

DEV_HOST = 'localhost:8080'

SERVICES = set([
    'capability_service',
    'images',
    'mail',
    'memcache',
    'taskqueue',
    'urlfetch',
    'xmpp',
])

# ------------------------------------------------------------------------------
# rpc server
# ------------------------------------------------------------------------------

class NonAuthHttpRpcServer(HttpRpcServer):

    def _DevAppServerAuthenticate(self):
        pass

# ------------------------------------------------------------------------------
# the main funktion
# ------------------------------------------------------------------------------

def setup(argv=None, ssl=False, shell=True):

    argv = argv or sys.argv[1:]

    if len(argv) < 1:
        if shell:
            print __doc__
            sys.exit(1)
        raise ValueError("You haven't passed the appropriate parameters.")

    app_root = realpath(argv.pop(0))
    if app_root not in sys.path:
        sys.path.insert(0, app_root)

    from config import REMOTE_KEY
    from pyutil.crypto import create_tamper_proof_string

    app_yaml_path = join_path(app_root, 'app.yaml')
    if not isfile(app_yaml_path):
        raise ValueError("Couldn't find: %s" % app_yaml_path)

    app_yaml_file = open(app_yaml_path, 'rb')
    app_id = None

    for line in app_yaml_file.readlines():
        if line.startswith('application:'):
            line = line.split('application:', 1)
            if len(line) == 1:
                continue
            _, app_id = line
            app_id = app_id.strip()
            break

    if not app_id:
        raise ValueError(
            "Couldn't find the Application ID in %r" % app_yaml_path
            )

    if not argv:
        host = DEV_HOST
    else:
        host = argv[0]
        if host.startswith('dev'):
            host = DEV_HOST
        elif host.startswith('prod'):
            host = '%s.appspot.com' % app_id
            ssl = True

    verifier = urlsafe_b64encode(os.urandom(24))
    mac = create_tamper_proof_string('remote', verifier, REMOTE_KEY)

    path = '/remote/%s' % mac
    os.environ['APPLICATION_ID'] = app_id

    server = NonAuthHttpRpcServer(
        host, None, GetUserAgent(), GetSourceName(), debug_data=False,
        secure=ssl
        )

    apiproxy_stub_map.apiproxy = apiproxy_stub_map.APIProxyStubMap()

    datastore_stub = RemoteDatastoreStub(server, path)
    apiproxy_stub_map.apiproxy.RegisterStub('datastore_v3', datastore_stub)

    stub = RemoteStub(server, path)
    for service in SERVICES:
        apiproxy_stub_map.apiproxy.RegisterStub(service, stub)

    if shell:
        code.interact(
            '[App Engine] Interactive Console [%s]' % host, None, {'db': db}
            )

# ------------------------------------------------------------------------------
# self runner
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    setup()
