/*
 * Released into the Public Domain by tav <tav@espians.com>
 *
 * The builtins provides some basic functions that should really be part of the
 * standard Javascript environment. You can have the functions available
 * globally by adding the following at the start of your programs:
 *
 *   require('builtins').install();
 *
 */

var sys = require('sys');

var dump = function (object, with_values) {
    var k;
    for (k in object) {
        if (object.hasOwnProperty(k)) { 
            sys.puts(k);
            if (with_values) {
                sys.puts(object[k]);
            }
            sys.puts('-----------------------------------------------------------');
        }
    }
};

exports.install = function () {
    if (GLOBAL.__builtins_installed__) {
        return;
    }
    GLOBAL.__builtins_installed__ = true;
    GLOBAL.dump = dump;
    GLOBAL.print = sys.print;
    GLOBAL.puts = sys.puts;
};

exports.dump = dump;