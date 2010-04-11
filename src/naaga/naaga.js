// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

/*jslint plusplus: false */

var fs = require('fs'),
    sys = require('sys'),
    HIGH_SURROGATE_START = 0xD800,
    HIGH_SURROGATE_END = 0xD8FF,
    LOW_SURROGATE_START = 0xDC00,
    LOW_SURROGATE_END = 0xDFFF,
    MAX_CODEPOINT = 0x10FFFF,
    REPLACEMENT_CHAR = 0xFFFD;

function tokenise(source) {

    var code,
        codepoint,
        high_surrogate = 0,
        i,
        length = source.length,
        runes = [];

    for (i = 0; i < length; i++) {

        // Get the UTF-16 character.
        code = source.charCodeAt(i);

        // Convert it to a Unicode Codepoint.
        if (high_surrogate) {
            if (code >= LOW_SURROGATE_START && code <= LOW_SURROGATE_END) {
                codepoint = (high_surrogate - HIGH_SURROGATE_START) *
                            1024 + 65536 +
                            (code - LOW_SURROGATE_START);
            } else {
                // Replace a malformed surrogate pair with the Unicode
                // Replacement Character.
                codepoint = REPLACEMENT_CHAR;
            }
            high_surrogate = 0;
        } else if (code >= HIGH_SURROGATE_START && code <= HIGH_SURROGATE_END) {
            high_surrogate = code;
            continue;
        } else {
            codepoint = code;
        }

        runes.push(codepoint);

        // non-ascii-seen

        // Do a sanity check to ensure that the codepoint isn't outside the
        // Unicode Character Set. If so, replace it with the Unicode Replacement
        // Character.
        // if (codepoint >= MAX_CODEPOINT) {
        //     codepoint = REPLACEMENT_CHAR;
        // }

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
        func.apply(tokenise, args);
    }
    stop = new Date().getTime() - start;
    sys.puts(stop);
    return stop;
}

var text = fs.readFileSync('/Users/tav/silo/ampify/src/naaga/src/foo.js');

sys.puts(tokenise("hello"));

timeit(1000, tokenise, text);
// timeit(1000, tokenise, text);
