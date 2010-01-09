/*
 * Released into the Public Domain by tav <tav@espians.com>
 *
 */

var assert = require('assert');
var sys = require('sys');

process.mixin(require('./builtins'));

var a = {'a': 1};
var b = {'b': 2};

assert.equal(
    extend(a, b)['b'],
    2,
    "Extended object doesn't have the expected new property"
);
