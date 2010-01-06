# Released into the Public Domain by tav <tav@espians.com>

"""Main site service."""

from tentapp.weblite import register_service, Raw

@register_service('site.hello', token_required=False)
def hello(ctx, name="world"):
    return Raw(u"Hello %s!" % name, 1)

@register_service('site.root_object', token_required=False)
def root(ctx):
    return Raw(u"Hello!")
