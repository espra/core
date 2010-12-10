# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

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
