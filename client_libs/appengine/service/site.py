# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""Main site service."""

from weblite import register_service, Raw

@register_service('site.hello', token_required=False)
def hello(ctx, name="world"):
    return Raw(u"Hello %s!" % name, 1)

@register_service('site.root_object', token_required=False)
def root(ctx):
    return Raw(u"Hello!")
