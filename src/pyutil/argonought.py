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

# def pack_integer(integer, StringIO=StringIO):
#     stream = StringIO()
#     write = stream.write
#     while 1:
#         left_bits = integer & 127
#         integer >>= 7
#         if integer:
#             left_bits += 128
#         write(chr(left_bits))
#         if not integer:
#             break
#     return stream.getvalue()

import sys

from cStringIO import StringIO

def _pack(num, write):
    lead, left = divmod(num, 254)
    print num, lead, left
    n = 0
    div = lead
    if div:
        div /= 254
        while div:
            div /= 254
            n += 1
    # print "n", n, div
    if n:
        write('\xff' * n)
        mod = lead - (254 ** n)
    print mod, lead, n
    if left:
        write(chr(mod+1))
        write(chr(left+1))
    elif mod:
        write(chr(mod+1))

#        lead_left = lead - (254 ** (n))

def _pack(num, write):
    remainder = num % 254
    lead = num - remainder
    print
    print "num:", num
    print "lead:", lead
    print "remainder:", remainder
    # lead, left = divmod(num, 254)
    if lead:
        n = 0
        while 1:
            if not (lead / (254 ** (n+1))):
                break
            n += 1
        lead_left, lead_remainder = divmod(lead, 254 ** n)
        print "n:", n
        print "lead_left:", lead_left
        if n:
            write('\xff' * n)
            write(chr(lead_left+1))
        if lead_remainder:
            while 1:
                div, lead_remainder = divmod(lead_remainder, 254)
                
    else:
        write('\x01')
        write(chr(remainder+1))
        return
    if lead_left:
        write(chr(lead_left+1))



def _pack(num, write):
    lead, left = divmod(num, 254)
    if 0:
        print
        print "num:", num
        print "lead:", lead
        print "left:", left
    if lead:
        lead_chars = []; append = lead_chars.append
        n = 0
        while 1:
            if not (lead / (254 ** (n+1))):
                break
            n += 1
        while lead:
            lead, mod = divmod(lead, 254)
            append(chr(mod+1))
        if n:
            write('\xff' * n)
        [write(char) for char in reversed(lead_chars)]
    if left:
        write('\x01')
        write(chr(left+1))

def _pack(num, write):
    lead, left = divmod(num, 254)
    if lead:
        lead_chars = []; append = lead_chars.append
        n = 0
        while 1:
            if not (lead / (254 ** (n+1))):
                break
            n += 1
        while lead:
            lead, mod = divmod(lead, 254)
            append(chr(mod+1))
        if n:
            write('\xff' * n)
        write(''.join(char for char in reversed(lead_chars)))
    if left:
        write('\x01')
        write(chr(left+1))

r = lambda res: [ord(char) for char in res]

from cStringIO import StringIO

def _pack2(num, write):
    lead, left = divmod(num, 256)
    if lead:
        lead_chars = []; append = lead_chars.append
        n = 0
        while 1:
            if not (lead / (254 ** (n+1))):
                break
            n += 1
        while lead:
            lead, mod = divmod(lead, 254)
            append(chr(mod+1))
        if n:
            write('\xff')
            n_chars = []; append = n_chars.append
            while n:
                n, mod = divmod(n, 254)
                append(chr(mod+1))
            write(''.join(char for char in reversed(n_chars)))
            write('\xff')
        write(''.join(char for char in reversed(lead_chars)))
    if left:
        write('\x01')
        write(chr(left))

def pack_number2(num, exp=0, StringIO=StringIO):
    stream = StringIO()
    write = stream.write
    if num >= 0:
        write('\x81')
    else:
        num = abs(num)
        write('\x80')
    _pack2(num, write)
    if exp:
        write('\x00')
        _pack2(exp, write)
    return stream.getvalue()


#print len(r(pack_number2(2 ** 3960)))
#print 2 ** 3960

#import sys
#sys.exit()

# ------------------------------------------------------------------------------
# 
# ------------------------------------------------------------------------------

from cStringIO import StringIO

def _pack(num, write):
    lead, left = divmod(num, 256)
    if lead:
        lead_chars = []; append = lead_chars.append
        n = 0
        while 1:
            if not (lead / (254 ** (n+1))):
                break
            n += 1
        while lead:
            lead, mod = divmod(lead, 254)
            append(chr(mod+1))
        if n:
            write('\xff')
            n_chars = []; append = n_chars.append
            while n:
                n, mod = divmod(n, 254)
                append(chr(mod+1))
            write(''.join(char for char in reversed(n_chars)))
            write('\xff')
        write(''.join(char for char in reversed(lead_chars)))
    if left:
        write('\x01')
        write(chr(left))

# 0 > 0000 0000
# 1 > 

def pack_number(num, frac=0):

    # optimise for small numbers
    if -8323072 < num < 8323072:
        if num >= 0:
            result = ['\x80', '\x00', '\x00']
            div, mod = divmod(num, 256)
            result[2] = chr(mod)
            if div:
                div, mod = divmod(div, 256)
                result[1] = chr(mod)
                if div:
                    result[0] = chr(div+128)
            return ''.join(result)
        else:
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

    if num >= 0:
        write('\x81')
    else:
        num = abs(num)
        write('\x80')
    _pack(num, write)
    if frac:
        if frac < 0:
            raise ValueError("The fractional part must be positive.")
        write('\x00')
        _pack(frac, write)
    return stream.getvalue()


import sys

prev = ''

def p(x):
    print r(pack_number(x))

#for i in range(-1000, 1000):

if 1:
    for i in range(-8323071, 8323071):
    #for i in range(-255, 1):
        cur = pack_number(i)
        #print i, '\t', r(cur)
        if cur < prev:
            print "Error!"
            sys.exit()

#p(8323071)
#p(8323072)

sys.exit()

# ------------------------------------------------------------------------------
# test
# ------------------------------------------------------------------------------

cache={}

def packtest(n, cache=cache):
    exp = 0
    if isinstance(n, str):
        _n = n.split('.', 1)
        if len(_n) == 1:
            num = int(_n[0])
        else:
            num, exp = map(int, _n)
    else:
        num = int(n)
    res = cache[n] = pack_number(num, exp)
    print "%20s\t%s" % (n, [ord(char) for char in res])

if 0:
    packtest(253)
    packtest(258)
    packtest(280381)
    packtest("280381.1")
    # sys.exit()

z = """
0 -> [1, 1]
1 -> [1, 2]
2 -> [1, 3]
...
254 -> [1, 255]
254.10 -> [1, 255, 0, 11]
255 -> [2, 1]
256 -> [2, 2]
"""

def hex_encoding(n):
    if n >= 0:
        sign = '+'
    else:
        n = abs(n)
        sign = '-'
    n = str(hex(n))
    if n.endswith('L'):
        return sign + n[2:-1]
    return sign + n[2:]

if 1:
    encoder = hex_encoding
    encoder = pack_number

    N = 100000
    prev = prev2 = ''
    longest = ''
    longest_i = 0

    for i in xrange(N):
        cur = encoder(i)
        if cur < prev:
            print "pre2:", i-2, '\t', r(prev2), '\t', repr(prev2)
            print "prev:", i-1, '\t', r(prev), '\t', repr(prev)
            print "curr:", i, '\t', r(cur), '\t', repr(cur)
            sys.exit()
        if len(cur) > len(longest):
            longest = cur
            longest_i = i
        prev2 = prev
        prev = cur

    print r(longest), longest_i, hex(longest_i)
    print r(cur), i, hex(i)

if 1:
    t = 739741763491463991364913793691363198613135193563641336134913593151951735444505930390570943708347180713087130841730830180173041376387183714139643746319545462724

    print
    print "ENC:", r(pack_number(t)), len(pack_number(t))

    print
    print filter(lambda x: x != 255, r(pack_number(t)))
    print len(filter(lambda x: x != 255, r(pack_number(t))))

    print
    print "PLAIN:", len(str(t))

    print
    print "HEX:", r(hex_encoding(t)), len(hex_encoding(t))

k = 1
print k, "\t", r(pack_number(k))

if 0:
    for i in range(100):
        n = 10 ** i
        print n, '\t', r(pack_number(n))

if 0:
    for i in xrange(N):
        packtest(i)
        if 0:
            for j in cache:
                if cache[j] > p:
                    print "Invalid:", i, j

    for i in xrange(N):
        print repr(i), '\t', [ord(char) for char in cache[i]]
