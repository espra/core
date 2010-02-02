/*
 * Released into the Public Domain by tav <tav@espians.com>
 *
 * The regular expressions are courtesy of Sean B. Palmer (sbp).
 *
 */

exports.double_quoted_strings = new RegExp('"[^"\\\\]*(?:\\\\.[^"\\\\]*)*"', 'g');
exports.single_quoted_strings = new RegExp("'[^'\\\\]*(?:\\\\.[^'\\\\]*)*'", 'g');
exports.star_comments = new RegExp('/\\*[^*]*(?:[^*]*|\\*(?!/))*\\*/', 'g');
