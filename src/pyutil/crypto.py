# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

"""Miscellaneous utility crypto functions."""

from base64 import b64encode, b32encode
from hmac2 import HMAC
from hashlib import sha384
from os import urandom
from time import time
from urllib import urlencode

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

def create_encoded_mac(value, key, hmac=HMAC, hasher=sha384):
    """Return a base64-encoded MAC."""

    digest = hmac(key, value, hasher).digest()
    return b64encode(digest, '-_').rstrip('=')


def create_tamper_proof_string(
    name, value, key, duration=None, hmac=HMAC, hasher=sha384
    ):
    """Return a tamper proof version of the passed in string value."""

    if not isinstance(name, basestring):
        raise ValueError("You can only tamper-proof str name/values.")

    if not isinstance(value, basestring):
        raise ValueError("You can only tamper-proof str name/values.")

    if duration:
        value = "%s:%s" % (int(time()) + duration, value)

    named_value = "%s|%s" % (name.replace('|', r'\|'), value)

    return "%s:%s" % (
        create_encoded_mac(named_value, key, hmac, hasher),
        value
        )


def validate_tamper_proof_string(
    name, value, key, timestamped=False, hmac=HMAC, hasher=sha384
    ):
    """Validate that the given value hasn't been tampered with."""

    if not isinstance(name, basestring):
        raise ValueError("You can only tamper-proof str name/values.")

    if not isinstance(value, basestring):
        raise ValueError("You can only tamper-proof str name/values.")

    try:
        mac, value = value.split(':', 1)
    except:
        return

    named_value = "%s|%s" % (name.replace('|', r'\|'), value)

    expected_mac = create_encoded_mac(named_value, key, hmac, hasher)
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

# ------------------------------------------------------------------------------
# pseudo-signature generation and validation
# ------------------------------------------------------------------------------

def create_signature_for_payload(payload, key):
    """Return the signature for the given payload."""

    payload_keys = sorted(payload.keys())
    output = []; append = output.append

    for payload_key in payload_keys:
        append(urlencode({payload_key: payload[payload_key]}))

    output = '&'.join(output)
    return create_encoded_mac(output, key)


def sign_payload(payload, key, nonce_name='nonce', nonce_size=20):
    """Return a signature and a modified payload with a generated nonce."""
    payload[nonce_name] = b32encode(urandom(nonce_size))
    return create_signature_for_payload(payload, key), payload


def validate_signed_payload(payload, key, signature):
    """Validate that the signature matches the given payload and key."""
    expected_sig = create_signature_for_payload(payload, key)
    if secure_string_comparison(signature, expected_sig):
        return payload
