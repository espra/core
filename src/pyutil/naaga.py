#! /usr/bin/env python

# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

from unicodedata import category as unicategory
from itertools import groupby
from time import time

def tokenise(text): 
    groups = groupby(text, unicategory)
    for category, token in groups: 
        yield ''.join(token)

def tokenise2(text): 
    groups = groupby(text, unicategory)
    return [''.join(token) for category, token in groups]

text = open('src/foo.js', 'rb').read()
text = unicode(text, 'utf-8')

def bench(duration, func, *args):
    print "benching <" + func.__name__ + "> for " + str(duration) + "s:"
    total = 0
    i = 0
    while 1:
        start = time()
        func(*args)
        total += time() - start
        i += 1
        if total > duration:
            break
    print total, '\t', i, 'runs'

bench(2.0, tokenise2, text)
