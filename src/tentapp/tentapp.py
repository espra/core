# Public Domain (-) 2008-2011 The Ampify Authors.
# See the UNLICENSE file for details.

"""
=======
Tentapp
=======

Tentapp is a microdata-powered collaboration platform inspired by IRC and Wikis.
It is being developed in order to help with the development of Ampify and Espra,
but it will hopefully also prove useful in other contexts.

::

                      __,--'\                _                  _   
                __,--'    :. \.             ( )_               ( )_ 
           _,--'              \`.           | ,_)   __    ___  | ,_)
          /|\       `          \ `.         | |   /'__`\/' _ `\| |  
         / | \        `:        \  `/       | |_ (  ___/| ( ) || |_ 
        / '|  \        `:.       \          `\__)`\____)(_) (_)`\__)
       / , |   \                  \     
      /    |:   \              `:. \              _ _  _ _    _ _   
     /| '  |     \ :.           _,-'`.          /'_` )( '_`\ ( '_`\ 
   \' |,  / \   ` \ `:.     _,-'_|    `/       ( (_| || (_) )| (_) )
      '._;   \ .   \   `_,-'_,-'               `\__,_)| ,__/'| ,__/'
    \'    `- .\_   |\,-'_,-'                          | |    | |    
                `--|_,`'                              (_)    (_)    
                        `/              

"""

from weblite import Context, handle_http_request, main, register_service

from config import BING_MAPS_KEY

Context.bing_maps_key = BING_MAPS_KEY

# ------------------------------------------------------------------------------
# Custom Router
# ------------------------------------------------------------------------------

def router(env, args):
    prime = args[0]
    if prime.startswith('~'):
        args[0] = prime[1:]
        return 'user', args

handle_http_request.router = router

# ------------------------------------------------------------------------------
# Root Service
# ------------------------------------------------------------------------------

@register_service('root', [])
def root(ctx, *args, **kwargs):
    return '%r <br> %r' % (args, kwargs)

@register_service('boo', ['boo'])
def boo(ctx, *args, **kwargs):
    return '%r <br> %r' % (args, kwargs)

# ------------------------------------------------------------------------------
# Root Service
# ------------------------------------------------------------------------------

@register_service('user', [])
def user(ctx, *args, **kwargs):
    return 'Hello! %r' % args

# ------------------------------------------------------------------------------
# Self Runner
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
