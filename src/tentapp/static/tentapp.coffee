# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# Tent App Javascript Client
# ==========================

# The client has a number of default parameters, e.g.

if LOCAL_INSTANCE?
  TENT_HOST: 'http://localhost:8080'
  STATIC_HOSTS: ['http://localhost:8080']
  LIVE_HOSTS: ['http://localhost:8040']
else if SECURE_INSTANCE?
  TENT_HOST: 'https://espra.appspot.com'
  STATIC_HOSTS: [
    'https://static1.espra.appspot.com',
    'https://static2.espra.appspot.com',
    'https://static3.espra.appspot.com'
    ]
  LIVE_HOSTS: [
    'https://tentlive1.espra.com',
    'https://tentlive2.espra.com',
    'https://tentlive3.espra.com'
    ]
else
  TENT_HOST: 'http://tent.espra.com'
  STATIC_HOSTS: [
    'http://static1.espra.com',
    'http://static2.espra.com',
    'http://static3.espra.com'
    ]
  LIVE_HOSTS: [
    'http://tentlive1.espra.com',
    'http://tentlive2.espra.com',
    'http://tentlive3.espra.com'
    ]

GOOGLE_ANALYTICS_ID: 'UA-90176-30'
GOOGLE_ANALYTICS_HOST: '.espra.com'

# ------------------------------------------------------------------------------
# Builtins
# ------------------------------------------------------------------------------

if exports?
  root: exports
  log: require('sys').puts
else
  root: this
  log: ->

# ------------------------------------------------------------------------------
# Prototype Extensions
# ------------------------------------------------------------------------------

String::startswith: (prefix) ->
  @match('^' + prefix) is prefix

# ------------------------------------------------------------------------------
# Environment Support
# ------------------------------------------------------------------------------

if WebSocket?


# ------------------------------------------------------------------------------
# Google Analytics
# ------------------------------------------------------------------------------

setup_google_analytics: () ->

  # Don't load Google Analytics if it looks like this is being loaded from a
  # nodejs process, local file or localhost.
  if exports?
    return

  if document.location.protocol is 'file:'
    return

  if document.location.hostname is 'localhost'
    return

  if GOOGLE_ANALYTICS_ID?

    _gaq: []
    _gaq.push ['_setAccount', GOOGLE_ANALYTICS_ID]
    _gaq.push ['_setDomainName', GOOGLE_ANALYTICS_HOST]
    _gaq.push ['_trackPageview']

    root._gaq: _gaq

    (() ->
      ga: document.createElement 'script'
      ga.type: 'text/javascript'
      ga.async: true
      if document.location.protocol is 'https:'
        ga.src: 'https://ssl.google-analytics.com/ga.js'
      else
        ga.src: 'http://www.google-analytics.com/ga.js'
      s: document.getElementsByTagName('script')[0]
      s.parentNode.insertBefore(ga, s)
    )()

setup_google_analytics()