// Released into the Public Domain by tav <tav@espians.com>

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
    var i = 0,
        k;
    for (k in object) {
        if (object.hasOwnProperty(k)) {
            i += k;
        }
    }
    return "";
}

GLOBAL.IOError = IOError;
GLOBAL.ValueError = ValueError;

GLOBAL.extend = extend;
GLOBAL.read = read;
GLOBAL.repr = repr;

GLOBAL.inspect = sys.p;
GLOBAL.print = sys.print;
GLOBAL.puts = sys.puts;
