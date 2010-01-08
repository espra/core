#! /usr/bin/env python2.5

# Released into the Public Domain by tav <tav@espians.com>

"""
Usage: ./remote-appengine [ dev | prod | localhost:8081 ]

"""

import code
import getpass
import os
import sys

from base64 import urlsafe_b64encode
from os.path import abspath, dirname, join as join_path

MAIN_ROOT = dirname(dirname(abspath(__file__)))
OUR_SDK_PATH = join_path(MAIN_ROOT, '.appengine_python_sdk')

sys.path.insert(0, OUR_SDK_PATH)
sys.path.insert(0, join_path(OUR_SDK_PATH, 'lib', 'yaml', 'lib'))
sys.path.insert(0, join_path(OUR_SDK_PATH, 'lib', 'webob'))
sys.path.insert(0, join_path(MAIN_ROOT, 'appengine'))

from google.appengine.api import apiproxy_stub_map
from google.appengine.tools.appengine_rpc import HttpRpcServer
from google.appengine.ext import db
from google.appengine.ext.remote_api.remote_api_stub import (
    GetSourceName, GetUserAgent, RemoteDatastoreStub, RemoteStub
    )

from ampify.core.config import REMOTE_TOKEN
from ampify.core import model
from ampify.util.crypto import create_tamper_proof_string

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

APP_ID = 'ampifyit'
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

def main(
    argv=None, app_id=APP_ID, dev_host=DEV_HOST, secure=False, key=REMOTE_TOKEN
    ):

    argv = argv or sys.argv[1:]

    if not argv:
        host = dev_host
    else:
        host = argv[0]
        if host.startswith('dev'):
            host = dev_host
        elif host.startswith('prod'):
            host = '%s.appspot.com' % app_id
            secure = True

    verifier = urlsafe_b64encode(os.urandom(24))
    mac = create_tamper_proof_string('remote', verifier, duration=None, key=key)

    path = '/.remote/%s' % mac
    os.environ['APPLICATION_ID'] = app_id

    server = NonAuthHttpRpcServer(
        host, None, GetUserAgent(), GetSourceName(), debug_data=False,
        secure=secure
        )

    apiproxy_stub_map.apiproxy = apiproxy_stub_map.APIProxyStubMap()

    datastore_stub = RemoteDatastoreStub(server, path)
    apiproxy_stub_map.apiproxy.RegisterStub('datastore_v3', datastore_stub)

    stub = RemoteStub(server, path)
    for service in SERVICES:
        apiproxy_stub_map.apiproxy.RegisterStub(service, stub)

    code.interact(
        'Ampify [App Engine] Interactive Console [%s]' % host, None,
        {'db': db, 'model': model}
        )

# ------------------------------------------------------------------------------
# self runner
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
