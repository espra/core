# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# Tent App Javascript Client
# ==========================

# The client has a number of default parameters, e.g.
ZERODATA_URL: "https://espra.appspot.com"
EVENTROUTE_URL: "https://sensor.espra.com"

GOOGLE_ANALYTICS_ID: 'UA-90176-30'
GOOGLE_ANALYTICS_HOST: '.espra.com'

root: this

# ------------------------------------------------------------------------------
# Prototype Extensions
# ------------------------------------------------------------------------------

String::startswith: (prefix) ->
  @match('^' + prefix) is prefix

# ------------------------------------------------------------------------------
# Google Analytics
# ------------------------------------------------------------------------------

setup_google_analytics: () ->

  # Don't load Google Analytics if it looks like this is being loaded from a
  # local file or localhost.
  if document.location.protocol is 'file:'
    return

  if document.location.hostname is 'localhost'
    return

  if GOOGLE_ANALYTICS_ID

    _gaq: ? []
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