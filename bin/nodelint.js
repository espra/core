#! /usr/bin/env node

/*
 * Adapted from rhino.js, Copyright (c) 2002 Douglas Crockford
 * Adapted from posixpath.py in the Python Standard Library.
 *
 * Changed released into the Public Domain by tav <tav@espians.com>
 *
 */

var posix = require('posix'),
    sys = require('sys');

// -----------------------------------------------------------------------------
// file manipulation utility funktions
// -----------------------------------------------------------------------------

function join_posix_path (p1, p2) {
  var path = p1;
  if (p2.match("^/") == "/") {
	path = p2;
  } else if (path == "" || path.match("/$") == "/") {
	path += p2;
  } else {
    path += "/" + p2;
  }
  return path;
}

function split_posix_path (path) {
  var i = path.lastIndexOf('/') + 1,
      head = path.slice(0, i),
      tail = path.slice(i);
  if (head && head != ('/' * head.length))
	head = head.replace(/\/*$/g, "");
  return [head, tail];
}

function dirname (path) {
  return split_posix_path(path)[0];
}

// -----------------------------------------------------------------------------
// core jslint loader
// -----------------------------------------------------------------------------

function load_jslint_source () {
  var jslint_path = join_posix_path(
	dirname(dirname(__filename)),
	'third_party/jslint/jslint.js'
    );
  eval(posix.cat(jslint_path).wait());
}

// -----------------------------------------------------------------------------
// skript main funktion
// -----------------------------------------------------------------------------

function main () {
  var file = process.ARGV[2];
  if (!file) {
	sys.puts("Usage nodelint.js file.js");
	process.exit(1);
  }
  print(file);
}

var print = sys.puts;

load_jslint_source();
main();

