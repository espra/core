<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<!--

site_author: Ampify Authors
site_description:
    Ampify is a vision of a better future. Help it happen with £10/month. Open
    Source + Reputation Economy = Web 4.0

site_license: This work has been placed into the Public Domain.
site_title: "Support Ampify: Create Weapons of Mass Construction!"
site_url: http://ampify.it

analytics_id: UA-90176-24
section_id: about

index_pages:
- bank-transfer.html: bank-transfer.genshi
- community.html: community.genshi
- index.html: index.genshi

-->
<head>
  <meta content="text/html; charset=utf-8" http-equiv="content-type" />
  <title>Espra Tent &#187; (page_title)</title>
  <meta http-equiv="imagetoolbar" content="no" />
  <meta name="MSSmartTagsPreventParsing" content="true" />
  <meta name="robots" content="index, follow" />
  <meta name="revisit-after" content="1 day" />
  <meta name="document-rating" content="general" />
  <link rel="icon" type="image/png" href="/favicon.ico" />
  <link rel="alternate" type="application/rss+xml" 
		title="RSS Feed for {site_title}"
		href="http://feeds2.feedburner.com/{site_nick}" />
  <link rel="stylesheet" type="text/css" media="screen" title="default"
		href="${STATIC('css/screen.css', True)}" />
  <style type="text/css" media="print">
    /* @import url("${STATIC('css/screen.css', True)}"); */
    #ignore-this { display: none; }
  </style>
  <!--[if lte IE 8]>
    <style type="text/css">
      ol { list-style-type: disc; }
    </style>
  <![endif]-->
  <!--[if lte IE 7]>
    <style type="text/css">
      #body-wrapper { height: 100%; }
    </style>
  <![endif]-->
  <script type="text/javascript" src="${ctx.req_scheme}://ajax.googleapis.com/ajax/libs/jquery/1.3.2/jquery.min.js"></script>
  <script type="text/javascript" src="http://eu.live.app.com/info.js"></script>
  <script type="text/javascript" src="http://usa.live.app.com/info.js"></script>
  <script type="text/javascript">
  var closest_cluster = null;
  $.getJSON('http://eu.live.app.com/ping.js?__callback__=?', function (data) { if (!closest_cluster) closest_cluster = data.cluster_id });
  $.getJSON('http://usa.live.app.com/ping.js?__callback__=?', function (data) { if (!closest_cluster) closest_cluster = data.cluster_id });
  $(function () {
    $('#info').html
  });
  </script>
  <script type="text/javascript" src="${STATIC('js/plugins.js', True)}"></script>
  <script type="text/javascript">
  // <![CDATA[

$(function () {
  $letterbox = $('#letterbox');
  $letterbox.autogrow({max_height: 300});

  $tools = $('#tools');
  $upload = $('#upload');

  $('.btn-thing').css({
    'padding' : '5px 10px 4px 10px',
    'font-size' : '12px'
  });

  $('#btn-preview').styledButton({
    'orientation': 'left'
  });

  $('#btn-upload').styledButton({
    'orientation': 'alone',
    'toggle': true,
    'action': {
      on: function () { $upload.show(); },
      off: function () { $upload.hide(); }
    }
  });

  $('#btn-tools').styledButton({
    'orientation': 'alone',
    'toggle': true,
    'action': {
      on: function () { $tools.show(); },
      off: function () { $tools.hide(); }
    }
  });

  $('#btn-draft').styledButton({
    'orientation': 'right'
  });

  $('#btn-send').styledButton({
    'orientation': 'alone'
  });

  $('.btn-thing-from').css({
    'padding' : '5px 10px 4px 10px',
    'font-size' : '12px'
  }).styledButton({
    'orientation': 'alone',
    'toggle': true
  });

var dock_visible = true;

$('#dock-toggler').click(function () {
  if (dock_visible) {
    $('#dock-inner').hide();
    dock_visible = false;
  } else {
    $('#dock-inner').show();
    dock_visible = true;
  }
});

    var timeout, keep_checking;

    function start_checking () {
      if (keep_checking)
        return;
	  keep_checking = true;
      update_text_length();
    };

    $letterbox_size = $('#letterbox-size');

    function update_text_length () {
      if (timeout) {
        window.clearTimeout(timeout);
        timeout = null;
      }
      $letterbox_size.text(get_string_length($letterbox.val()));
      if (keep_checking)
        timeout = window.setTimeout(update_text_length, 150);
    };

    function stop_checking () {
      keep_checking = false;
      window.clearTimeout(timeout);
    };

    $letterbox.focus(start_checking).blur(stop_checking);

  $tools.hide();
  $upload.hide();
  $letterbox.focus();

  $('.msg').click(function () {
    $('.msg').removeClass('msg-selected');
    $(this).addClass('msg-selected');
  });

/*

*/

});

http://www.inter-locale.com/demos/countBytes.html

  function get_string_length(text) {
        var escapedStr = encodeURI(text)
        if (escapedStr.indexOf("%") != -1) {
            var count = escapedStr.split("%").length - 1
            if (count == 0) count++  //perverse case; can't happen with real UTF-8
            var tmp = escapedStr.length - (count * 3)
            count = count + tmp
        } else {
            count = escapedStr.length
        }
        return count;
     }

/*
$(function () {
  //$ueberlein = $('#ueberlein');
  //$ueberlein.autogrow({max_height: 300});
  //$ueberlein.focus();
  $letterbox = $('#letterbox');
  $letterbox.autogrow({max_height: 300});
  $letterbox.focus();
  $top = $('#top');
  $bottom = $('#bottom');
  $window = $(window);
  var window_height = $window.height();
  var dolumn_height = parseInt((window_height - 120) / 2);
  var new_css = {height: dolumn_height};
  $top.css(new_css);
  $bottom.css(new_css);
  $bottom.html($window.width());
});
*/

  // ]]>
  </script>
</head>
<body>

<!--
<div id="topbar" style="margin-left: 10px; margin-right: 10px;">
  <form action="/.search" method="get" id="topbar-search-form" name="topbar-search-form">
  <div id="topbar-items">
    <div id="topbar-search-segment">
      <input id="topbar-search" type="text" name="q" autosave="com.espra.www" placeholder="Enter your query" results="10" /><a href="" id="topbar-search-button"></a>
    </div>
  % if ctx.is_logged_in():
    <a href="/.logout" id="topbar-logout-link">logout</a>
  % else:
    <a href="/.login" id="topbar-login-link">login</a>
  % endif
&nbsp;
<a href="" id="topbar-login-link">settings</a>
&nbsp;
<a href="" id="topbar-login-link">write</a>
&nbsp;
<a href="" id="topbar-login-link">faves</a>
&nbsp;
<a href="" id="topbar-login-link">help</a>
  </div>
  </form>
  <a href="/" title="Espra"><img src="${STATIC('gfx/logo.espra.white.png')}" alt="Espra" width="109px" height="32px" /></a>
</div>
-->

<div id="dock">
<div id="dock-inner">
<a href="/" title="Espra Tent" class="dock-link dock-home-link"><img src="${STATIC('gfx/logo.espra.tent.png')}" width="113px" height="20px" alt="Espra Tent" /></a>
<a class="dock-link" href="/.login" id="login-link"></a>
<a href="" class="dock-link">+settings</a>
<img src="http://a3.twimg.com/profile_images/463199849/gfx.espra.profile.tav_bigger.jpg" width="20px" height="20px" class="absmiddle" />
<a href="" class="dock-link">@tav</a>
<a href="" class="dock-link">#mainstream</a>
<a href="" class="dock-link">#espians</a>
<a href="" class="dock-link">+inbox</a>
<a href="" class="dock-link">+drafts</a>
<a href="" class="dock-link">+notifications</a>
<a href="" class="dock-link">+readlater</a>
<a href="" class="dock-link">~/help</a>
<a href="" class="dock-link">~/sessions</a>
<br /><br />
<a href="" class="dock-recent-link">~happyseaurchin/worldtimezero</a>
<a href="" class="dock-recent-link">#espians:plexnet</a>
</div>
</div>
<div id="dock-toggler"></div>

<div id="visi">

<table cellspacing="5px" cellpadding="5px" width="100%" style="border: 1px solid #f00; width: 100%;">
<tr>
<td>
<textarea id="letterbox" spellcheck="false" rows="2" style="padding: 0px">Yo!</textarea>
</td>
</tr>
</table>


<div id="letterbox-wrapper" style="background-color: #fff; margin-top: 10px; padding: 10px; -webkit-border-radius: 8px;">
<div style="background-color: #fff; margin-bottom: 10px;">
<!--
From:
<span id="" class="btn-thing-from"><img src="http://cloud.github.com/downloads/tav/plexnet/gfx.icon.twitter.png" width="16px" height="16px" class="absmiddle" /> tav</span>
<span id="" class="btn-thing-from"><img src="http://cloud.github.com/downloads/tav/plexnet/gfx.icon.facebook.gif" width="16px" height="16px" class="absmiddle" /> Tav Espian</span>
<br />
-->

<a href="#" title="@tav"><img src="http://cloud.github.com/downloads/tav/plexnet/gfx.icon.twitter.png" width="16px" height="16px" class="absmiddle" style="border: 1px solid #ccc;" /></a>

<a href="#" title="Tav Espian"><img src="http://cloud.github.com/downloads/tav/plexnet/gfx.icon.facebook.gif" width="16px" height="16px" class="absmiddle" style="border: 1px solid #ccc;" /></a>

&rarr;
 <input id="letterbox-to" size="20" />
<div style="float: right">
<span id="btn-tools" class="btn-thing">Tools</span>
&nbsp;
<span id="btn-send" class="btn-thing"><strong>Send</strong></span>
</div>
<div style="float: right; padding-top: 5px; margin-right: 10px;">
<span style="color: #069849; font-size: 14px;"><span id="letterbox-size" style="color: #666;">0</span></span>
</div>
</div>
<div style="clear: right;"></div>
<div id="tools">
<span id="btn-upload" class="btn-thing">Upload File</span>
&nbsp;
<span id="btn-preview" class="btn-thing">Preview</span><span id="btn-draft" class="btn-thing">Save Draft</span>
</div>
<div id="upload">
 <input type="file" value=" Upload " />
</div>
<div id="letterbox-meta">

</div>
</div>

<div id="dolumn-wrapper">

<div class="dolumn-lead">
<img src="http://a3.twimg.com/profile_images/463199849/gfx.espra.profile.tav_bigger.jpg" />
<div class="dolumn-notify">23</div>
</div>
<div class="dolumn">
<div class="dolumn-top">
<div class="dolumn-search">
<input type="search" autosave="com.espra.www" placeholder="Enter your query" results="10" />
<img src="http://dryicons.com/images/icon_sets/grace_icons_set/png/128x128/pin.png" width="32px" height="32px" />

&nbsp;
X
</div>
@tav
<img src="http://www.trustmap.org/static/img/default_profile_normal.png" width="16px" height="16px" class="absmiddle" />
</div>
Hello
</div>

<div class="dolumn-lead">
<img src="http://a3.twimg.com/profile_images/331435219/SocialBusiness-headerforlogo_bigger.jpg" />
<div class="dolumn-notify">PIN</div>
</div>
<div class="dolumn">
<div id="i1234" class="msg">
<img src="http://a3.twimg.com/profile_images/448318615/Photo_2_bigger.jpg" width="40px" height="40px" class="profile-img" />
@MosesKoinange Even so, it must be possible to come up with good solutions for these sort of issues.
<img src="http://www.trustmap.org/static/img/default_profile_normal.png" width="16px" height="16px" class="msg-pie" />
<hr class="clear" />
</div>
<div id="i1235" class="msg">
<img src="http://a3.twimg.com/profile_images/463199849/gfx.espra.profile.tav_bigger.jpg" width="40px" height="40px" class="profile-img" />
@evangineer will this work?
<img src="http://www.trustmap.org/static/img/default_profile_normal.png" width="16px" height="16px" class="msg-pie" />
<hr class="clear" />
</div>
<object width="425" height="344"><param name="movie" value="http://www.youtube.com/v/TU8iKWFcux0&hl=en&fs=1&"></param><param name="allowFullScreen" value="true"></param><param name="allowscriptaccess" value="always"></param><embed src="http://www.youtube.com/v/TU8iKWFcux0&hl=en&fs=1&" type="application/x-shockwave-flash" allowscriptaccess="always" allowfullscreen="true" width="425" height="344"></embed></object>

<img src="${STATIC('gfx/logo.espra.tent.large.png')}" width="342px" height="60px" />

</div>




</div>

</div>

<!--

<div id="page">

<div id="main-wrapper"><div id="main">

<div id="topbar">
  <form action="/.search" method="get" id="topbar-search-form" name="topbar-search-form">
  <div id="topbar-items">
    <div id="topbar-search-segment">
      <input id="topbar-search" type="text" name="q" autosave="com.espra.www" placeholder="Enter your query" results="10" /><a href="" id="topbar-search-button"></a>
    </div>
  % if ctx.is_logged_in():
    <a href="/.logout" id="topbar-logout-link">logout</a>
  % else:
    <a href="/.login" id="topbar-login-link">login</a>
  % endif
  </div>
  </form>
  <a href="/" title="Espra"><img src="${STATIC('gfx/logo.espra.white.png')}" alt="Espra" width="109px" height="32px" /></a>
</div>

<div id="ueberlein-wrapper">
<textarea id="ueberlein" spellcheck="false" rows="2">Yo!</textarea>
</div>

<div id="message-box">
&larr; creating new item... DONE!
<br />
&rarr; getting data from facebook<marquee>...</marquee>
</div>

<hr class="clear" />

<div style="background-color: #fff;" id="foobar">

Okay

  % if content_slot:
  <div id="content">
    <div>${content_slot|n,unicode}</div>
  </div>
  % endif

Meow
</div>

</div></div>

<div id="footer-wrapper"><div id="footer">

<ul id="footer-link">
  <li><a href="/" title="Espra Tent Home">home</a></li>
  <li><a href="/" title="">settings</a></li>
  <li><a href="/" title="">about</a></li>
  <li><a href="/" title="">contact</a></li>
  <li><a href="/" title="">status</a></li>
  <li><a href="/" title="">api</a></li>
  <li><a href="/" title="">terms</a></li>
  <li><a href="/" title="">privacy</a></li>
  <li class="last"><a href="http://help.espra.com" title="Help">help</a></li>
</ul>

</div></div>

</div>

-->

<script type="text/javascript">
if (document.location.hostname == 'localhost') {
  var GOOGLE_ANALYTICS_CODE = '';
} else {
  if (document.location.protocol == 'https:') {
    var gaJsHost = 'https://ssl.',
        GOOGLE_ANALYTICS_CODE = "UA-90176-20";
  } else {
    var gaJsHost = 'http://www.',
        GOOGLE_ANALYTICS_CODE = "UA-90176-19";
  };
  document.write(unescape("%3Cscript src='" + gaJsHost + "google-analytics.com/ga.js' type='text/javascript'%3E%3C/script%3E"));
}
</script>
<script type="text/javascript">
if (GOOGLE_ANALYTICS_CODE) {
  try {
    var pageTracker = _gat._getTracker(GOOGLE_ANALYTICS_CODE);
    pageTracker._trackPageview();
  } catch(err) {}
}
</script>

</body>
</html>
