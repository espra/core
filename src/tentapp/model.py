# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

from google.appengine.ext import db

# ------------------------------------------------------------------------------
# Item Model
# ------------------------------------------------------------------------------

class I(db.Model):
    v = db.IntegerProperty(default=0)

Item = I

# ------------------------------------------------------------------------------
# User Model
# ------------------------------------------------------------------------------

class U(db.Model):
    v = db.IntegerProperty(default=0)

User = U
