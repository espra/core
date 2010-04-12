#! /usr/bin/env node

// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

/*jslint plusplus: false */

// Naaga
// =====
var fs = require('fs'),
    sys = require('sys'),
    ucd = require('./ucd'),
    cat = ucd.cat,
    catranges = ucd.catranges,
    catrange_length = catranges.length,
    catnames = ucd.catnames,
    Cc = ucd.Cc,
    Cf = ucd.Cf,
    Co = ucd.Co,
    Cs = ucd.Cs,
    Ll = ucd.Ll,
    Lm = ucd.Lm,
    Lo = ucd.Lo,
    Lt = ucd.Lt,
    Lu = ucd.Lu,
    Mc = ucd.Mc,
    Me = ucd.Me,
    Mn = ucd.Mn,
    Nd = ucd.Nd,
    Nl = ucd.Nl,
    No = ucd.No,
    Pc = ucd.Pc,
    Pd = ucd.Pd,
    Pe = ucd.Pe,
    Pf = ucd.Pf,
    Pi = ucd.Pi,
    Po = ucd.Po,
    Ps = ucd.Ps,
    Sc = ucd.Sc,
    Sk = ucd.Sk,
    Sm = ucd.Sm,
    So = ucd.So,
    Zl = ucd.Zl,
    Zp = ucd.Zp,
    Zs = ucd.Zs,
    CasedLetter = ucd.CasedLetter,
    Letter = ucd.Letter,
    Mark = ucd.Mark,
    Number_ = ucd.Number,
    Other = ucd.Other,
    Punctuation = ucd.Punctuation,
    Separator = ucd.Separator,
    Symbol = ucd.Symbol,
    Unknown = ucd.Unknown,
    ASCII = 0x80,
    HIGH_SURROGATE_START = 0xD800,
    HIGH_SURROGATE_END = 0xD8FF,
    LOW_SURROGATE_START = 0xDC00,
    LOW_SURROGATE_END = 0xDFFF,
    MAX_CODEPOINT = 0x10FFFF,
    REPLACEMENT_CHAR = 0xFFFD;

// The ``tokenise`` function uses a novel approach put forward by sbp (Sean B.
// Palmer). It splits the given ``source`` into tokens of distinct
// categories as specified in Unicode 5.2.
//
// It returns a sequence of ``[category, ascii_flag, codepoints, string]``
// tokens:
//
// * The ``category`` points to a constant representing the unicode category.
//
// * The ``ascii_flag`` indicates whether the segment is comprised of only ASCII
//   characters.
//
// * The ``codepoints`` are the unicode codepoints for use by any normalisation
//   function.
//
// * The ``string`` is the segment of the source for the current token.
function tokenise(source) {

    var ascii,
        category,
        catrange,
        codepoint,
        high_surrogate = 0,
        idx_ascii = true,
        idx_cat = -1,
        idx_lst = [],
        idx_start = 0,
        i,
        j,
        length = source.length,
        pushed = false,
        tokens = [];

    for (i = 0; i < length; i++) {

        // Get the UTF-16 unit and convert it to a Unicode Codepoint.
        codepoint = source.charCodeAt(i);
        ascii = false;

        // Deal with surrogate pairs.
        if (high_surrogate !== 0) {
            if ((codepoint >= LOW_SURROGATE_START) &&
                (codepoint <= LOW_SURROGATE_END)) {
                codepoint = (high_surrogate - HIGH_SURROGATE_START) *
                            1024 + 65536 +
                            (codepoint - LOW_SURROGATE_START);
            } else {
                // Replace a malformed surrogate pair with the Unicode
                // Replacement Character.
                codepoint = REPLACEMENT_CHAR;
            }
            high_surrogate = 0;
        } else if ((codepoint >= HIGH_SURROGATE_START) &&
                   (codepoint <= HIGH_SURROGATE_END)) {
            high_surrogate = codepoint;
            continue;

        // Handle the common case of ASCII before doing a check for the worst
        // case.
        } else if (codepoint < ASCII) {
            ascii = true;

        // Do a sanity check to ensure that the codepoint isn't outside the
        // Unicode Character Set. If so, replace it with the Unicode Replacement
        // Character.
        } else if (codepoint >= MAX_CODEPOINT) {
            codepoint = REPLACEMENT_CHAR;
        }

        category = cat[codepoint];

        if (typeof category === "undefined") {
            for (j = 0; j < catrange_length; j++) {
                catrange = catranges[j];
                if ((codepoint >= catrange[0]) &&
                    (codepoint <= catrange[1])) {
                    category = catrange[2];
                    break;
                }
            }
            if (typeof category === "undefined") {
                category = Unknown;
            }
        }

        if (category === idx_cat) {
            idx_ascii = idx_ascii && ascii;
            idx_lst.push(codepoint);
        } else {
            tokens.push(
                [idx_cat, idx_ascii, idx_lst, source.slice(idx_start, i)]
            );
            idx_ascii = ascii;
            idx_cat = category;
            idx_lst = [codepoint];
            idx_start = i;
        }

    }

    tokens.push(
        [idx_cat, idx_ascii, idx_lst, source.slice(idx_start, i)]
    );

    return tokens;

}

function print_tokens(tokens) {
    var i,
        token,
        number_of_tokens = tokens.length;
    for (i = 1; i < number_of_tokens; i++) {
        token = tokens[i];
        sys.puts(JSON.stringify([catnames[token[0]], token[1], token[3]]));
    }
}

function parse() {
}

function evaluate() {
}

function compile() {
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

print_tokens(tokenise("hello, wo—rld__.{(foo-bar\\"));

// timeit(1000, tokenise, text);
// bench("tokeniser", 2.0, tokenise, text);
