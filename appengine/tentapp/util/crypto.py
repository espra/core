# Changes released into the Public Domain by tav <tav@espians.com>

"""Miscellaneous utility crypto functions."""

from base64 import b64encode, b64decode
from hmac import HMAC
from hashlib import sha384
from time import time

from tentapp.core.config import TAMPER_PROOF_KEY, TAMPER_PROOF_DEFAULT_PERIOD

# ------------------------------------------------------------------------------
# http://rdist.root.org/2009/05/28/timing-attack-in-google-keyczar-library/
# ------------------------------------------------------------------------------

def secure_string_comparison(s1, s2, ord=ord):
    """Securely compare 2 strings in a manner which avoids timing attacks."""

    if len(s1) != len(s2):
        return False

    total = 0

    for x, y in zip(s1, s2):
        total |= ord(x) ^ ord(y)

    return total == 0

# ------------------------------------------------------------------------------
# tamper-proof value generation
# ------------------------------------------------------------------------------

def create_rehashed_mac(value, key, hmac, hasher, n=10):
    """Return a base64-encoded MAC re-hashed a few times."""

    digest = hmac(key, value, hasher).digest()

    for i in xrange(n):
        digest = hasher(digest).digest()

    return b64encode(digest, '-_')

def create_tamper_proof_string(
    value, key=TAMPER_PROOF_KEY, valid_for=TAMPER_PROOF_DEFAULT_PERIOD.seconds,
    hmac=HMAC, hasher=sha384, time=time
    ):
    """Return a tamper proof version of the passed in string value."""

    if not isinstance(value, str):
        raise ValueError("You can only tamper-proof str objects.")

    if valid_for:
        value = "%s:%s" % (int(time()) + valid_for, value)

    return "%s:%s" % (create_rehashed_mac(value, key, hmac, hasher), value)

def validate_tamper_proof_string(
    value, key=TAMPER_PROOF_KEY, timestamped=True, hmac=HMAC, hasher=sha384,
    time=time
    ):
    """Validate that the given value hasn't been tampered with."""

    try:
        mac, value = value.split(':', 1)
    except:
        return

    expected_mac = create_rehashed_mac(value, key, hmac, hasher)
    if not secure_string_comparison(mac, expected_mac):
        return

    if timestamped:
        try:
            timestamp, value = value.split(':', 1)
            timestamp = int(timestamp)
        except:
            return
        if time() > timestamp:
            return

    return value
