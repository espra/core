# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# Ampify Javascript Client
# ========================

# The client has a number of default parameters, e.g.
HOST_URL = "https://espra.com/"
EVENTROUTE_URL = "https://sensor.espra.com"

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
          <img src="#{HOST_URL}static/gfx/browser.#{id}.png" alt="#{name}" />
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
  return true
