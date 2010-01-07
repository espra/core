#! /usr/bin/env node

/*
 * This is a command line runner of jslint using Node.js
 *
 * Changes released into the Public Domain by tav <tav@espians.com>
 *
 * Adapted from rhino.js, Copyright (c) 2002 Douglas Crockford
 * Adapted from posixpath.py in the Python Standard Library.
 *
 */

/*global JSLINT */
/*jslint evil: true */

var posix = require('posix'),
    sys = require('sys');

// -----------------------------------------------------------------------------
// file manipulation utility funktions
// -----------------------------------------------------------------------------

function join_posix_path(p1, p2) {
    var path = p1;
    if (p2.charAt(0) === "/") {
        path = p2;
    } else if (path === "" || path.charAt(path.length - 1) === "/") {
        path += p2;
    } else {
        path += "/" + p2;
    }
    return path;
}

function split_posix_path(path) {
    var i = path.lastIndexOf('/') + 1,
        head = path.slice(0, i),
        tail = path.slice(i);
    if (head && head !== ('/' * head.length)) {
        head = head.replace(/\/*$/g, "");
    }
    return [head, tail];
}

function dirname(path) {
    return split_posix_path(path)[0];
}

// -----------------------------------------------------------------------------
// skript main funktion
// -----------------------------------------------------------------------------

function main() {
    var file = process.ARGV[2],
        source,
        i,
        error,
        jslint_path = join_posix_path(
          dirname(dirname(__filename)),
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
    } else {
        sys.puts("Success: No problems found in <" + file + ">");
    }
}

main();
