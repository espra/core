// Released into the Public Domain by tav <tav@espians.com>

/*=

The module provides some utility functions and extensions of builtin objects to
complement the standard Javascript environment.

The `Object.extend` method allows you to easily update an object with the
contents of another:

    > a = {'a': 1}

    > b = {'b': 2}

    > a.extend(a, b)
    {'a': 1, 'b': 2}

    > foo = 2

    > bar = 3

    > foo + bar
    5

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

Object.prototype.dump = function (object, with_values) {
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

Object.prototype.extend = function (update) {
    return extend(this, update);
};

function read(filepath) {
    try {
        return posix.cat(filepath).wait();
    } catch (err) {
        throw new IOError(err.message + ": " + filepath);
    }
}

function repr(object) {
}


GLOBAL.extend = extend;
GLOBAL.read = read;
GLOBAL.repr = repr;

GLOBAL.inspect = sys.p;
GLOBAL.print = sys.print;
GLOBAL.puts = sys.puts;
