// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

var NOW = new Date();

var get_relative_time = function (timestamp) {
    var delta = (NOW - timestamp) / 1000;
    if (delta < 60)
        return "less than " + Math.floor(delta) + " seconds ago";
    delta = Math.floor(delta / 60);
    if (delta === 0)
        return 'less than a minute ago';
    if (delta === 1)
        return 'a minute ago';
    if (delta < 60)
        return delta + ' minutes ago';
    if (delta < 62)
        return 'about 1 hour ago';
    if (delta < 120)
        return 'about 1 hour and ' + (delta - 60) + ' minutes ago';
    if (delta < 1440)
        return 'about ' + Math.floor(delta / 60) + ' hours ago';
    if (delta < 2880)
        return '1 day ago';
    if (delta < 43200)
        return Math.floor(delta / 1440) + ' days ago';
    if (delta < 86400)
        return 'about 1 month ago';
    if (delta < 525960)
        return Math.floor(delta / 43200) + ' months ago';
    if (delta < 1051199)
        return 'about 1 year ago';
    return 'over ' + Math.floor(delta / 525960) + ' years ago';
};


$(function () {
    $('.timestamp').each(function () {
        var $span = $(this);
        $span.text(get_relative_time(new Date(parseInt($span.text()) * 1000)));
    });
});