# Changes released into the Public Domain by tav <tav@espians.com>

"""Miscellaneous utility functions."""

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
