// Public Domain (-) 2009-2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

var language_options_revealed = false,
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
    ESPIANS_COUNT = ESPIANS.length;

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
        $('<td class="footer-follow">'+tweet_prefix+'<img src="http://static.ampify.net/profile.'+selected+'.jpg" alt="@'+selected+'" width="69px" height="86px" />'+tweet_suffix_img+'<div>'+tweet_prefix+tweet_suffix_lnk+extra+'</div></td>').appendTo(container);
    }
});
