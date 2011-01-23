# Public Domain (-) 2010-2011 The Ampify Authors.
# See the UNLICENSE file for details.

from config import SITE_ADMINS
from model import User

# ------------------------------------------------------------------------------
# Context Extensions
# ------------------------------------------------------------------------------

def get_admin_status(ctx):
    username = ctx.username
    if username and username in SITE_ADMINS:
        return 1

def get_current_user(ctx):
    username = ctx.username
    if not username:
        return
    return User.get_by_key_name(username)

def get_username(ctx):
    return ctx.get_secure_cookie('user')
