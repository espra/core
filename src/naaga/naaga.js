// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

/*jslint plusplus: false */

// Naaga
// =====
var fs = require('fs'),
    sys = require('sys'),
//    unicodedata = require('./unicodedata'),
    ASCII = 0x80,
    HIGH_SURROGATE_START = 0xD800,
    HIGH_SURROGATE_END = 0xD8FF,
    LOW_SURROGATE_START = 0xDC00,
    LOW_SURROGATE_END = 0xDFFF,
    MAX_CODEPOINT = 0x10FFFF,
    REPLACEMENT_CHAR = 0xFFFD;

function tokenise(source) {

    var ascii,
        code,
        codepoint,
        high_surrogate = 0,
        i,
        length = source.length,
        runes = [];

    for (i = 0; i < length; i++) {

        // Get the UTF-16 character and convert it to a Unicode Codepoint.
        codepoint = source.charCodeAt(i);
        ascii = false;

        // Deal with surrogate pairs.
        if (high_surrogate) {
            if (codepoint >= LOW_SURROGATE_START && codepoint <= LOW_SURROGATE_END) {
                codepoint = (high_surrogate - HIGH_SURROGATE_START) *
                            1024 + 65536 +
                            (codepoint - LOW_SURROGATE_START);
            } else {
                // Replace a malformed surrogate pair with the Unicode
                // Replacement Character.
                codepoint = REPLACEMENT_CHAR;
            }
            high_surrogate = 0;
        } else if (codepoint >= HIGH_SURROGATE_START && codepoint <= HIGH_SURROGATE_END) {
            high_surrogate = codepoint;
            continue;

        // Handle the common case of ASCII before doing a check for the worst
        // case.
        } else if (codepoint <= ASCII) {
            ascii = true;

        // Do a sanity check to ensure that the codepoint isn't outside the
        // Unicode Character Set. If so, replace it with the Unicode Replacement
        // Character.
        } else if (codepoint >= MAX_CODEPOINT) {
            codepoint = REPLACEMENT_CHAR;
        }

        runes.push(codepoint);

    }

    return runes;
}

function timeit(n, func) {
    var args,
        start,
        stop,
        i;
    args = Array.prototype.slice.call(arguments, 2, arguments.length);
    start = new Date().getTime();
    for (i = 0; i < n; i++) {
        func.apply(func, args);
    }
    stop = new Date().getTime() - start;
    sys.puts(stop);
    return stop;
}

function bench(duration) {
    var args,
        func,
        i = 0,
        name = "unknown function",
        start,
        stop,
        total = 0;
    args = Array.prototype.slice.call(arguments, 1, arguments.length);
    if (typeof duration === 'string') {
        name = duration;
        duration = args[0];
        args = args.slice(1, args.length);
    }
    if (typeof duration === 'number') {
        func = args.shift();
    } else {
        func = duration;
        duration = 1;
    }
    sys.puts("benching <" + name + "> for " + duration + "s:\n");
    duration *= 1000;
    while (true) {
        start = new Date().getTime();
        func.apply(func, args);
        total = total + (new Date().getTime() - start);
        if (total > duration) {
            break;
        }
        i++;
    }
    sys.puts("    " + (total / 1000) + "\t" + i + " runs\n");
    return total;
}

var text = fs.readFileSync('/Users/tav/silo/ampify/src/naaga/src/foo.js');

sys.puts(tokenise("hello"));

timeit(1000, tokenise, text);

bench("tokeniser", 2.0, tokenise, text);
