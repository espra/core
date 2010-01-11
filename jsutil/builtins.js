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

var posix = require('posix'),
    sys = require('sys');

// -----------------------------------------------------------------------------
// exceptions to complement builtin ones like Error and TypeError
// -----------------------------------------------------------------------------

function IOError(message) {
    this.name = "IOError";
    this.message = message;
}

IOError.prototype = new Error();

function ValueError(message) {
    this.name = "ValueError";
    this.message = message;
}

ValueError.prototype = new Error();

// -----------------------------------------------------------------------------
// builtin funktions
// -----------------------------------------------------------------------------

function dump(object, with_values) {
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
}

function extend(object, update) {

    var p,
        getter,
        setter;

    for (p in update) {
        if (update.hasOwnProperty(p)) { 
            getter = update.__lookupGetter__(p);
            setter = update.__lookupSetter__(p);
            if (getter || setter) {
                if (getter) {
                    object.__defineGetter__(p, getter);
                }
                if (setter) {
                    object.__defineSetter__(p, setter);
                }
            } else {
                object[p] = update[p];
            }
        }
    }

    return object;

}

function read(filepath) {
    try {
        return posix.cat(filepath).wait();
    } catch (err) {
        throw new IOError(err.message + ": " + filepath);
    }
}

exports.install = function () {
    if (GLOBAL.__builtins_installed__) {
        return;
    }
    GLOBAL.__builtins_installed__ = true;
    for (var name in exports) {
        if (exports.hasOwnProperty(name)) {
            GLOBAL[name] = exports[name];
        }
    }
};

exports.dump = dump;
exports.extend = extend;
exports.read = read;

exports.inspect = sys.p;
exports.print = sys.print;
exports.puts = sys.puts;
