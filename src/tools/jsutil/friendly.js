// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// friendly.js -- a widget that displays live search results from
// Friendfeed, Twitter and more

/*global $, console*/

// Dates in JSON and dates in the FriendFeed extension elements in the
// Atom and RSS feeds are in RFC 3339 format in UTC.

// convert rfc3339 formatted date to a normal javascript date object
// see http://www.johngirvin.com/blog/archives/reading-friendfeed-with-jquery.html
function rfc3339_to_date(val) {
    var pattern = /^(\d{4})(?:-(\d{2}))?(?:-(\d{2}))?(?:[Tt](\d{2}):(\d{2}):(\d{2})(?:\.(\d*))?)?([Zz])?(?:([+\-])(\d{2}):(\d{2}))?$/,
        m = pattern.exec(val),
        year = m[1] ? m[1] : 0,
        month = m[2] ? m[2] - 1 : 0,
        day = m[3] ? m[3] : 0,
        hour = m[4] ? m[4] : 0,
        minute = m[5] ? m[5] : 0,
        second = m[6] ? m[6] : 0,
        millis = m[7] ? m[7] : 0,
        gmt = m[8],
        dir = m[9],
        offhour = m[10] ? m[10] : 0,
        offmin = m[11] ? m[11] : 0,
        offset = 0;

    if (dir && offhour && offmin) {
        offset = ((offhour * 60) + offmin);

        if (dir === "+") {
            minute -= offset;
        } else if (dir === "-") {
            minute += offset;
        }
    }

    return new Date(Date.UTC(year, month, day, hour, minute, second, millis));
}


var get_relative_time = function (from) {
    var NOW = new Date(),
        date = new Date(),
        distance_in_seconds = null,
        distance_in_minutes = null;

    date.setTime(Date.parse(from));
    distance_in_seconds = ((NOW - date) / 1000);
    distance_in_minutes = Math.floor(distance_in_seconds / 60);

    if (distance_in_minutes === 0) {
        return 'less than a minute ago';
    }

    if (distance_in_minutes === 1) {
        return 'a minute ago';
    }

    if (distance_in_minutes < 45) {
        return distance_in_minutes + ' minutes ago';
    }

    if (distance_in_minutes < 90) {
        return 'about 1 hour ago';
    }

    if (distance_in_minutes < 1440) {
        return 'about ' + Math.floor(distance_in_minutes / 60) + ' hours ago';
    }

    if (distance_in_minutes < 2880) {
        return '1 day ago';
    }

    if (distance_in_minutes < 43200) {
        return Math.floor(distance_in_minutes / 1440) + ' days ago';
    }

    if (distance_in_minutes < 86400) {
        return 'about 1 month ago';
    }

    if (distance_in_minutes < 525960) {
        return Math.floor(distance_in_minutes / 43200) + ' months ago';
    }

    if (distance_in_minutes < 1051199) {
        return 'about 1 year ago';
    }

    return 'over ' + Math.floor(distance_in_minutes / 525960) + ' years ago';
};


var search_friendfeed = function (query, count) {
    var results = null;

    results = $.getJSON(
        // construct the fetch url
        'http://friendfeed-api.com/v2/search?q=' + query + '&amp;num=' +
            count + '&amp;callback=?',

        // build content from api results
        function (data) {
            return data;
        }
    );

    return results;
};

var search_twitter = function (query, count) {
    var results = null;

    results = $.getJSON(
        // construct the fetch url
        'http://search.twitter.com/search.json?q=' + query + '&rpp=' + count +
            '&callback=?',

        // build content from api results
        function (data, status) {
            console.log('Twitter returned the status: ' + status);
            return data;
        }
    );

    return results;
};

// http://friendfeed-api.com/v2/search?q=ampify+OR+espians+OR+group%3Aampify+OR+group%3Aespians#
