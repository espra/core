// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The following is the QUnit Test Suite for friendly.js

/*global test, expect, ok, stop, setTimeout, start, asyncTest, $, equals,
 console, search_friendfeed, search_twitter, rfc3339_to_date, get_relative_time*/


module('Friendly Tests');

//  number of entries to retrieve from API calls
var _num = 5,
    _earliest_date = null;


test("friendfeed test", function () {
    // friendfeed: name of user whose feed we want to retrieve
    var usr = 'scobleizer';

    expect(1);
    stop();

    // call friendfeed api
    $.getJSON(
        // construct the fetch url
        'http://friendfeed.com/api/feed/user/' + usr + '?num=' + _num + '&amp;callback=?',

        // build content from api results
        function (data) {
            // loop for each friendfeed entry retrieved
            $.each(data.entries, function (i, entry) {
                // ignore entry if it is marked as 'hidden'
                if (entry.hidden !== true) {
                    console.log(entry.title);
                }
            });

            ok(data, 'friendfeed returned a feed');
        }
    );

    setTimeout(function () {
        start();
    }, 1000);
});


test("friendfeed search test", function () {
    // friendfeed: define the search term we want to find entries for
    var query = 'ipad',
        ffdata = null;

    expect(1);
    stop();
/*
    // call friendfeed api
    $.getJSON(
        // construct the fetch url
        'http://friendfeed-api.com/v2/search?q=' + query + '&amp;num=' + _num + '&amp;callback=?',

        // build content from api results
        function (data) {
            // loop for each friendfeed entry retrieved
            $.each(data.entries, function (i, entry) {
                // ignore entry if it is marked as 'hidden'
                if (entry.hidden !== true) {
                    console.log(entry.body);
                    if (i === 4) {
                        _earliest_date = entry.date;
                    }
                }
            });

            ok(data, 'friendfeed search returned results');
        }
    );
*/

    if ((ffdata = search_friendfeed(query, _num))) {
        // loop for each friendfeed entry retrieved
        $.each(ffdata.entries, function (i, entry) {

            // ignore entry if it is marked as 'hidden'
            if (entry.hidden !== true) {
                console.log(entry.body);

                if (i === _num - 1) {
                    _earliest_date = entry.date;
                }
            }
        });

        ok(ffdata, 'friendfeed search returned valid results');
    }

    setTimeout(function () {
        start();
    }, 3000);
});


test("friendfeed date processing", function () {
    var ffdate = null,
        reltime = null;

    expect(2);

    if (_earliest_date) {
        console.log(_earliest_date);
        ffdate = rfc3339_to_date(_earliest_date);
        ok(ffdate, 'friendfeed returned a valid date: ' + ffdate);
        reltime = get_relative_time(ffdate);
        ok(reltime, 'get_relative_time returned the following: ' + reltime);
    }
});


test("twitter test", function () {
    // twitter: name of user whose feed we want to retrieve
    var usr = 'guykawasaki';

    expect(1);
    stop();

    // call twitter api
    $.getJSON(
        // construct the fetch url
        'http://twitter.com/statuses/user_timeline/' + usr + '.json?count=' + _num + '&callback=?',

        // build content from api results
        function (data, status) {
            // loop for each twitter status retrieved
            $.each(data, function (i) {
                console.log(this.text);
            });

            ok(data, 'twitter returned returned a timeline');
        }
    );

    setTimeout(function () {
        start();
    }, 2000);
});


test("twitter search test", function () {
    // twitter: define the search term we want to find entries for
    var query = 'notion ink adam',
        twdata = null;

    expect(1);
    stop();
/*
    // call twitter api
    $.getJSON(
        // construct the fetch url
        'http://search.twitter.com/search.json?q=' + query + '&rpp=' + _num + '&callback=?',

        // build content from api results
        function (data, status) {
            // loop for each twitter status retrieved
            $.each(data.results, function (i, result) {
                console.log(result.text);
            });

            ok(data, 'twitter search returned results');
        }
    );
*/

    if ((twdata = search_twitter(query, _num))) {
        // loop for each twitter status retrieved
        $.each(twdata.results, function (i, result) {
            console.log(result.text);
        });

        ok(twdata, 'twitter search returned results');
    }

    setTimeout(function () {
        start();
    }, 2000);
});

