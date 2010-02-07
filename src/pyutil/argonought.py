"""
Argonought -- the support extension to JSON for a richer experience.

A base unit datatype for dealing with arbitrary unit values.

Note that this class doesn't currently implement fully fledged unit math.
There is no multiplication/division or conversion.


"""

from datetime import date, datetime, timedelta
from decimal import Decimal, getcontext, ROUND_DOWN

from simplejson import dumps as encode_json, loads as decode_json


__all__ = [
    'number', 'unit',
    'NUMERIC_TYPES'
    ]


decimal_context = getcontext()
decimal_context.prec = 20
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


MAX_VALUE = int('9' * 80)
DEFAULT_PRECISION = 20

# ------------------------------------------------------------------------------
# the number datatype
# ------------------------------------------------------------------------------

# http://jsfromhell.com/classes/bignumber

class Number(object):
    """A Number datatype."""

    __slots__ = ('n', 'd', 'p')

    def __init__(self, num=0, den=1, prec=DEFAULT_PRECISION):
        if not isinstance(prec, BUILTIN_INT_TYPES):
            try:
                prec = int(prec)
            except Exception:
                raise TypeError(
                    "The precision value for Number is not an integer: %r"
                    % prec
                    )
        if prec > MAX_PRECISION:
            raise ValueError(
                "Cannot have precision greater than %i decimal places."
                % MAX_PRECISION
                )
        self.p = prec
        self.n = n

    def __init__(self, n, unit):
        if not isinstance(n, NUMERIC_TYPES):
            raise TypeError("You can only create units out of numeric values.")
        self.n = n
        self.u = unit

    def _coerce(self, other, op):
        if isinstance(other, unit):
            if self.u and (self.u != other.u):
                raise TypeError(
                    'Unit mismatch for %s: %r and %r' % (op, self.u, other.u)
                    )
        else:
            try:
                other = unit(self.n.__class__(other), None)
            except Exception:
                raise TypeError(
                    'Cannot %s a "unit" object with %r' % (op, other.__class__)
                    )
        return other

    def __repr__(self):
        return u'unit(%s, %s)' % (self.n, self.u)

    def __str__(self):
        return u'%s %s' % (self.n, self.u)
 
    def __int__(self):
        return int(self.n)

    __long__ = __int__

    def __float__(self):
        return float(self.n)

    def __complex__(self):
        return complex(self.n)

    def __add__(self, other):
        other = self._coerce(other, 'add')
        return unit(self.n + other.n, self.u)
 
    __radd__ = __add__

    def __sub__(self, other): 
        other = self._coerce(other, 'subtract')
        return unit(self.n - other.n, self.u)
 
    def __rsub__(self, other): 
        other = self._coerce(other, 'subtract')
        return unit(other.n - self.n, self.u)
 
    def __pos__(self): 
        return unit(+self.n, self.u)

    def __neg__(self): 
        return unit(-self.n, self.u)

    def __abs__(self):
        return unit(abs(self.n), self.u)

    def __cmp__(self, other):
        other = self._coerce(other, 'cmp')
        return cmp(self.n, other.n)

    def __hash__(self):
        return hash((self.n, self.u))

BUILTIN_INT_TYPES = (int, long)
NUMERIC_TYPES = (int, long, float, Number)

# ------------------------------------------------------------------------------
# the unit datatype
# ------------------------------------------------------------------------------

class Unit(object):
    """A Unit datatype."""

    __slots__ = ('n', 'u')

    def __init__(self, n, unit):
        if not isinstance(n, NUMERIC_TYPES):
            raise TypeError("You can only create Units out of numeric values.")
        self.n = n
        self.u = unit

    def _coerce(self, other, op):
        if isinstance(other, Unit):
            if self.u and (self.u != other.u):
                raise TypeError(
                    'Unit mismatch for %s: %r and %r' % (op, self.u, other.u)
                    )
        else:
            try:
                other = Unit(self.n.__class__(other), None)
            except Exception:
                raise TypeError(
                    'Cannot %s a "Unit" object with %r' % (op, other.__class__)
                    )
        return other

    def __repr__(self):
        return u'unit(%s, %r)' % (self.n, self.u)

    def __str__(self):
        return u'%s %s' % (self.n, self.u)
 
    def __int__(self):
        return int(self.n)

    __long__ = __int__

    def __float__(self):
        return float(self.n)

    def __complex__(self):
        return complex(self.n)

    def __add__(self, other):
        other = self._coerce(other, 'add')
        return Unit(self.n + other.n, self.u)
 
    __radd__ = __add__

    def __sub__(self, other): 
        other = self._coerce(other, 'subtract')
        return Unit(self.n - other.n, self.u)
 
    def __rsub__(self, other): 
        other = self._coerce(other, 'subtract')
        return Unit(other.n - self.n, self.u)
 
    def __pos__(self): 
        return Unit(+self.n, self.u)

    def __neg__(self): 
        return Unit(-self.n, self.u)

    def __abs__(self):
        return Unit(abs(self.n), self.u)

    def __cmp__(self, other):
        other = self._coerce(other, 'cmp')
        return cmp(self.n, other.n)

    def __hash__(self):
        return hash((self.n, self.u))


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
    lead, left = divmod(num, 254)
    # print "plead", lead, left
    n = 0
    while 1:
        if not (lead / (253 ** (n+2))):
            break
        n += 1
    if n:
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
        write(chr(left+1))
    return ''.join(result)

def pack_big_negative_int(num):
    result = ['\x00']; write = result.append
    num = abs(num)
    num -= 8258175
    lead, left = divmod(num, 255)
    n = 0
    while 1:
        if not (lead / (253 ** (n+2))):
            break
        n += 1
    if n:
        size_chars = []; append = size_chars.append
        while n:
            n, mod = divmod(n, 253)
            append(chr(254-mod))
        write('\x00')
        write(''.join(reversed(size_chars)))
    lead_chars = []; append = lead_chars.append
    while lead:
        lead, mod = divmod(lead, 253)
        append(chr(254-mod))
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

def pack_number(num, frac=0):
    """Encode a number according to the Argonought spec."""

    # calculate the encoding of the fractional part
    if frac:
        if frac < 0:
            raise ValueError("The fractional part must be positive.")
        if frac < 8323072:
            frac = pack_small_positive_int(frac)
        else:
            frac = pack_big_positive_int(frac)

    result = []; write = result.append

    # optimise for small numbers
    if -8258175 < num < 8258175:
        if num >= 0:
            write(pack_small_positive_int(num))
            if frac:
                write('\x00')
                write(frac)
        else:
            write(pack_small_negative_int(num))
            if frac:
                write('\xff')
                write(frac)
        return ''.join(result)

    # deal with the big numbers
    if num >= 0:
        write(pack_big_positive_int(num))
        if frac:
            write('\x00')
            write(frac)
    else:
        write(pack_big_negative_int(num))
        if frac:
            write('\xff')
            write(frac)
    return ''.join(result)

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
                left = ord(s[1]) - 1
            lead = 0
            #print "ilead", r(lead_chars), left
            for char in lead_chars:
                lead += (lead * 252) + (ord(char) - 2)
            #print "ulead", lead, left
            num = (lead * 254) + left + 8258175
        # small +ve
        else:
            num = ((ord(s[0]) - 128) * 255) + (ord(s[1]) - 1)
            num = (num * 255) + (ord(s[2]) - 1)
    else:
        # big -ve
        if first == '\x00':
            num = (lead * 254) + left + 8258175
        # small -ve
        else:
            num = ((127 - ord(s[0])) * 255) + (254 - ord(s[1]))
            num = (num * 255) + (254 - ord(s[2]))
        return -num
    return num

    while lead:
        lead, mod = divmod(lead, 253)
        append(chr(mod+2))

def unpack_number(s):
    """Decode a number according to the Argonought spec."""

    num = frac = 0

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
        frac = _unpack_number(frac)

    return num, frac

# ------------------------------------------------------------------------------
# more test crap
# ------------------------------------------------------------------------------

import sys

r = lambda res: [ord(char) for char in res]

def p(x):
    print r(pack_number(x))

def verify_packing(a, b, debug=True, continue_on_error=False, step=1, dec=0):
    prev = ''
    i = a
    while 1:
        i += step
        if i >= b:
            break
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

# verify_packing(-518000, 518000, 0)
# verify_packing(-8323070-10000000, -8323071-1000000, 0)

# verify_packing(-(2**1024), 2 ** 1024, debug=1, step=2**1014)
# verify_packing(-(2**4096), 2 ** 4096, debug=0, step=2**4090)
# verify_packing(-(2**1024), 2 ** 1024, 0, 1, step=2**1014)

# p(-8323070-100000000)
# cmp_pack_length(4096)

# print pack_number(2 ** 4096).count('\x00')

# print len(pack_number(3133731337313373133731337))

if 0:
    i = -3932
    e = pack_number(i)
    print i
    print r(e)
    print unpack_number(e)


verify_packing(8258175, 8258175+82583, 0, 0, dec=1)

sys.exit()


# p(82581756)
i = 8258176023322749
print i
p = pack_number(i)
print r(p)
print unpack_number(p)[0]
# verify_packing(-8258175, 8258175, 0, 1, dec=1)
