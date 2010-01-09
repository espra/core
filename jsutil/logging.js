/*
 * Released into the Public Domain by tav <tav@espians.com>
 *
 * The logging module provides an implementation of the Console API as
 * implemented by Firebug with support for log4j-esque configuration of loggers.
 *
 */

var sys = require('sys');

var dummy_function = function () {};

// Need to implement these functions properly ...

exports.log = function (message) {
    sys.puts(message);
};

exports.debug = dummy_function;
exports.info = dummy_function;
exports.warn = dummy_function;
exports.error = dummy_function;

exports.assert = dummy_function;
exports.dir = dummy_function;
exports.dirxml = dummy_function;
exports.trace = dummy_function;

exports.group = dummy_function;
exports.groupCollapsed = dummy_function;
exports.groupEnd = dummy_function;

exports.time = dummy_function;
exports.timeEnd = dummy_function;

exports.profile = dummy_function;
exports.profileEnd = dummy_function;

exports.count = dummy_function;