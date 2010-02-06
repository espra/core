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
    result = []; write = result.append
    num -= 8323072
    lead, left = divmod(num, 256)
    n = 0
    while 1:
        if not (lead / (254 ** (n+2))):
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
        lead, mod = divmod(lead, 254)
        append(chr(mod+1))
    write(''.join(reversed(lead_chars)))
    if left:
        write('\x01')
        write(chr(left))
    return ''.join(result)

def pack_big_negative_int(num):
    result = []; write = result.append
    num = abs(num)
    num -= 8323072
    lead, left = divmod(num, 256)
    n = 0
    while 1:
        if not (lead / (254 ** (n+2))):
            break
        n += 1
    if n:
        size_chars = []; append = size_chars.append
        while n:
            n, mod = divmod(n, 254)
            append(chr(254-mod))
        write(''.join(reversed(size_chars)))
        write('\x00')
    lead_chars = []; append = lead_chars.append
    while lead:
        lead, mod = divmod(lead, 254)
        append(chr(254-mod))
    write(''.join(reversed(lead_chars)))
    write('\xfe')
    if left:
        write(chr(255-left))
    else:
        write('\xfe')
    return ''.join(result)

def pack_small_positive_int(num):
    result = ['\x80', '\x00', '\x00']
    div, mod = divmod(num, 256)
    result[2] = chr(mod)
    if div:
        div, mod = divmod(div, 256)
        result[1] = chr(mod)
        if div:
            result[0] = chr(div+128)
    return ''.join(result)

def pack_small_negative_int(num):
    num = abs(num)
    result = ['\x7f', '\xff', '\xff']
    div, mod = divmod(num, 256)
    result[2] = chr(255-mod)
    if div:
        div, mod = divmod(div, 256)
        result[1] = chr(255-mod)
        if div:
            result[0] = chr(127-div)
    return ''.join(result)

# ------------------------------------------------------------------------------
# the core number encoder
# ------------------------------------------------------------------------------

cache = {}

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
    if -8323072 < num < 8323072:
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
        write('\xff')
        write(pack_big_positive_int(num))
        if frac:
            write('\x00')
            write(frac)
    else:
        write('\x00')
        write(pack_big_negative_int(num))
        if frac:
            write('\xff')
            write(frac)
    return ''.join(result)

# ------------------------------------------------------------------------------
# more test crap
# ------------------------------------------------------------------------------

import sys

r = lambda res: [ord(char) for char in res]

def p(x):
    print r(pack_number(x))

def verify_packing(a, b, debug=True, continue_on_error=False, step=1):
    prev = ''
    for i in xrange(a, b, step):
        cur = pack_number(i)
        if debug:
            print i, '\t', r(cur)
        if cur < prev:
            print "Error!"
            if not continue_on_error:
                if not debug:
                    print i-1, '\t', r(prev)
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
verify_packing(-8323071-100000, -8323070, 0)

sys.exit()
