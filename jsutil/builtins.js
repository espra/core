// Released into the Public Domain by tav <tav@espians.com>

/*=

The builtins provides some basic functions that should really be part of the
standard Javascript environment. You can have the functions available globally
by adding the following at the start of your programs:

    require('builtins').install();

*/

var posix = require('posix'),
    sys = require('sys');

// -----------------------------------------------------------------------------
// inheritance support
// -----------------------------------------------------------------------------

Function.prototype.inherits = function (Parent) {
    this.prototype = new Parent();
    this.prototype.constructor = this;
};

// -----------------------------------------------------------------------------
// exceptions to complement builtin ones like Error and TypeError
// -----------------------------------------------------------------------------

function IOError(message) {
    this.message = message;
}

IOError.inherits(Error);
IOError.prototype.name = "IOError";

function ValueError(message) {
    this.message = message;
}

ValueError.inherits(Error);
ValueError.prototype.name = "ValueError";

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

/*=

The extend function allows you to easily extend:

    > a = {'a': 1}

    > b = {'b': 2}

    > extend(a, b)
    {'a': 1, 'b': 2}

    > foo = 2

    > bar = 3

    > foo + bar
    5

*/

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

function repr(object) {
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
exports.repr = repr;

exports.inspect = sys.p;
exports.print = sys.print;
exports.puts = sys.puts;
