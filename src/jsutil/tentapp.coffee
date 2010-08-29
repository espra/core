# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# Tent App Javascript Client
# ==========================

# The client has a number of default parameters, e.g.

if LocalInstance?
    TentHost = 'http://localhost:8080'
    StaticHosts = ['http://localhost:8080']
    LiveHosts = ['http://localhost:8040']
else if SecureInstance?
    TentHost = 'https://espra.appspot.com'
    StaticHosts = [
        'https://static1.espra.appspot.com',
        'https://static2.espra.appspot.com',
        'https://static3.espra.appspot.com'
        ]
    LiveHosts = [
        'https://tentlive1.espra.com',
        'https://tentlive2.espra.com',
        'https://tentlive3.espra.com'
        ]
else
    TentHost = 'http://tent.espra.com'
    StaticHosts = [
        'http://static1.espra.com',
        'http://static2.espra.com',
        'http://static3.espra.com'
        ]
    LiveHosts = [
        'http://tentlive1.espra.com',
        'http://tentlive2.espra.com',
        'http://tentlive3.espra.com'
        ]

GoogleAnalyticsID = 'UA-90176-30'
GoogleAnalyticsHost = '.espra.com'

# ------------------------------------------------------------------------------
# Builtins
# ------------------------------------------------------------------------------

if exports?
    root = exports
    log = require('sys').puts
else
    root = this
    log = ->

# ------------------------------------------------------------------------------
# Prototype Extensions
# ------------------------------------------------------------------------------

String::startswith = (prefix) ->
    @match('^' + prefix) is prefix

# ------------------------------------------------------------------------------
# Environment Support
# ------------------------------------------------------------------------------

if WebSocket?
    initWebSockets = true

# ------------------------------------------------------------------------------
# Google Analytics
# ------------------------------------------------------------------------------

setupGoogleAnalytics = () ->

    # Don't load Google Analytics if it looks like this is being loaded from a
    # nodejs process, local file or localhost.
    if exports?
        return

    if document.location.protocol is 'file:'
        return

    if document.location.hostname is 'localhost'
        return

    if GoogleAnalyticsID?

        _gaq = []
        _gaq.push ['_setAccount', GoogleAnalyticsID]
        _gaq.push ['_setDomainName', GoogleAnalyticsHost]
        _gaq.push ['_trackPageview']

        root._gaq = _gaq

        (() ->
            ga = document.createElement 'script'
            ga.type = 'text/javascript'
            ga.async = true
            if document.location.protocol is 'https:'
                ga.src = 'https://ssl.google-analytics.com/ga.js'
            else
                ga.src = 'http://www.google-analytics.com/ga.js'
            s = document.getElementsByTagName('script')[0]
            s.parentNode.insertBefore(ga, s)
        )()

setupGoogleAnalytics()