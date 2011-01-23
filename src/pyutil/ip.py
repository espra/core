# Public Domain (-) 2010-2011 The Ampify Authors.
# See the UNLICENSE file for details.

"""
===================
IPv4/IPv6 Addresses
===================

This module provides a lightweight library for manipulating IP addresses:

  >>> IPv4('123.45.67.89')
  123.45.67.89

  >>> IPv4('123.45.67.89') == IPv4('123.45.67.89')
  True

  >>> IPv4('123.45.67.89') > IPv4('123.45.67.88')
  True

  >>> IPv6('::1')
  ::1

  >>> IPv6('0:0::0:1')
  ::1

  >>> IPv6('::')
  ::

  >>> IPv6('::ffff:123.45.67.89')
  ::ffff:7b2d:4359

The ``get_teredo_ipv4_address`` method unobfuscates the public IPv4 address
embedded within Teredo IPv6 addresses:

  >>> ip = IPv6('2001::53aa:64c:0:3ffe:a3e2:b3d8')
  >>> ip.get_teredo_ipv4_address()
  92.29.76.39

The method returns ``None`` for non-Teredo addresses:

  >>> ip = IPv6('::1')
  >>> ip.get_teredo_ipv4_address()

"""

from ipaddr import IPv6Address as UtilityClass

# ------------------------------------------------------------------------------
# Helper Objects
# ------------------------------------------------------------------------------

compress_hextets = UtilityClass(0)._compress_hextets

class InvalidIPAddress(ValueError):
    pass

# ------------------------------------------------------------------------------
# IPv4 Address
# ------------------------------------------------------------------------------

class IPv4(long):

    ip_str = None
    max_ip = (2 ** 32) - 1

    def __new__(klass, ip, int_types=(int, long)):
        if isinstance(ip, int_types):
            if not 0 <= ip <= klass.max_ip:
                raise InvalidIPAddress(ip)
            return super(IPv4, klass).__new__(klass, ip)
        try:
            octets = ip.split('.')
        except Exception:
            raise InvalidIPAddress(ip)
        if len(octets) != 4:
            raise InvalidIPAddress(ip)
        ip_val = 0
        for octet in octets:
            octet = int(octet, 10)
            if not 0 <= octet <= 255:
                raise InvalidIPAddress(ip)
            ip_val = (ip_val << 8) | octet
        self = super(IPv4, klass).__new__(klass, ip_val)
        self.ip_str = ip
        return self

    def __repr__(self):
        if not self.ip_str:
            ip_val = self
            octets = []; out = octets.append
            for i in range(4):
                out(str(ip_val & 255))
                ip_val >>= 8
            self.ip_str = '.'.join(octets[::-1])
        return self.ip_str

# ------------------------------------------------------------------------------
# IPv6 Address
# ------------------------------------------------------------------------------

class IPv6(long):

    ip_str = None
    max_ip = (2 ** 128) - 1

    def __new__(klass, ip, int_types=(int, long)):

        if isinstance(ip, int_types):
            if not 0 <= ip <= klass.max_ip:
                raise InvalidIPAddress(ip)
            return super(IPv6, klass).__new__(klass, ip)

        if not isinstance(ip, basestring):
            raise InvalidIPAddress(ip)

        if ip.count('.') == 3:
            hextets = ip.split(':')
            try:
                ipv4 = IPv4(hextets.pop())
            except ValueError:
                raise InvalidIPAddress(ip)
            hextets.extend([
                '%04x' % ((ipv4 >> 16) & 65535),
                '%04x' % (ipv4 & 65535),
                ])
            ip = ':'.join(hextets)

        if ip.count('::'):
            head, tail = ip.split('::')
            ip = [
                '%04x' % (int(hextet, 16) if hextet else 0)
                for hextet in head.split(':')
                ]
            ip += ['0000'] * (8 - len(head.split(':')) - len(tail.split(':')))
            ip += [
                '%04x' % (int(hextet, 16) if hextet else 0)
                for hextet in tail.split(':')
                ]
            ip = ':'.join(ip)

        hextets = ip.split(':')
        if len(hextets) != 8:
            raise InvalidIPAddress(ip)

        ip_val = 0
        for hextet in hextets:
            hextet = int(hextet, 16)
            if not 0 <= hextet <= 65535:
                raise InvalidIPAddress(ip)
            ip_val = (ip_val << 16) | hextet

        return super(IPv6, klass).__new__(klass, ip_val)

    def __repr__(self):
        if not self.ip_str:
            hex_str = '%032x' % self
            hextets = ['%x' % int(hex_str[x:x+4], 16) for x in range(0, 32, 4)]
            self.ip_str = ':'.join(compress_hextets(hextets))
        return self.ip_str

    def get_teredo_ipv4_address(self, prefix='20010000'):
        ip = '%032x' % self
        if ip[:8] != prefix:
            return
        return IPv4(int('0x' + ip[-8:], 16) ^ 0xffffffff)


if __name__ == '__main__':
    import doctest
    doctest.testmod()
