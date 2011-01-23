# Public Domain (-) 2010-2011 The Ampify Authors.
# See the UNLICENSE file for details.

"""
Argonought -- the support extension to JSON for a richer experience.

"""

from datetime import date, datetime, timedelta
from decimal import Decimal, getcontext, ROUND_DOWN
from struct import pack as struct_pack, unpack as struct_unpack

from simplejson import dumps as encode_json, loads as decode_json


__all__ = [
    'number', 'unit',
    'NUMERIC_TYPES'
    ]

serialisation_cache = {}
serialisation_map = {}
deserialisation_map = {}

# ------------------------------------------------------------------------------
# some utility funktions
# ------------------------------------------------------------------------------

def register_serialiser(argo_type, type, cache=True, cache_size=1000):
    def _register_serialiser(serialiser):
        serialisation_map[type] = (argo_type, serialiser, cache)
        if argo_type not in serialisation_cache:
            serialisation_cache[argo_type] = CachingDict(cache_size)
        return serialiser
    return _register_serialiser

def register_deserialiser(argo_type):
    def _register_deserialiser(deserialiser):
        deserialisation_map[type_string2id_map[argo_type]] = deserialiser
        return deserialiser
    return _register_deserialiser

def pack(object, stream=None, retval=False):
    if not stream:
        stream = StringIO()
        retval = True
    stream.write(serialise_object(object, SerialisationContext(stream)))
    stream.flush()
    if retval:
        return stream.getvalue()

# IPv4Address(struct.unpack('!I', data)[0])

decimal_context = getcontext()
decimal_context.prec = 40
decimal_context.rounding = ROUND_DOWN

# print decimal_context
# print Decimal(1) / Decimal(7)

decimal_context.prec = 5
decimal_context.Emax = 10000000000000000000000000000000000000000

def parse_constant(s):
    if s == 'Infinity':
        return Infinity
    if s == '-Infinity':
        return -Infinity

# print parse_constant('Infinity')
# print Decimal('901234567890123456.123459') / Decimal(1)
# print repr(decode_json('"\\"foo\\bar"'))
# print repr(decode_json('"hello"'))
# print decode_json('{"a": [nan]}', allow_nan=True)
# encode_json(allow_nan=False)
# decode_json()

# ------------------------------------------------------------------------------
# utility functions to help with the encoding of numbers
# ------------------------------------------------------------------------------

def pack_big_positive_int(num):
    result = ['\xff']; write = result.append
    num -= 8258175
    lead, left = divmod(num, 255)
    n = 1
    while 1:
        if not (lead / (253 ** n)):
            break
        n += 1
    size_chars = []; append = size_chars.append
    while n:
        n, mod = divmod(n, 254)
        append(chr(mod+1))
    write(''.join(reversed(size_chars)))
    write('\xff')
    lead_chars = []; append = lead_chars.append
    while lead:
        lead, mod = divmod(lead, 253)
        append(chr(mod+2))
    write(''.join(reversed(lead_chars)))
    if left:
        write('\x01')
        write(chr(left))
    return ''.join(result)

def pack_big_negative_int(num):
    result = ['\x00']; write = result.append
    num = abs(num)
    num -= 8258175
    lead, left = divmod(num, 254)
    n = 1
    while 1:
        if not (lead / (253 ** n)):
            break
        n += 1
    size_chars = []; append = size_chars.append
    while n:
        n, mod = divmod(n, 254)
        append(chr(254-mod))
    write('\x00')
    write(''.join(reversed(size_chars)))
    lead_chars = []; append = lead_chars.append
    while lead:
        lead, mod = divmod(lead, 253)
        append(chr(253-mod))
    if len(lead_chars) > 1:
        write('\x00')
    write(''.join(reversed(lead_chars)))
    write('\xfe')
    if left:
        write(chr(254-left))
    else:
        write('\xfe')
    return ''.join(result)

def pack_small_positive_int(num):
    result = ['\x80', '\x01', '\x01']
    div, mod = divmod(num, 255)
    result[2] = chr(mod+1)
    if div:
        div, mod = divmod(div, 255)
        result[1] = chr(mod+1)
        if div:
            result[0] = chr(div+128)
    return ''.join(result)

def pack_small_negative_int(num):
    num = abs(num)
    result = ['\x7f', '\xfe', '\xfe']
    div, mod = divmod(num, 255)
    result[2] = chr(254-mod)
    if div:
        div, mod = divmod(div, 255)
        result[1] = chr(254-mod)
        if div:
            result[0] = chr(127-div)
    return ''.join(result)

# ------------------------------------------------------------------------------
# the core number encoder
# ------------------------------------------------------------------------------

_pack_cache = {}

def pack_number(num, frac=None):
    """Encode a number according to the Argonought spec."""

    if isinstance(num, (basestring, Decimal)):
        split_num = str(num).split('.')
        if len(split_num) == 1:
            num = int(split_num[0])
        elif len(split_num) == 2:
            num = int(split_num[0])
            frac = '1' + split_num[1]
            frac_len = len(frac)
            if not frac_len == 41:
                frac = frac + ('0' * (41 - frac_len))
            frac = int(frac)
        else:
            raise ValueError("Invalid number %r" % num)
    elif frac:
        frac_check = str(frac)
        if not (frac_check.startswith('1') and len(frac_check) == 41):
            raise ValueError("Invalid fractional part %r" % frac)

    if (num, frac) in _pack_cache:
        return _pack_cache[(num, frac)]

    if num >= 0:
        positive = 1
    else:
        positive = 0

    frac_str = 0

    # calculate the encoding of the fractional part
    if frac:
        if frac < 0:
            raise ValueError("The fractional part must be positive.")
        if frac < 8323072:
            if positive:
                frac_str = pack_small_positive_int(frac)
            else:
                frac_str = pack_small_negative_int(frac)
        else:
            if positive:
                frac_str = pack_big_positive_int(frac)
            else:
                frac_str = pack_big_negative_int(frac)

    result = []; write = result.append

    # optimise for small numbers
    if -8258175 < num < 8258175:
        if positive:
            write(pack_small_positive_int(num))
            if frac_str:
                write('\x00')
                write(frac_str)
        else:
            write(pack_small_negative_int(num))
            if frac_str:
                write('\xff')
                write(frac_str)
    # deal with the big numbers
    else:
        if positive:
            write(pack_big_positive_int(num))
            if frac_str:
                write('\x00')
                write(frac_str)
        else:
            write(pack_big_negative_int(num))
            if frac_str:
                write('\xff')
                write(frac_str)

    return _pack_cache.setdefault((num, frac), ''.join(result))

# ------------------------------------------------------------------------------
# the core number decoder
# ------------------------------------------------------------------------------

_unpack_cache = {}

def _unpack_number(s):
    first = s[0]
    if first >= '\x80':
        # big +ve
        if first == '\xff':
            if s == '\xff':
                return 8258175
            s = s.rsplit('\xff', 1)[-1].split('\x01', 1)
            lead_chars = s[0]
            if len(s) == 1:
                left = 0
            else:
                left = ord(s[1])
            lead = 0
            for char in lead_chars:
                lead += (lead * 252) + (ord(char) - 2)
            num = (lead * 255) + left + 8258175
        # small +ve
        else:
            num = ((ord(s[0]) - 128) * 255) + (ord(s[1]) - 1)
            num = (num * 255) + (ord(s[2]) - 1)
    else:
        # big -ve
        if first == '\x00':
            s = s.rsplit('\x00', 1)[-1].split('\xfe', 1)
            lead_chars = s[0]
            n = lead_chars.count('\x00')
            if len(s) == 1:
                left = 0
            else:
                left = s[1]
                if left:
                    if left == '\xfe':
                        left = 0
                    else:
                        left = 254 - ord(left)
                else:
                    left = 0
            lead = 0
            for char in lead_chars:
                lead += (lead * 252) + (253 - ord(char))
            num = (lead * 254) + left + 8258175 - (254 * n)
        # small -ve
        else:
            num = ((127 - ord(s[0])) * 255) + (254 - ord(s[1]))
            num = (num * 255) + (254 - ord(s[2]))
        return -num
    return num

def unpack_number(s):
    """Decode a number according to the Argonought spec."""

    if s in _unpack_cache:
        return _unpack_cache[s]

    num = frac = 0
    ori = s

    if not s:
        raise ValueError("Cannot decode an empty string.")

    first = s[0]

    if first >= '\x80':
        split = s.split('\x00')
        if len(split) == 1:
            frac = 0
        else:
            s, frac = split
    else:
        split = s.split('\xff')
        if len(split) == 1:
            frac = 0
        else:
            s, frac = split

    num = _unpack_number(s)
    if frac:
        frac = abs(_unpack_number(frac))

    return _unpack_cache.setdefault(ori, (num, frac))

# ------------------------------------------------------------------------------
# testing
# ------------------------------------------------------------------------------

import sys

r = lambda res: [ord(char) for char in res]

def p(x):
    print r(pack_number(x))

def verify_packing(a, b, debug=True, continue_on_error=False, step=1, dec=0):
    istep = 50000 * step
    pstep = a
    try:
        prev = ''
        i = a
        while 1:
            i += step
            if i >= b:
                break
            if (i - pstep) >= istep:
                print "i:", i
                pstep = i
            cur = pack_number(i)
            if debug:
                print i, '\t', r(cur)
            if dec:
                cur_unpacked = unpack_number(cur)[0]
                if i != cur_unpacked:
                    print "Error!"
                    if not continue_on_error:
                        if not debug:
                            print i, '\t', r(cur), '\t', cur_unpacked
                        sys.exit()
            if cur < prev:
                print "Error!"
                if not continue_on_error:
                    if not debug:
                        print i-step, '\t', r(prev)
                        print i, '\t', r(cur)
                    sys.exit()
            prev = cur
    except KeyboardInterrupt:
        print "KB:", i, r(cur)
        raise

def cmp_pack_length(bits):
    n = 2 ** bits
    string = str(n)
    packed = pack_number(n)
    print "bits:", bits
    print "packed length:", len(packed)
    print "string length:", len(string)
    print "pack overhead:", len(packed) - (bits / 8)
    print
    print r(packed)

# cmp_pack_length(4096)

# verify_packing(-20000001, 20000001, 0, 0, dec=1)
# verify_packing(-(2**1024), 2 ** 1024, 0, 1, step=2**1014)
# verify_packing(-(2**4096), 2 ** 4096, 0, 0, step=2**4090)

if 0:
    i = -(8258175+82583+224)
    e = pack_number(i)
    print i
    print r(e)
    print unpack_number(e)[0]

# print r(pack_number(-1, 12))
# print r(pack_number(-1, 13))

# print pack_number(-1, 12) > pack_number(1, 1138974929462846286486282)

# print unpack_number(pack_number(-1, 12))
# print unpack_number(pack_number(1, 1138974929462846286486282))
