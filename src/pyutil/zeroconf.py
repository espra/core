# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

"""
================
ZeroConf Support
================

This module provides support functions to register and query ZeroConf records.

The ``register`` function returns either a ``1`` or a ``0`` to indicate a
successful or failed registration.

  >>> register('foo', '_test._tcp', 1234)
  1

And, similarly, the ``query`` function can be used to find registrations for a
given ``regtype``. It takes an optional ``timeout`` value (in seconds) as a
second parameter, e.g.

  >>> query('_test._tcp', 1.0)
  {u'foo._test._tcp.local.': {...'port': 1234...}}

"""

import atexit
import threading

from select import select
from time import time

try:
    import pybonjour
except Exception:
    pybonjour = None

state = threading.local()
state.announce = None
state.query = None
state.current = None
state.timeout = None

def registration_callback(sdRef, flags, errorCode, name, regtype, domain):
    if errorCode == pybonjour.kDNSServiceErr_NoError:
        state.announce = 1
    else:
        state.announce = 0

def register(name, regtype, port):
    if not pybonjour:
        return
    sdRef = pybonjour.DNSServiceRegister(
        name=name, regtype=regtype, port=port, callBack=registration_callback
        )
    try:
        while 1:
            ready = select([sdRef], [], [])
            if sdRef in ready[0]:
                pybonjour.DNSServiceProcessResult(sdRef)
                return state.announce
    finally:
        state.announce = None
        atexit.register(sdRef.close)

def resolve_callback(
    sdRef, flags, interfaceIndex, errorCode, fullname, hosttarget, port,
    txtRecord
    ):
    if errorCode == pybonjour.kDNSServiceErr_NoError:
        record = state.query[fullname] = state.current
        record['host'] = hosttarget
        record['port'] = port

def query_callback(
    sdRef, flags, interfaceIndex, errorCode, serviceName, regtype, replyDomain
    ):
    if errorCode != pybonjour.kDNSServiceErr_NoError:
        return
    if not (flags & pybonjour.kDNSServiceFlagsAdd):
        return
    if state.timeout:
        timeout = state.timeout
    else:
        timeout = None
    state.current = {
        'name': serviceName,
        'type': regtype
        }
    sdRef = pybonjour.DNSServiceResolve(
        0, interfaceIndex, serviceName, regtype, replyDomain, resolve_callback
        )
    try:
        while 1:
            ready = select([sdRef], [], [], timeout)
            if sdRef not in ready[0]:
                break
            return pybonjour.DNSServiceProcessResult(sdRef)
    finally:
        state.current = None
        sdRef.close()

def query(regtype, timeout=5.0):
    if not pybonjour:
        return {}
    sdRef = pybonjour.DNSServiceBrowse(regtype=regtype, callBack=query_callback)
    start = time()
    if timeout:
        state.timeout = timeout
    state.query = {}
    try:
        while (time() - start) <= timeout:
            ready = select([sdRef], [], [], timeout)
            if sdRef in ready[0]:
                pybonjour.DNSServiceProcessResult(sdRef)
        return state.query
    finally:
        state.query = None
        state.timeout = None
        sdRef.close()

if __name__ == '__main__':
    import doctest
    doctest.testmod(optionflags=doctest.ELLIPSIS + doctest.NORMALIZE_WHITESPACE)
