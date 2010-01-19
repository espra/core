#! /usr/bin/env node

/*
 * This is a command line runner of jslint using Node.js
 *
 * Changes released into the Public Domain by tav <tav@espians.com>
 *
 * Adapted from rhino.js, Copyright (c) 2002 Douglas Crockford
 *
 */

/*global JSLINT */
/*jslint evil: true, regexp: false */

var posix = require('posix'),
    posixpath = require('./posixpath'),
    sys = require('sys');

// -----------------------------------------------------------------------------
// skript main funktion
// -----------------------------------------------------------------------------

function main() {
    var file = process.ARGV[2],
        source,
        i,
        error,
        jslint_path = posixpath.join(
            posixpath.dirname(posixpath.dirname(posixpath.dirname(__filename))),
            'third_party/jslint/jslint.js'
        );
    eval(posix.cat(jslint_path).wait());
    if (!file) {
	    sys.puts("Usage: nodelint.js file.js");
        process.exit(1);
    }
    try {
        source = posix.cat(file).wait();
    } catch (err) {
        sys.puts("Error: Opening file <" + file + ">");
        sys.puts(err);
        process.exit(1);
    }
    source = source.replace(/^\#\!.*/, ''); // remove any shebangs
    if (!JSLINT(source, {bitwise: true, eqeqeq: true, immed: true,
                newcap: true, nomen: false, onevar: true, plusplus: true,
                predef: ['exports', 'module', 'require', 'process', '__filename', 'GLOBAL'],
                regexp: true, rhino: false, undef: true, white: true})) {
        for (i = 0; i < JSLINT.errors.length; i += 1) {
            error = JSLINT.errors[i];
            if (error) {
                sys.puts(
                  'Lint at line ' + error.line + ' character ' +
                  error.character + ': ' + error.reason
                );
                sys.puts((error.evidence || '')
                   .replace(/^\s*(\S*(\s+\S+)*)\s*$/, "$1"));
                sys.puts('');
            }
        }
        process.exit(2);
    }
}

main();
