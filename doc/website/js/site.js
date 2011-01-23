/**
 * Cookie plugin
 *
 * Copyright (c) 2006 Klaus Hartl (stilbuero.de)
 * Licensed under the MIT license:
 * http://www.opensource.org/licenses/mit-license.php
 *
 */

jQuery.cookie = function(name, value, options) {
    if (typeof value != 'undefined') { // name and value given, set cookie
        options = options || {};
        if (value === null) {
            value = '';
            options.expires = -1;
        }
        var expires = '';
        if (options.expires && (typeof options.expires == 'number' || options.expires.toUTCString)) {
            var date;
            if (typeof options.expires == 'number') {
                date = new Date();
                date.setTime(date.getTime() + (options.expires * 24 * 60 * 60 * 1000));
            } else {
                date = options.expires;
            }
            expires = '; expires=' + date.toUTCString(); // use expires attribute, max-age is not supported by IE
        }
        // CAUTION: Needed to parenthesize options.path and options.domain
        // in the following expressions, otherwise they evaluate to undefined
        // in the packed version for some reason...
        var path = options.path ? '; path=' + (options.path) : '';
        var domain = options.domain ? '; domain=' + (options.domain) : '';
        var secure = options.secure ? '; secure' : '';
        document.cookie = [name, '=', encodeURIComponent(value), expires, path, domain, secure].join('');
    } else { // only name given, get cookie
        var cookieValue = null;
        if (document.cookie && document.cookie != '') {
            var cookies = document.cookie.split(';');
            for (var i = 0; i < cookies.length; i++) {
                var cookie = jQuery.trim(cookies[i]);
                // Does this cookie string begin with the name we want?
                if (cookie.substring(0, name.length + 1) == (name + '=')) {
                    cookieValue = decodeURIComponent(cookie.substring(name.length + 1));
                    break;
                }
            }
        }
        return cookieValue;
    }
};

// toJSON

(function ($) {
    var m = {
            '\b': '\\b',
            '\t': '\\t',
            '\n': '\\n',
            '\f': '\\f',
            '\r': '\\r',
            '"' : '\\"',
            '\\': '\\\\'
        },
        s = {
            'array': function (x) {
                var a = ['['], b, f, i, l = x.length, v;
                for (i = 0; i < l; i += 1) {
                    v = x[i];
                    f = s[typeof v];
                    if (f) {
                        v = f(v);
                        if (typeof v == 'string') {
                            if (b) {
                                a[a.length] = ',';
                            }
                            a[a.length] = v;
                            b = true;
                        }
                    }
                }
                a[a.length] = ']';
                return a.join('');
            },
            'boolean': function (x) {
                return String(x);
            },
            'null': function (x) {
                return "null";
            },
            'number': function (x) {
                return isFinite(x) ? String(x) : 'null';
            },
            'object': function (x) {
                if (x) {
                    if (x instanceof Array) {
                        return s.array(x);
                    }
                    var a = ['{'], b, f, i, v;
                    for (i in x) {
                        v = x[i];
                        f = s[typeof v];
                        if (f) {
                            v = f(v);
                            if (typeof v == 'string') {
                                if (b) {
                                    a[a.length] = ', ';
                                }
                                a.push(s.string(i), ':', v);
                                b = true;
                            }
                        }
                    }
                    a[a.length] = '}';
                    return a.join('');
                }
                return 'null';
            },
            'string': function (x) {
                if (/["\\\x00-\x1f]/.test(x)) {
                    x = x.replace(/([\x00-\x1f\\"])/g, function(a, b) {
                        var c = m[b];
                        if (c) {
                            return c;
                        }
                        c = b.charCodeAt();
                        return '\\u00' +
                            Math.floor(c / 16).toString(16) +
                            (c % 16).toString(16);
                    });
                }
                return '"' + x + '"';
            }
        };

	$.toJSON = function(v) {
		var f = isNaN(v) ? s[typeof v] : s['number'];
		if (f) return f(v);
	};

})(jQuery);

// browser detekt from ppk

var BrowserDetect = {
	init: function () {
		this.browser = this.searchString(this.dataBrowser) || "An unknown browser";
		this.version = this.searchVersion(navigator.userAgent)
			|| this.searchVersion(navigator.appVersion)
			|| "an unknown version";
		this.OS = this.searchString(this.dataOS) || "an unknown OS";
	},
	searchString: function (data) {
		for (var i=0;i<data.length;i++)	{
			var dataString = data[i].string;
			var dataProp = data[i].prop;
			this.versionSearchString = data[i].versionSearch || data[i].identity;
			if (dataString) {
				if (dataString.indexOf(data[i].subString) != -1)
					return data[i].identity;
			}
			else if (dataProp)
				return data[i].identity;
		}
	},
	searchVersion: function (dataString) {
		var index = dataString.indexOf(this.versionSearchString);
		if (index == -1) return;
		return parseFloat(dataString.substring(index+this.versionSearchString.length+1));
	},
	dataBrowser: [
		{
			string: navigator.userAgent,
			subString: "Chrome",
			identity: "Chrome"
		},
		{ 	string: navigator.userAgent,
			subString: "OmniWeb",
			versionSearch: "OmniWeb/",
			identity: "OmniWeb"
		},
		{
			string: navigator.vendor,
			subString: "Apple",
			identity: "Safari",
			versionSearch: "Version"
		},
		{
			prop: window.opera,
			identity: "Opera"
		},
		{
			string: navigator.vendor,
			subString: "iCab",
			identity: "iCab"
		},
		{
			string: navigator.vendor,
			subString: "KDE",
			identity: "Konqueror"
		},
		{
			string: navigator.userAgent,
			subString: "Firefox",
			identity: "Firefox"
		},
		{
			string: navigator.vendor,
			subString: "Camino",
			identity: "Camino"
		},
		{		// for newer Netscapes (6+)
			string: navigator.userAgent,
			subString: "Netscape",
			identity: "Netscape"
		},
		{
			string: navigator.userAgent,
			subString: "MSIE",
			identity: "Explorer",
			versionSearch: "MSIE"
		},
		{
			string: navigator.userAgent,
			subString: "Gecko",
			identity: "Mozilla",
			versionSearch: "rv"
		},
		{ 		// for older Netscapes (4-)
			string: navigator.userAgent,
			subString: "Mozilla",
			identity: "Netscape",
			versionSearch: "Mozilla"
		}
	],
	dataOS : [
		{
			string: navigator.platform,
			subString: "Win",
			identity: "Windows"
		},
		{
			string: navigator.platform,
			subString: "Mac",
			identity: "Mac"
		},
		{
			   string: navigator.userAgent,
			   subString: "iPhone",
			   identity: "iPhone/iPod"
	    },
		{
			string: navigator.platform,
			subString: "Linux",
			identity: "Linux"
		}
	]
};

try {
  BrowserDetect.init();
} catch (err) {};

/* The rest of this file is:
 *
 * Public Domain (-) 2009-2011 The Ampify Authors.
 * See the Ampify UNLICENSE file for details.
 *
 */

// some konstants

var country_selected = false,
    ipinfo = {},
    ipinfodb_queried = false,
    language_options_revealed = false,
    ESPIANS = [
        'alextomkins',
        'cre8radix',
        'evangineer',
        'evangineer',
        'happyseaurchin',
        'jeffarch',
        'jmccanewhitney',
        'olasofia',
        'oierw',
        'sbp',
        'tav',
        'tav',
        'thruflo',
        'yncyrydybyl'
    ],
    ESPIAN_BLOGS = {
        'cre8radix': 'http://cre8radix.net',
        'happyseaurchin': 'http://2020worldwalk.blogspot.com',
        'jeffarch': 'http://adkblueline.blogspot.com',
        'olasofia': 'http://sofiabustamante.com',
        'sbp': 'http://inamidst.com',
        'tav': 'http://tav.espians.com',
        'thruflo': 'http://thruflo.com',
        'yncyrydybyl': 'http://www.c-base.org'
    },
    ESPIANS_NO_TWITTER = {
        'oierw': true
    },
    ESPIANS_COUNT = ESPIANS.length,
    IPINFO_KEYS = [
        ['ip', 'Ip'],
        ['country', 'CountryName'],
        ['country_id', 'CountryCode'],
        ['lat', 'Latitude'],
        ['lon', 'Longitude'],
        ['region', 'RegionName'],
        ['city', 'City']
    ];

// cache the ipinfo to be nice to the wonderful ipinfodb.com guys

function ipinfo_handle (form, on_main_form) {
  if (!ipinfodb_queried) {
    if ($.cookie('ipinfo')) {
      ipinfo = eval('(' + $.cookie('ipinfo') + ')');
    } else {
      ipinfo_get(form, on_main_form);
      return;
    }
  }
  ipinfodb_queried = true;
  for (var i=0; i < IPINFO_KEYS.length; i++) {
    var form_key = IPINFO_KEYS[i][0];
    try {
      if (ipinfo[form_key])
        form[form_key].value = ipinfo[form_key];
    } catch (err) {};
  }
  /*
  if ((on_main_form) && (!country_selected)) {
    if (ipinfo['country_id'])
      $('#country-' + ipinfo['country_id'].toLowerCase()).attr('selected', 1);
  };
  */
}

function ipinfo_get (form, on_main_form) {
  $.getJSON('http://ipinfodb.com/ip_query.php?output=json&callback=?', function (data) {
    ipinfodb_queried = true;
    var error = null;
    for (var i=0; i < IPINFO_KEYS.length; i++) {
      var form_key = IPINFO_KEYS[i][0],
          data_key = IPINFO_KEYS[i][1];
      try {
        var value = data[data_key];
        if (value) {
          ipinfo[form_key] = value;
          form[form_key].value = value;
        }
      } catch (err) {
        error = true;
      }
    }
    if (!error)
      $.cookie('ipinfo', $.toJSON(ipinfo), {expires: 300});
    /*
    if ((on_main_form) && (!country_selected)) {
      if (ipinfo['country_id'])
        $('#country-' + ipinfo['country_id'].toLowerCase()).attr('selected', 1);
    };
    */
  });
}

function set_user_info (form, referrer) {
    var first_referrer = $.cookie('referrer');
    if (referrer) {
        if (!first_referrer) {
            first_referrer = referrer;
            $.cookie('referrer', referrer, {expires: 300});
        }
        form['referrer'].value = referrer;
    }
    if (first_referrer)
        form['first_referrer'].value = first_referrer;
    try {
        form['browser'].value = BrowserDetect.browser;
        form['browser_version'].value = BrowserDetect.version;
        form['os'].value = BrowserDetect.OS;
    } catch (err) {};
    try {
        form['useragent'].value = navigator.userAgent;
    } catch (err) {};
}

// ondocumentready funktion for the support page

function init_support_page () {
    var supporter_form = document.forms['supporterform'];
    set_user_info(supporter_form, document.referrer);
    ipinfo_handle(supporter_form, true);
    var share_elem_1 = document.getElementById('sharethis-1');
    if (share_elem_1) {
        var share_button_1 = SHARETHIS.addEntry({url: 'http://ampify.it'}, {button: false});
        share_button_1.attachButton(share_elem_1);
    }
}

// utility funktions

function change_gender_option () {
    var gender = $('#supporter-gender').val();
    if (gender == 'female')
        $('#supporter-form').removeClass('male').addClass('female');
    if (gender == 'male')
        $('#supporter-form').removeClass('female').addClass('male');
    return false;
}

function reveal_language_options () {
    if (language_options_revealed) {
        $('#menu-lang-form').hide();
        language_options_revealed = false;
    } else {
        $('#menu-lang-form').show();
        language_options_revealed = true;
    }
    return false;
}

function google_translate_page (lang_options) {
    var google_translate_url = 'http://translate.google.com/translate?u='+encodeURIComponent(window.location);
    if (lang_options.options[lang_options.selectedIndex].value !="") {
        parent.location=google_translate_url+lang_options.options[lang_options.selectedIndex].value;
    }
}

var table_of_contents_display_status = false;

function show_table_of_contents () {
    if (!table_of_contents_display_status) {
        table_of_contents_display_status = true;
        document.getElementById('table-of-contents').style.display = 'block';
    } else {
        table_of_contents_display_status = false;
        document.getElementById('table-of-contents').style.display = 'none';
        return false;
    }
    return false;
}

$(function () {
    var chosen = [],
        i,
        selected,
        notfound,
        extra,
        tweet_prefix,
        tweet_suffix_img,
        tweet_suffix_lnk,
        container,
        content_handlers=$('.table-of-contents-handler');
    if (content_handlers.length) {
        if (window.location.hash.substr(1) === 'table-of-contents') {
            $('.contents').show();
            table_of_contents_display_status = true;
        };
        content_handlers.click(show_table_of_contents);
    }
    $('a[href=#ignore-this]').parent().hide();
    for (i=0; i < 6; i++) {
        notfound = true;
        while (notfound) {
            selected = ESPIANS[Math.floor(ESPIANS_COUNT * Math.random())];
            if (chosen.indexOf(selected) == -1) {
                chosen.push(selected);
                notfound = false;
            }
        }
    }
    container = $('#footer-espians-tr');
    chosen.sort();
    for (i=0; i < chosen.length; i++) {
        selected = chosen[i];
        extra = '';
        if (ESPIAN_BLOGS[selected]) {
            extra = ', <a href="'+ESPIAN_BLOGS[selected]+'">blog</a>';
        }
        if (ESPIANS_NO_TWITTER[selected]) {
            tweet_prefix = tweet_suffix_img = "";
            tweet_suffix_lnk = '<a href="#">@'+selected+'</a>';
        } else {
            tweet_prefix = '<a href="http://twitter.com/'+selected+'" title="Follow @'+selected+'">';
            tweet_suffix_img = "</a>";
            tweet_suffix_lnk = '@'+selected+'</a>';
        }
        $('<td class="footer-follow">'+tweet_prefix+'<img src="http://static.ampify.it/profile.'+selected+'.jpg" alt="@'+selected+'" width="69px" height="86px" />'+tweet_suffix_img+'<div>'+tweet_prefix+tweet_suffix_lnk+extra+'</div></td>').appendTo(container);
    }
    $('.sharethislink').each(function () {
        SHARETHIS.addEntry({url: 'http://ampify.it'}, {button: false}).attachButton(this);
    });
});
