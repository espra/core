# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# Ampify Javascript Client
# ========================

# The client has a number of default parameters.
GoogleAnalyticsID = 'UA-90176-30'
GoogleAnalyticsHost = '.espra.com'

if LocalInstance?
  HostURL = 'https://localhost:9040/'
  StaticHosts = ['https://localhost:9040/']
  LiveHosts = ['http://localhost:8040/']
else
  HostURL = 'https://espra.com/'
  StaticHosts = [
    'https://s1.espfile.com',
    'https://s2.espfile.com',
    'https://s3.espfile.com'
    ]
  LiveHosts = [
    'https://l1.espra.com',
    'https://l2.espra.com',
    'https://l3.espra.com'
    ]

# ------------------------------------------------------------------------------
# Builtins
# ------------------------------------------------------------------------------

if exports?
  root = exports
else
  root = this

# ------------------------------------------------------------------------------
# Browser Support
# ------------------------------------------------------------------------------

# The ``validateBrowserSupport`` function checks if certain "modern" browser
# features are available and prompts the user to upgrade if not.
validateBrowserSupport = ->
  updateBrowser() if not WebSocket? or not postMessage? or not Object.defineProperty?

# The various modern HTML5 browsers that are supported. It's quite possible that
# other popular browsers like IE and Opera will also be compatible at some point
# soon, but testing is needed before adding them to this list.
supportedBrowsers = [
  ["chrome", "Chrome", "http://www.google.com/chrome"]
  ["firefox", "Firefox", "http://www.mozilla.com/en-US/firefox/all-beta.html"]
  ["safari", "Safari", "http://www.apple.com/safari/"]
]

updateBrowser = ->
  $container = $ '''
    <div class="update-browser">
      <h1>Please Upgrade to a Recent Browser</h1>
    </div>
    '''
  $browserListDiv = $ '<div class="listing"></div>'
  $browserList = $ '<ul></ul>'
  for [id, name, url] in supportedBrowsers
    $browser = $ """
      <li>
        <a href="#{url}" title="Upgrade to #{name}" class="img">
          <img src="#{HostURL}static/gfx/browser.#{id}.png" alt="#{name}" />
        </a>
        <div>
          <a href="#{url}" title="Upgrade to #{name}">
            #{name}
          </a>
        </div>
        </a>
      </li>
      """ # emacs "
    $browser.appendTo $browserList
  $browserList.appendTo $browserListDiv
  $browserListDiv.appendTo $container
  $container.appendTo 'body'
  return

# ------------------------------------------------------------------------------
# Prototype Extensions
# ------------------------------------------------------------------------------

String::startswith = (prefix) ->
    @match('^' + prefix) is prefix

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
