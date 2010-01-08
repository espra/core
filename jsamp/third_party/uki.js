/*!
 * uki JavaScript Library v0.0.5
 * Licensed under the MIT license
 *
 * Copyright (c) 2009 Vladimir Kolesnikov
 *
 * Parts of code derived from jQuery JavaScript Library v1.3.2
 * Copyright (c) 2009 John Resig
 */
(function() {
/**
 * Global uki constanst, for speed optimization and better merging
 */
var root = this,
    doc  = document,
    nav = navigator,
    ua  = nav.userAgent,
    expando = 'uki' + (+new Date),

    PROTOTYPE = 'prototype',
    MAX = Math.max,
    MIN = Math.min,
    FLOOR = Math.floor,
    CEIL = Math.ceil,

    PX = 'px';

/**
 * Shortcut access to uki.build, uki.Selector.find and uki.Collection constructor
 * uki('#id') is also a shortcut for search by id
 *
 * @param {String|uki.view.Base|Object|Array.<uki.view.Base>} val
 * @param {Array.<uki.view.Base>=} optional context for selector
 * @class
 * @name uki
 * @return {uki.Collection}
 */
root.uki = function(val, context) {
    if (typeof val == 'string') {
        var m = val.match(/^#((?:[\w\u00c0-\uFFFF_-]|\\.)+)$/),
            e = m && uki._ids[m[1]];
        if (m && !context) {
            return new uki.Collection( e ? [e] : [] );
        }
        return uki.find(val, context);
    }
    if (val.length === undefined) val = [val];
    if (val.length > 0 && uki.isFunction(val[0].typeName)) return new uki.Collection(val);
    return uki.build(val);
};

/**
 * @type string
 * @field
 */
uki.version = '0.0.5';

/**
 * Empty function
 * @type function():boolean
 */
uki.F = function() { return false; };
uki._ids = {};

uki.registerId = function(comp) {
    uki._ids[uki.attr(comp, 'id')] = comp;
};
uki.unregisterId = function(comp) {
    uki._ids[uki.attr(comp, 'id')] = undefined;
};

var toString = Object[PROTOTYPE].toString;

var utils =
/**
 * Utility functions. Can be called both as uki.utils.function or uki.function
 * @namespace
 */
uki.utils = {

    /**
     * <p>Sets or retrieves attribute on an object. </p>
     * <p>If target has function with attr it will be called target[attr](value=)
     * If no function present attribute will be set/get directly: target[attr] = value or return target[attr]</p>
     *
     * @example
     *   uki.attr(view, 'name', 'funny') // sets name to funny on view
     *   uki.attr(view, 'id') // gets id attribute of view
     *
     * @param {object} target
     * @param {string} attr Attribute name
     * @param {object=} value Value to set
     * @returns {object} target if value is being set, retrieved value otherwise
     */
    attr: function(target, attr, value) {
        if (value !== undefined) {
            if (utils.isFunction(target[attr])) {
                target[attr](value);
            } else {
                target[attr] = value;
            }
            return target;
        } else {
            if (utils.isFunction(target[attr])) {
                return target[attr]();
            } else {
                return target[attr];
            }
        }
    },

    /**
     * Checks if obj is a function
     *
     * @param {object} object Object to check
     * @returns {boolean}
     */
	isFunction: function( obj ) {
		return toString.call(obj) === "[object Function]";
	},

    /**
     * Checks if obj is an Array
     *
     * @param {object} object Object to check
     * @returns {boolean}
     */
    isArray: function( obj ) {
        return toString.call(obj) === "[object Array]";
    },

    /**
     * Trims the string
     *
     * @param {string} text
     * @returns {string} trimmed text
     */
    trim: function( text ) {
        return (text || "").replace( /^\s+|\s+$/g, "" );
    },

    /**
     * Converts unsafe symbols to entities
     *
     * @param {string} html
     * @returns {string} escaped html
     */
    escapeHTML: function( html ) {
        var trans = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#x27;'
        };
        return html.replace(/[&<>\"\']/g, function(c) { return trans[c]; });
    },

    /**
     * Iterates through all non empty values of object or an Array
     *
     * @param {object|Array} object Object to iterate through
     * @param {function(number, object):boolean} callback Called for every item, may return false to stop iteration
     * @param {object} context Context in which callback should called. If not specified context will be set to
     *                         current item
     * @returns {object}
     */
    each: function( object, callback, context ) {
        var name, i = 0, length = object.length;

        if ( length === undefined ) {
            for ( name in object ) {
                if ( !name || object[ name ] === undefined || !object.hasOwnProperty(name) ) continue;
                if ( callback.call( context || object[ name ], name, object[ name ] ) === false ) { break; }
            }
        } else {
            for ( var value = object[0]; i < length && callback.call( context || value, i, value ) !== false; value = object[++i] ){}
        }
        return object;
    },

    /**
     * <p>Checks if elem is in array</p>
     *
     * @param {object} elem
     * @param {object} array
     * @returns {boolean}
     */
    inArray: function( elem, array ) {
        for ( var i = 0, length = array.length; i < length; i++ ) {
            if ( array[ i ] === elem ) { return i; }
        }

        return -1;
    },

    /**
     * <p>Returns unique elements in array</p>
     *
     * @param {Array} array
     * @returns {Array}
     */
    unique: function( array ) {
        var ret = [], done = {};

        try {

            for ( var i = 0, length = array.length; i < length; i++ ) {
                var id = array[ i ];

                if ( !done[ id ] ) {
                    done[ id ] = true;
                    ret.push( array[ i ] );
                }
            }

        } catch( e ) {
            ret = array;
        }

        return ret;
    },

    /**
     * <p>Searches for all items matchign given criteria</p>
     *
     * @param {Array} elems Element to search through
     * @param {function(object, number)} callback Returns true for every matched element
     * @returns {Array} matched elements
     */
    grep: function( elems, callback ) {
        var ret = [];

        for ( var i = 0, length = elems.length; i < length; i++ ) {
            if ( callback( elems[ i ], i ) ) { ret.push( elems[ i ] ); }
        }

        return ret;
    },

    /**
     * <p>Maps elements passing them to callback</p>
     * @example
     *   x = uki.map([1, 2, 3], function(item) { return -item });
     *
     * @param {Array} elems Elements to map
     * @param {function(object, number)} mapping function
     * @param {object} context Context in which callback should called. If not specified context will be set to
     *                         current item
     * @returns {Array} mapped values
     */
    map: function( elems, callback, context ) {
        var ret = [],
            mapper = utils.isFunction(callback) ? callback :
                     function(e) { return utils.attr(e, callback); };

        for ( var i = 0, length = elems.length; i < length; i++ ) {
            var value = mapper.call( context || elems[ i ], elems[ i ], i );

            if ( value != null ) { ret[ ret.length ] = value; }
        }

        return ret;
    },

    /**
     * <p>Reduces array</p>
     * @example
     *   x = uki.reduce(1, [1, 2, 3], function(p, x) { return p * x}) // calculates product
     *
     * @param {object} initial Initial value
     * @param {Array} elems Elements to reduce
     * @param {function(object, number)} reduce function
     * @param {object} context Context in which callback should called. If not specified context will be set to
     *                         current item
     * @returns {object}
     */
    reduce: function( initial, elems, callback, context ) {
        for ( var i = 0, length = elems.length; i < length; i++ ) {
            initial = callback.call( context || elems[ i ], initial, elems[ i ], i );
        }
        return initial;
    },

    /**
     * <p>Copies properties from one object to another</p>
     * @example
     *   uki.extend(x, { width: 13, height: 14 }) // sets x.width = 13, x.height = 14
     *   options = uki.extend({}, defaultOptions, options)
     *
     * @param {object} target Object to copy properties into
     * @param {...object} sources Objects to take properties from
     * @returns Describe what it returns
     */
    extend: function() {
        var target = arguments[0] || {}, i = 1, length = arguments.length, options;

        for ( ; i < length; i++ ) {
            if ( (options = arguments[i]) != null ) {

                for ( var name in options ) {
                    var copy = options[ name ];

                    if ( copy !== undefined ) {
                        target[ name ] = copy;
                    }

                }
            }
        }

        return target;
    },

    /**
     * <p>Creates a new class inherited from base classes. Init function is used as constructor</p>
     * @example
     *   baseClass = uki.newClass({
     *      init: function() { this.x = 3 }
     *   });
     *
     *   childClass = uki.newClass(baseClass, {
     *      getSqrt: function() { return this.x*this.x }
     *   });
     *
     * @param {object=} superClass If superClass has prototype "real" protototype base inheritance is used,
     *                             otherwise superClass properties ar simply copied to newClass prototype
     * @param {Array.<object>=} mixins
     * @param {object} methods
     * @returns Describe what it returns
     */
    newClass: function(/* [[superClass], mixin1, mixin2, ..] */ methods) {
        var klass = function() {
                this.init.apply(this, arguments);
            };

        var inheritance, i, startFrom = 0;

        if (arguments.length > 1) {
            if (arguments[0][PROTOTYPE]) { // real inheritance
                /** @ignore */
                inheritance = function() {};
                inheritance[PROTOTYPE] = arguments[0][PROTOTYPE];
                klass[PROTOTYPE] = new inheritance();
                startFrom = 1;
            }
        }
        for (i=startFrom; i < arguments.length; i++) {
            utils.extend(klass[PROTOTYPE], arguments[i]);
        };
        return klass;
    },

    /**
     * <p>Creates default uki property function</p>
     * <p>If value is given to this function it sets property to value
     * If no arguments given than function returns current property value</p>
     *
     * <p>Optional setter can be given. In this case setter will be called instead
     * of simple this[field] = value</p>
     *
     * <p>If used as setter function returns self</p>
     *
     * @example
     *   x.width = uki.newProperty('_width');
     *   x.width(12); // x._width = 12
     *   x.width();   // return 12
     *
     * @param {string} field Field name
     * @param {function(object)=} setter
     * @returns {function(object=):object}
     */
    newProp: function(field, setter) {
        return function(value) {
            if (value === undefined) return this[field];
            if (setter) { setter.call(this, value); } else { this[field] = value; };
            return this;
        };
    },

    /**
     * <p>Adds several properties (uki.newProp) to a given object.</p>
     * <p>Field name equals to '_' + property name</p>
     *
     * @example
     *   uki.addProps(x, ['width', 'height'])
     *
     * @param {object} proto Object to add properties to
     * @param {Array.<string>} props Property names
     */
    addProps: function(proto, props) {
        utils.each(props, function() { proto[this] = utils.newProp('_' + this); });
    },

    delegateProp: function(proto, name, target) {
        var propName = '_' + name;
        proto[name] = function(value) {
            if (this[target]) return utils.attr(this[target], name, value);
            if (value === undefined) return this[propName];
            this[propName] = value;
            return this;
        };
    }
};

utils.extend(uki, utils);/**
 * @namespace
 * @name uki.geometry
 */
var geometry = uki.geometry = {};


/**
 * Point with x and y properties
 *
 * @author voloko
 * @name uki.geometry.Point
 * @constructor
 *
 * @param {Integer=} x defaults to 0
 * @param {Integer=} y defaults to 0
 */
var Point = geometry.Point = function(x, y) {
    this.x = x || 0.0;
    this.y = y || 0.0;
};

Point[PROTOTYPE] = /** @lends uki.geometry.Point.prototype */ {

    /**
     * Converts to "100 50" string
     *
     * @this {uki.geometry.Point}
     * @return {string}
     */
    toString: function() {
        return this.x + ' ' + this.y;
    },

    /**
     * Creates a new Point with the same properites
     *
     * @this {uki.geometry.Point}
     * @return {uki.geometry.Point}
     */
    clone: function() {
        return new Point(this.x, this.y);
    },

    /**
     * Checks if this equals to another Point
     *
     * @param {uki.geometry.Point} point Point to compare with
     * @this {uki.geometry.Point}
     * @return {boolean}
     */
    eq: function(point) {
        return this.x == point.x && this.y == point.y;
    },

    /**
     * Moves point by x, y
     *
     * @this {uki.geometry.Point}
     * @return {uki.geometry.Point} self
     */
    offset: function(x, y) {
        this.x += x;
        this.y += y;
        return this;
    },

    constructor: Point
};

/**
 * Creates point from "x y" string
 * x and y may have optional units=px specified: Point.fromString("20% 10px")
 *
 * @name uki.geometry.Point.fromString
 * @function
 *
 * @param {string} string String representation of point
 * @param {uki.geometry.Size=} relative Relative size to calculate % values
 *
 * @returns {uki.geometry.Point} created point
 */
Point.fromString = function(string, relative) {
    var parts = string.split(/\s+/);
    return new Point( unitsToPx(parts[0], relative && relative.width), unitsToPx(parts[1], relative && relative.height) );
};


/**
 * Size with width and height properties
 *
 * @param {number=} width defaults to 0
 * @param {number=} height defaults to 0
 * @name uki.geometry.Size
 * @constructor
 */
var Size = geometry.Size = function(width, height) {
    this.width  = width ||  0.0;
    this.height = height || 0.0;
};

Size[PROTOTYPE] = /** @lends uki.geometry.Size.prototype */ {
    /**
     * Converts size to "300 100" string
     *
     * @this {uki.geometry.Size}
     * @return {string}
     */
    toString: function() {
        return this.width + ' ' + this.height;
    },

    /**
     * Creates a new Size with same properties
     *
     * @this {uki.geometry.Size}
     * @return {uki.geometry.Size} new Size
     */
    clone: function() {
        return new Size(this.width, this.height);
    },

    /**
     * Checks if this equals to another Size
     *
     * @param {uki.geometry.Size} size Size to compare with
     * @this {uki.geometry.Size}
     * @return {boolean}
     */
    eq: function(size) {
        return this.width == size.width && this.height == size.height;
    },

    /**
     * Checks if this size has non-positive width or height
     *
     * @this {uki.geometry.Size}
     * @return {boolean}
     */
    empty: function() {
        return this.width <= 0 || this.height <= 0;
    },

    constructor: Size
};

/**
 * Creates size from "width height" string
 * x and y may have optional units=px specified: Size.fromString("20% 10px")
 *
 * @name uki.geometry.Size.fromString
 * @function
 *
 * @param {string} string String representation of size
 * @param {uki.geometry.Size=} relative Relative size to calculate % values
 *
 * @returns {uki.geometry.Size} created size
 */
Size.fromString = function(string, relative) {
    var parts = string.split(/\s+/);
    return new Size( unitsToPx(parts[0], relative && relative.width), unitsToPx(parts[1], relative && relative.height) );
};

/**
 * Creates size from different representations
 * - if no params given returns null
 * - if uki.geometry.Size given returns it
 * - if "200 300" string converts it to size
 * - if two params given creates size from them
 *
 * @name uki.geometry.Size.create
 * @function
 *
 * @param {...string|number|uki.geometry.Size} var_args Size representation
 *
 * @returns {uki.geometry.Size} created size
 */
Size.create = function(a1, a2) {
    if (a1 === undefined) return null;
    if (a1.width !== undefined) return a1;
    if (/\S+\s+\S+/.test(a1 + '')) return Size.fromString(a1, a2);
    return new Size(a1, a2);
};


/**
 * Rectangle with x, y, width and height properties
 * May be used as uki.geometry.Point or uki.geometry.Size
 * - if 4 arguments given creates size with x,y,width,height set to the given arguments
 * - if 2 number arguments given creates size with x = y = 0 and width and height set
 *   set to the given arguments
 * - if a Point and a Size given creates rect with point as an origin and given size
 *
 * @param {...number|uki.geometry.Point|uki.geometry.Size} var_args
 * @name uki.geometry.Rect
 * @constructor
 */
var Rect = geometry.Rect = function(a1, a2, a3, a4) {
    if (a3 !== undefined) {
        this.x      = a1;
        this.y      = a2;
        this.width  = a3;
        this.height = a4;
    } else if (a1 === undefined || a1.x === undefined) {
        this.x      = 0;
        this.y      = 0;
        this.width  = a1 || 0;
        this.height = a2 || 0;
    } else {
        this.x      = a1 ? a1.x      : 0;
        this.y      = a1 ? a1.y      : 0;
        this.width  = a2 ? a2.width  : 0;
        this.height = a2 ? a2.height : 0;
    }
};

Rect[PROTOTYPE] = /** @lends uki.geometry.Rect.prototype */ {
    /**
     * Converts Rect to "x y width height" string
     *
     * @this {uki.geometry.Rect}
     * @returns {string}
     */
    toString: function() {
        return [this.x, this.y, this.width, this.height].join(' ');
    },

    /**
     * Converts Rect to "x y maxX maxY" string
     *
     * @this {uki.geometry.Rect}
     * @returns {string}
     */
    toCoordsString: function() {
        return [this.x, this.y, this.maxX(), this.maxY()].join(' ');
    },

    /**
     * Creates a new Rect with same properties
     *
     * @this {uki.geometry.Size}
     * @return {uki.geometry.Size} new Size
     */
    clone: function() {
        return new Rect(this.x, this.y, this.width, this.height);
    },

    /**
     * Equals to .x
     *
     * @this {uki.geometry.Rect}
     * @returns {number}
     */
    minX: function() {
        return this.x;
    },

    /**
     * Equals to x + width
     *
     * @this {uki.geometry.Rect}
     * @returns {number}
     */
    maxX: function() {
        return this.x + this.width;
    },

    /**
     * Point between minX and maxX
     *
     * @this {uki.geometry.Rect}
     * @returns {number}
     */
    midX: function() {
        return this.x + this.width / 2.0;
    },

    /**
     * Equals to .y
     *
     * @this {uki.geometry.Rect}
     * @returns {number}
     */
    minY: function() {
        return this.y;
    },

    /**
     * Point between minY and maxY
     *
     * @this {uki.geometry.Rect}
     * @returns {number}
     */
    midY: function() {
        return this.y + this.height / 2.0;
    },

    /**
     * Equals to y + height
     *
     * @this {uki.geometry.Rect}
     * @returns {number}
     */
    maxY: function() {
        return this.y + this.height;
    },

    /**
     * Moves origin to 0,0 point
     *
     * @this {uki.geometry.Rect}
     * @returns {uki.geometry.Point} self
     */
    normalize: function() {
        this.x = this.y = 0;
        return this;
    },

    /**
     * Checks if this rect has non-positive width or height
     *
     * @this {uki.geometry.Rect}
     * @return {boolean}
     */
    empty: Size[PROTOTYPE].empty,

    /**
     * Checks if this equals to another Rect
     *
     * @param {uki.geometry.Rect} rect Rect to compare with
     * @this {uki.geometry.Rect}
     * @return {boolean}
     */
    eq: function(rect) {
        return rect && this.x == rect.x && this.y == rect.y && this.height == rect.height && this.width == rect.width;
    },

    /**
     * Insets size with dx and dy
     *
     * @param {number} dx
     * @param {number} dy
     * @this {uki.geometry.Rect}
     * @returns {uki.geometry.Rect} sefl
     */
    inset: function(dx, dy) {
        this.x += dx;
        this.y += dy;
        this.width -= dx*2.0;
        this.height -= dy*2.0;
        return this;
    },

    /**
     * Moves origin point by x, y
     *
     * @this {uki.geometry.Rect}
     * @return {uki.geometry.Rect} self
     */
    offset: Point[PROTOTYPE].offset,

    /**
     * Intersects this with given rect
     *
     * @this {uki.geometry.Rect}
     * @param {uki.geometry.Rect} rect Rect to intersect with
     * @returns {uki.geometry.Rect} intersection
     */
    intersection: function(rect) {
        var origin = new Point(
                MAX(this.x, rect.x),
                MAX(this.y, rect.y)
            ),
            size = new Size(
                MIN(this.maxX(), rect.maxX()) - origin.x,
                MIN(this.maxY(), rect.maxY()) - origin.y
            );
        return size.empty() ? new Rect() : new Rect(origin, size);
    },

    /**
     * Union rect of this and given rect
     *
     * @this {uki.geometry.Rect}
     * @param {uki.geometry.Rect} rect
     * @returns {uki.geometry.Rect} union
     */
    union: function(rect) {
        return Rect.fromCoords(
            MIN(this.x, rect.x),
            MIN(this.y, rect.y),
            MAX(this.maxX(), rect.maxX()),
            MAX(this.maxY(), rect.maxY())
        );
    },

    /**
     * Checks if point is within this
     *
     * @this {uki.geometry.Rect}
     * @param {uki.geometry.Point} point
     * @returns {boolean}
     */
    containsPoint: function(point) {
        return point.x >= this.minX() &&
               point.x <= this.maxX() &&
               point.y >= this.minY() &&
               point.y <= this.maxY();
    },

    /**
     * Checks if this contains given rect
     *
     * @this {uki.geometry.Rect}
     * @param {uki.geometry.Rect} rect
     * @returns {boolean}
     */
    containsRect: function(rect) {
        return this.eq(this.union(rect));
    },

    constructor: Rect
};

Rect[PROTOTYPE].left = Rect[PROTOTYPE].minX;
Rect[PROTOTYPE].top  = Rect[PROTOTYPE].minY;

/**
 * Creates Rect from minX, minY, maxX, maxY
 *
 * @name uki.geometry.Rect.fromCoords
 * @function
 *
 * @param {number} minX
 * @param {number} maxX
 * @param {number} minY
 * @param {number} maxY
 * @returns {uki.geometry.Rect}
 */
Rect.fromCoords = function(minX, minY, maxX, maxY) {
    if (maxX === undefined) {
        return new Rect(
            minX.x,
            minX.y,
            minY.x - minX.x,
            minY.y - minX.y
        );
    }
    return new Rect(minX, minY, maxX - minX, maxY - minY);
};

/**
 * Creates Rect from "minX minY maxX maxY" string
 *
 * @name uki.geometry.Rect.fromCoordsString
 * @function
 *
 * @param {string} string
 * @param {uki.geometry.Size=} relative Relative size to calculate % values
 * @returns {uki.geometry.Rect}
 */
Rect.fromCoordsString = function(string, relative) {
    var parts = string.split(/\s+/);
    return Rect.fromCoords(
        unitsToPx(parts[0], relative && relative.width),
        unitsToPx(parts[1], relative && relative.height),
        unitsToPx(parts[2], relative && relative.width),
        unitsToPx(parts[3], relative && relative.height)
    ) ;
};

/**
 * Creates Rect from "x y width height" or "width height" string
 *
 * @name uki.geometry.Rect.fromString
 * @function
 *
 * @param {string} string
 * @param {uki.geometry.Size=} relative Relative size to calculate % values
 * @returns {uki.geometry.Rect}
 */
Rect.fromString = function(string, relative) {
    var parts = string.split(/\s+/);

    if (parts.length > 2) return new Rect(
        unitsToPx(parts[0], relative && relative.width),
        unitsToPx(parts[1], relative && relative.height),
        unitsToPx(parts[2], relative && relative.width),
        unitsToPx(parts[3], relative && relative.height)
    );
    return new Rect(
        unitsToPx(parts[0], relative && relative.width),
        unitsToPx(parts[1], relative && relative.height)
    ) ;
};

/**
 * Creates rect from different representations
 * - if no params given returns null
 * - if uki.geometry.Rect given returns it
 * - if "200 300" or "0 10 200 300" string converts it to rect
 * - if two or four params given creates rect from them
 *
 * @name uki.geometry.Rect.create
 * @function
 *
 * @param {...string|number|uki.geometry.Rect} var_args Rect representation
 *
 * @returns {uki.geometry.Rect} created size
 */
Rect.create = function(a1, a2, a3, a4) {
    if (a1 === undefined) return null;
    if (a1.x !== undefined) return a1;
    if (/\S+\s+\S+/.test(a1 + '')) return Rect.fromString(a1, a2);
    if (a3 === undefined) return new Rect(a1, a2);
    return new Rect(a1, a2, a3, a4);
};


/**
 * Inset with top, right, bottom and left properties
 * - if no params given top = right = bottom = left = 0
 * - if two params given top = bottom and right = left
 *
 * @param {number=} top
 * @param {number=} right
 * @param {number=} bottom
 * @param {number=} left
 *
 * @name uki.geometry.Inset
 * @constructor
 */
var Inset = geometry.Inset = function(top, right, bottom, left) {
    this.top    = top   || 0;
    this.right  = right || 0;
    this.bottom = bottom === undefined ? this.top : bottom;
    this.left   = left === undefined ? this.right : left;
};

Inset[PROTOTYPE] = /** @lends uki.geometry.Inset.prototype */ {

    /**
     * Converts Inset to "top right bottom left" string
     *
     * @returns {string}
     */
    toString: function() {
        return [this.top, this.right, this.bottom, this.left].join(' ');
    },

    /**
     * Creates a new Inset with same properties
     *
     * @this {uki.geometry.Inset}
     * @return {uki.geometry.Inset} new Inset
     */
    clone: function() {
        return new Inset(this.top, this.right, this.bottom, this.left);
    },

    /**
     * left + right
     *
     * @this {uki.geometry.Inset}
     * @return {number}
     */
    width: function() {
        return this.left + this.right;
    },

    /**
     * top + bottom
     *
     * @this {uki.geometry.Inset}
     * @return {number}
     */
    height: function() {
        return this.top + this.bottom;
    },

    /**
     * True if any property < 0
     *
     * @this {uki.geometry.Inset}
     * @return {boolean}
     */
    negative: function() {
        return this.top < 0 || this.left < 0 || this.right < 0 || this.bottom < 0;
    },

    /**
     * True if all properties = 0
     *
     * @this {uki.geometry.Inset}
     * @return {boolean}
     */
    empty: function() {
        return !this.top && !this.left && !this.right && !this.bottom;
    }
};

/**
 * Creates Rect from "top right bottom left" or "top right" string
 *
 * @name uki.geometry.Inset.fromString
 * @function
 *
 * @param {string} string
 * @param {uki.geometry.Size=} relative Relative size to calculate % values
 * @returns {uki.geometry.Inset}
 */
Inset.fromString = function(string, relative) {
    var parts = string.split(/\s+/);
    if (parts.length < 3) parts[2] = parts[0];
    if (parts.length < 4) parts[3] = parts[1];

    return new Inset(
        unitsToPx(parts[0], relative),
        unitsToPx(parts[1], relative),
        unitsToPx(parts[2], relative),
        unitsToPx(parts[3], relative)
    );
};

/**
 * Creates rect from different representations
 * - if no params given returns null
 * - if uki.geometry.Inset given returns it
 * - if "200 300" or "0 10 200 300" string converts it to inset
 * - if two or four params given creates inset from them
 *
 * @name uki.geometry.Inset.create
 * @function
 *
 * @param {...string|number|uki.geometry.Inset} var_args Rect representation
 *
 * @returns {uki.geometry.Inset} created inset
 */
Inset.create = function(a1, a2, a3, a4) {
    if (a1 === undefined) return null;
    if (a1.top !== undefined) return a1;
    if (/\S+\s+\S+/.test(a1 + '')) return Inset.fromString(a1, a2);
    if (a3 === undefined) return new Inset(a1, a2);
    return new Inset(a1, a2, a3, a4);
};


function unitsToPx (units, relative) {
    var m = (units + '').match(/([-0-9\.]+)(\S*)/),
        v = parseFloat(m[1], 10),
        u = (m[2] || '').toLowerCase();

    if (u) {
        // if (u == 'px') v = v;
        if (u == '%' && relative) v *= relative / 100;
    }
    if (v < 0 && relative) v = relative + v;
    return v;
}

/**
 * Basic utils to work with the dom tree
 * @namespace
 * @author voloko
 */
uki.dom = {
    guid: 1,

    /**
     * Convinience wrapper around document.createElement
     * Creates dom element with given tagName, cssText and innerHTML
     *
     * @param {string} tagName
     * @param {string=} cssText
     * @param {string=} innerHTML
     * @returns {Element} created element
     */
    createElement: function(tagName, cssText, innerHTML) {
        var e = doc.createElement(tagName);
        if (cssText) e.style.cssText = cssText;
        if (innerHTML) e.innerHTML = innerHTML;
        e[expando] = uki.dom.guid++;
        return e;
    },

    /**
     * Adds a probe element to page dom tree, callbacks, removes the element
     *
     * @param {Element} dom Probing dom element
     * @param {function(Element)} callback
     */
    probe: function(dom, callback) {
        var target = doc.body;
        target.appendChild(dom);
        callback(dom);
        target.removeChild(dom);
    },

    /**
     * Assigns layout style properties to an element
     *
     * @param {CSSStyleDeclaration} style Target declaration
     * @param {object} layout Properties to assign
     * @param {object=} prevLayout If given assigns only differenct between layout and prevLayout
     */
    layout: function(style, layout, prevLayout) {
        prevLayout = prevLayout || {};
        if (prevLayout.left   != layout.left)   style.left   = layout.left + PX;
        if (prevLayout.top    != layout.top)    style.top    = layout.top + PX;
        if (prevLayout.right  != layout.right)  style.right  = layout.right + PX;
        if (prevLayout.bottom != layout.bottom) style.bottom = layout.bottom + PX;
        if (prevLayout.width  != layout.width)  style.width  = MAX(layout.width, 0) + PX;
        if (prevLayout.height != layout.height) style.height = MAX(layout.height, 0) + PX;
        return layout;
    },

    /**
     * Computed style for a give element
     *
     * @param {Element} el
     * @returns {CSSStyleDeclaration} style declaratioin
     */
    computedStyle: function(el) {
        if (doc && doc.defaultView && doc.defaultView.getComputedStyle) {
            return doc.defaultView.getComputedStyle( el, null );
        } else if (el.currentStyle) {
            return el.currentStyle;
        }
    },

    /**
     * Checks if parent contains child
     *
     * @param {Element} parent
     * @param {Element} child
     * @return {Boolean}
     */
    contains: function(parent, child) {
        try {
            if (parent.contains) return parent.contains(child);
            if (parent.compareDocumentPosition) return parent.compareDocumentPosition(child) & 16;
        } catch (e) {}
        while ( child && child != parent ) {
            try { child = child.parentNode } catch(e) { child = null };
        }
        return parent == child;
    }

};

uki.each(['createElement'], function(i, name) {
    uki[name] = uki.dom[name];
});/** @namespace */
uki.view = {};/**
 * @class
 */
uki.view.Observable = {
    // dom: function() {
    //     return null; // should implement
    // },

    bind: function(name, callback) {
        uki.each(name.split(' '), function(i, name) {
            if (!this._bound(name)) this._bindToDom(name);
            this._observersFor(name).push(callback);
        }, this);
    },

    unbind: function(name, callback) {
        uki.each(name.split(' '), function(i, name) {
            this._observers[name] = uki.grep(this._observers[name], function(observer) {
                return observer != callback;
            });
            if (this._observers[name].length == 0) {
                this._unbindFromDom(name);
            }
        }, this);
    },

    trigger: function(name/*, data1, data2*/) {
        var attrs = Array[PROTOTYPE].slice.call(arguments, 1);
        uki.each(this._observersFor(name, true), function(i, callback) {
            callback.apply(this, attrs);
        }, this);
    },

    _unbindFromDom: function(name, target) {
        if (!this._domHander || !this._eventTargets[name]) return;
        uki.dom.unbind(this._eventTargets[name], name, this._domHander);
    },

    _bindToDom: function(name, target) {
        var _this = this;
        this._domHander = this._domHander || function(e) {
            _this.trigger(e.type, {domEvent: e, source: _this});
        };
        this._eventTargets = this._eventTargets || {};
        this._eventTargets[name] = target || this.dom();
        uki.dom.bind(this._eventTargets[name], name, this._domHander);
        return true;
    },

    _bound: function(name) {
        return this._observers && this._observers[name];
    },

    _observersFor: function(name, skipCreate) {
        if (skipCreate && (!this._observers || !this._observers[name])) return [];
        if (!this._observers) this._observers = {};
        if (!this._observers[name]) this._observers[name] = [];
        return this._observers[name];
    }
};(function() {

    var self = uki.Attachment = uki.newClass(uki.view.Observable, /** @lends uki.Attachment.prototype */ {
        /**
         * Attachment serves as a connection between a uki view and a dom container.
         * It notifies its view with parentResized on window resize.
         * Attachment supports part of uki.view.Base API like #domForChild or #rectForChild
         *
         * @param {Element} dom Container element
         * @param {uki.view.Base} view Attached view
         * @param {uki.geometry.Rect} rect Initial size
         *
         * @see uki.view.Base#parentResized
         * @name uki.Attachment
         * @base uki.view.Observable
         * @constructor
         */
        init: function( dom, view, rect ) {
            uki.initNativeLayout();

            this._dom     = dom = dom || root;
            this._view    = view;
            this._rect = Rect.create(rect) || this.rect();

            uki.dom.offset.initialize();

            view.parent(this);

            if (dom != root && dom.tagName != 'BODY') {
                var computedStyle = dom.runtimeStyle || dom.ownerDocument.defaultView.getComputedStyle(dom, null);
                if (!computedStyle.position || computedStyle.position == 'static') dom.style.position = 'relative';
            }
            self.register(this);

            this.layout();
        },

        /**
         * Returns document.body if attached to window. Otherwise returns dom
         * uki.view.Base api
         *
         * @type Element
         */
        domForChild: function() {
            return this._dom === root ? doc.body : this._dom;
        },

        /**
         * uki.view.Base api
         *
         * @type uki.geometry.Rect
         */
        rectForChild: function(child) {
            return this.rect();
        },

        /**
         * uki.view.Base api
         */
        scroll: function() {
            // TODO: support window scrolling
        },

        /**
         * uki.view.Base api
         */
        scrollTop: function() {
            // TODO: support window scrolling
            return this._dom.scrollTop || 0;
        },

        /**
         * uki.view.Base api
         */
        scrollLeft: function() {
            // TODO: support window scrolling
            return this._dom.scrollLeft || 0;
        },

        /**
         * uki.view.Base api
         */
        parent: function() {
            return null;
        },

        /**
         * On window resize resizes and laysout its child view
         */
        layout: function() {
            var oldRect = this._rect;

            // if (rect.eq(this._rect)) return;
            this._rect = this.rect();

            this._view.parentResized(oldRect, this._rect);
            if (this._view._needsLayout) this._view.layout();
            this.trigger('layout', {source: this, rect: this._rect});
        },

        /**
         * @return {Element} Container dom
         */
        dom: function() {
            return this._dom;
        },

        /**
         * @return {Element} Child view
         */
        view: function() {
            return this._view;
        },

        /**
         * @private
         * @return {uki.geometry.Rect} Size of the container
         */
        rect: function() {
            var width = this._dom === root || this._dom === doc.body ? MAX(getRootElement().clientWidth, this._dom.offsetWidth || 0) : this._dom.offsetWidth,
                height = this._dom === root || this._dom === doc.body ? MAX(getRootElement().clientHeight, this._dom.offsetHeight || 0) : this._dom.offsetHeight;

            return new Rect(width, height);
        }
    });

    function getRootElement() {
        return doc.compatMode == "CSS1Compat" && doc.documentElement || doc.body;
    }

    self.instances = [];

    self.register = function(a) {
        if (self.instances.length == 0) {
            uki.dom.bind(root, 'resize', function() {
                uki.each(self.instances, function() { this.layout(); });
            });
        }
        self.instances.push(a);
    };

    self.childViews = function() {
        return uki.map(self.instances, 'view');
    };

    uki.top = function() {
        return [self];
    };
})();
(function() {

    /**
     * <p>Collection performs group operations on uki.view objects.</p>
     * <p>Behaves much like result jQuery(dom nodes).
     * Most methods are chainable like .attr('text', 'somevalue').bind('click', function() { ... })</p>
     *
     * <p>Its easier to call uki([view1, view2]) or uki('selector') instead of creating collection directly</p>
     *
     * @author voloko
     * @constructor
     * @class
     */
    uki.Collection = function( elems ) {
        this.length = 0;
    	Array[PROTOTYPE].push.apply( this, elems );
    };

    /** @type uki.Collection.prototype */
    var self = uki.fn = uki.Collection[PROTOTYPE] = {};

    /**
     * Iterates trough all items within itself
     *
     * @name uki.Collection#each
     * @function
     *
     * @param {function(this:uki.view.Base, number, uki.view.Base)} callback Callback to call for every item
     * @returns {uki.view.Collection} self
     */
    self.each = function( callback ) {
        uki.each( this, callback );
        return this;
    };

    /**
     * Creates a new uki.Collection populated with found items
     *
     * @name uki.Collection#grep
     * @function
     *
     * @param {function(uki.view.Base, number):boolean} callback Callback to call for every item
     * @returns {uki.view.Collection} created collection
     */
    self.grep = function( callback ) {
        return new uki.Collection( uki.grep(this, callback) );
    };

    /**
     * Sets an attribute on all views or gets the value of the attribute on the first view
     *
     * @example
     * c.attr('text', 'my text') // sets text to 'my text' on all collection views
     * c.attr('name') // gets name attribute on the first view
     *
     * @name uki.Collection#attr
     * @function
     *
     * @param {string} name Name of the attribute
     * @param {object=} value Value to set
     * @returns {uki.view.Collection|Object} Self or attribute value
     */
    self.attr = function( name, value ) {
        if (value !== undefined) {
            this.each(function() {
                uki.attr( this, name, value );
            });
            return this;
        } else {
            return this[0] ? uki.attr( this[0], name ) : '';
        }
    };

    /**
     * Finds views within collection context
     * @example
     * c.find('Button')
     *
     * @name uki.Collection#find
     * @function
     *
     * @param {string} selector
     * @returns {uki.view.Collection} Collection of found items
     */
    self.find = function( selector ) {
        return uki.find( selector, this );
    };

    /**
     * Attaches all child views to dom container
     *
     * @name uki.Collection#attachTo
     * @function
     *
     * @param {Element} dom Container dom element
     * @param {uki.geometry.Rect} rect Default size
     * @returns {uki.view.Collection} self
     */
    self.attachTo = function( dom, rect ) {
        this.each(function() {
            new uki.Attachment( dom, this, rect );
        });
        return this;
    };

    /**
     * Appends views to the first item in collection
     *
     * @name uki.Collection#append
     * @function
     *
     * @param {Array.<uki.view.Base>} views Views to append
     * @returns {uki.view.Collection} self
     */
    self.append = function( views ) {
        if (!this[0]) return this;
        views = views.length !== undefined ? views : [views];
        for (var i=0; i < views.length; i++) {
            this[0].appendChild(views[i]);
        };
        return this;
    };

    uki.each(['parent'], function(i, name) {
        self[name] = function() {
            return new uki.Collection( uki.map(this, name) );
        };
    });

    uki.each(['scrollableParent'], function(i, name) {
       self[name] = function() {
            return new uki.Collection( uki.map(this, function(c) { return uki.view[name](c); }) );
        };
    });

    uki.each(('bind,unload,trigger,layout,appendChild,removeChild,addRow,removeRow,' +
        'resizeToContents,toggle').split(','), function(i, name) {
        self[name] = function() {
            for (var i=0; i < this.length; i++) {
                this[i][name].apply(this[i], arguments);
            };
            return this;
        };
    });

    uki.each(('html,text,background,value,rect,checked,selectedIndex,' +
        'childViews,typeName,id,name,visible,disabled,focusable').split(','), function(i, name) {
        self[name] = function( value ) { return this.attr( name, value ); };
    });

    uki.each( ("blur,focus,load,resize,scroll,unload,click,dblclick," +
    	"mousedown,mouseup,mousemove,mouseover,mouseout,mouseenter,mouseleave," +
    	"change,select,submit,keydown,keypress,keyup,error").split(","), function(i, name){
    	self[name] = function( handler ){
    	    if (handler) {
        		this.bind(name, handler);
    	    } else {
                for (var i=0; i < this.length; i++) {
                    this[i][name]();
                };
    	    }
    		return this;
    	};
    });
})();
(function() {

    /**
     * Creates uki view tree from JSON-like markup
     *
     * @example
     * uki.build( {view: 'Button', rect: '100 100 100 24', text: 'Hello world' } )
     * // Creates uki.view.Button with '100 100 100 24' passed to constructor,
     * // and calls text('Hello world') on it
     *
     * @function
     * @name uki.build
     *
     * @param {object} ml JSON-like markup
     * @returns {uki.view.Collection} collection of created elemens
     */
    uki.build = function(ml) {
        if (ml.length === undefined) ml = [ml];
        return new uki.Collection(createMulti(ml));
    };

    function createMulti (ml) {
        return uki.map(ml, function(mlRow) { return createSingle(mlRow); });
    }

    function createSingle (mlRow) {
        if (uki.isFunction(mlRow.typeName)) {
            return mlRow;
        }

        var c = mlRow.view || mlRow.type,
            result;
        if (uki.isFunction(c)) {
            result = c();
        } else if (typeof c === 'string') {
            var parts = c.split('.'),
                obj   = root;
            if (!root[parts[0]] || parts[0] == 'Image') {
                parts = ['uki', 'view'].concat(parts); // try with default prefix
            }
            for (var i=0; i < parts.length; i++) {
                obj = obj[parts[i]];
            };
            result = new obj(mlRow.rect);
        } else {
            result = c;
        }

        copyAttrs(result, mlRow);
        return result;
    }

    function copyAttrs(comp, mlRow) {
        uki.each(mlRow, function(name, value) {
            if (name == 'view' || name == 'type' || name == 'rect') return;
            uki.attr(comp, name, value);
        });
        return comp;
    }

    uki.build.copyAttrs = copyAttrs;
})();
/* Ideas and code parts borrowed from Sizzle (http://sizzlejs.com/) */
(function() {
    var self,
        marked = '__selector_marked',
        attr = uki.attr,
        chunker = /((?:\((?:\([^()]+\)|[^()]+)+\)|\[(?:\[[^[\]]*\]|['"][^'"]*['"]|[^[\]'"]+)+\]|\\.|[^ >+~,(\[\\]+)+|[>+~])(\s*,\s*)?/g,
        regexps = [ // enforce order
    		{ name: 'ID', regexp: /#((?:[\w\u00c0-\uFFFF_-]|\\.)+)/ },
    		{ name: 'ATTR', regexp: /\[\s*((?:[\w\u00c0-\uFFFF_-]|\\.)+)\s*(?:(\S?=)\s*(['"]*)(.*?)\3|)\s*\]/ },
    		{ name: 'TYPE', regexp: /^((?:[\w\u00c0-\uFFFF\*_\.-]|\\.)+)/ },
    		{ name: 'POS',  regexp: /:(nth|eq|gt|lt|first|last|even|odd)(?:\((\d*)\))?(?=[^-]|$)/ }
	    ],
	    posRegexp = regexps.POS,
	    posFilters = {
    		first: function(i){
    			return i === 0;
    		},
    		last: function(i, match, array){
    			return i === array.length - 1;
    		},
    		even: function(i){
    			return i % 2 === 0;
    		},
    		odd: function(i){
    			return i % 2 === 1;
    		},
    		lt: function(i, match){
    			return i < match[2] - 0;
    		},
    		gt: function(i, match){
    			return i > match[2] - 0;
    		},
    		nth: function(i, match){
    			return match[2] - 0 == i;
    		},
    		eq: function(i, match){
    			return match[2] - 0 == i;
    		}
    	},
	    reducers = {
    		TYPE: function(comp, match) {
    		    var expected = match[1];
    		    if (expected == '*') return true;
    		    var typeName = attr(comp, 'typeName');
    		    return typeName && typeName.length >= expected.length &&
    		           ('.' + typeName).indexOf('.' + expected) == (typeName.length - expected.length);
    		},

    		ATTR: function(comp, match) {
    			var result = attr(comp, match[1]),
    			    value = result + '',
    				type = match[2],
    				check = match[4];

                return result == null ? type === "!=" :
                       type === "="   ? value === check :
                       type === "*="  ? value.indexOf(check) >= 0 :
                       type === "~="  ? (" " + value + " ").indexOf(check) >= 0 :
                       !check         ? value && result !== false :
                       type === "!="  ? value != check :
                       type === "^="  ? value.indexOf(check) === 0 :
                       type === "$="  ? value.substr(value.length - check.length) === check :
                       false;
    		},

    		ID: function(comp, match) {
    		    return reducers.ATTR(comp, ['', 'id', '=', '', match[1]]);
    		},

    		POS: function(comp, match, i, array) {
    			var filter = posFilters[match[1]];
				return filter ? filter( i, match, array ) : false;
    		}
	    },
	    mappers = {
    		"+": function(context){
    		},

    		">": function(context){
    		    return flatten(uki.map(context, 'childViews'));
    		},

    		"": function(context) {
    		    return recChildren(flatten(uki.map(context, 'childViews')));
    		},

    		"~": function(context){
    		}
	    };

	function recChildren (comps) {
	    return flatten(uki.map(comps, function(comp) {
	        return [comp].concat( recChildren(attr(comp, 'childViews')) );
	    }));
	}

	function flatten (array) {
	   return uki.reduce( [], array, reduceFlatten );
	}

	function reduceFlatten (x, e) {
	   return x.concat(e);
	}

	function removeDuplicates (array) {
	    var result = [],
	        i;

	    for (i = 0; i < array.length; i++) {
	        if (!array[i][marked]) { result[result.length] = array[i]; }
	        array[i][marked] = true;
	    };
	    for (i = 0; i < result.length; i++) { result[i][marked] = undefined; };
	    return result;
	}

    self =
    /**
     * @namespace
     */
    uki.Selector = {
    	/**
    	 * <p>Finds views by CSS3 selectors in view tree.</p>
    	 * <p>Can be called as uki(selector) instead of uki.Selector.find(selector)</p>
    	 *
    	 * @example
    	 *   uki('Label') find all labels on page
    	 *   uki('Box[name=main] > Label') find all imidiate descendant Labels in a box with name = "main"
    	 *   uki('> Slider', context) find all direct descendant Sliders within given context
    	 *   uki('Slider,Checkbox') find all Sliders and Checkboxes
    	 *   uki('Slider:eq(3)') find 3-d slider
    	 *
    	 * @param {string} selector
    	 * @param {Array.<uki.view.Base>} context to search in
    	 *
    	 * @return {uki.Collection} found views
    	 */
        find: function(selector, context, skipFiltering) {
            context = context || uki.top();
            if (context.length === undefined) context = [context];

            var tokens = self.tokenize(selector),
                expr   = tokens[0],
                extra  = tokens[1],
                result = context,
                mapper;

            while (expr.length > 0) {
                mapper = mappers[expr[0]] ? mappers[expr.shift()] : mappers[''];
                result = mapper(result);
                if (expr.length == 0) break;
                result = self.reduce(expr.shift(), result);
            }

            if (extra) {
                result = result.concat(self.find(extra, context, true));
            }

            return skipFiltering ? result : new uki.Collection(removeDuplicates(result));
        },

        /** @ignore */
        reduce: function(exprItem, context) {
            if (!context || !context.length) return [];

            var match, found;

            while (exprItem != '') {
                found = false;
                uki.each(regexps, function(index, row) {

                    /*jsl:ignore*/
                    if (match = exprItem.match(row.regexp)) {
                    /*jsl:end*/
                        found = true;
                        context = uki.grep(context, function(comp, index) {
                            return reducers[row.name](comp, match, index, context);
                        });
                        exprItem = exprItem.replace(row.regexp, '');
                        return false;
                    }
                });
                if (!found) break;
            }
            return context;
        },

        /** @ignore */
        tokenize: function(expr) {
        	var parts = [], match, extra;

        	chunker.lastIndex = 0;

        	while ( (match = chunker.exec(expr)) !== null ) {
        		parts.push( match[1] );

        		if ( match[2] ) {
        			extra = RegExp.rightContext;
        			break;
        		}
        	}

            return [parts, extra];
        }
    };

    uki.find = self.find;
})();
uki.extend(uki.dom, /** @lends uki.dom */ {
    bound: {},
    handles: {},

    events: ("blur,focus,load,resize,scroll,unload,click,dblclick," +
    	"mousedown,mouseup,mousemove,mouseover,mouseout,mouseenter,mouseleave," +
    	"change,select,submit,keydown,keypress,keyup,error").split(","),

    bind: function(el, types, handler) {
		if ( el.setInterval && el != window )
			el = window;

        handler.huid = handler.huid || uki.dom.guid++;

        var id = el[expando] = el[expando] || uki.dom.guid++,
            handle = uki.dom.handles[id] = uki.dom.handles[id] || function() {
                uki.dom.handler.apply(arguments.callee.elem, arguments);
            },
            i, type;

        handle.elem = el;

        if (!uki.dom.bound[id]) uki.dom.bound[id] = {};

        types = types.split(' ');
        for (i=0; i < types.length; i++) {
            type = types[i];
            if (!uki.dom.bound[id][type]) {
                el.addEventListener ? el.addEventListener(type, handle, false) : el.attachEvent('on' + type, handle);
                uki.dom.bound[id][type] = [];
            }
            uki.dom.bound[id][type].push(handler);
        };
        handler = handle = el = null;
    },

    unbind: function(el, types, handler) {
        var id = el[expando],
            huid = handler.huid,
            i, type;
        types = types.split(' ');
        for (i=0; i < types.length; i++) {
            type = types[i];
            if (!huid || !id || !uki.dom.bound[id] || !uki.dom.bound[id][type]) continue;
            uki.dom.bound[id][type] = uki.grep(uki.dom.bound[id][type], function(h) { return h.huid !== huid; });
        }
    },

    /** @ignore */
    handler: function( e ) {
        e = uki.dom.fix( e || root.event );

        var type = e.type,
            id = this[expando],
            handlers = uki.dom.bound[id],
            i;

        if (!id || !handlers || !handlers[type]) return;

        for (i=0, handlers = handlers[type]; i < handlers.length; i++) {
            handlers[i].call(this, e);
        };
    },

    preventDefault: function(e) {
        e.preventDefault ? e.preventDefault() : e.returnValue = false;
    },

    /**
     * Taken from jQuery
     * @ignore
     */
    fix: function( event ) {
		// Fix target property, if necessary
		if ( !event.target )
			event.target = event.srcElement || doc;

		// check if target is a textnode (safari)
		if ( event.target.nodeType == 3 )
			event.target = event.target.parentNode;

		// Add relatedTarget, if necessary
		if ( !event.relatedTarget && event.fromElement )
			event.relatedTarget = event.fromElement == event.target ? event.toElement : event.fromElement;

		// Calculate pageX/Y if missing and clientX/Y available
		if ( event.pageX == null && event.clientX != null ) {
			var de = doc.documentElement, body = doc.body;
			event.pageX = event.clientX + (de && de.scrollLeft || body && body.scrollLeft || 0) - (de.clientLeft || 0);
			event.pageY = event.clientY + (de && de.scrollTop  || body && body.scrollTop || 0)  - (de.clientTop || 0);
		}

		// Add which for key events
		if ( !event.which && ((event.charCode || event.charCode === 0) ? event.charCode : event.keyCode) )
			event.which = event.charCode || event.keyCode;

		// Add metaKey to non-Mac browsers (use ctrl for PC's and Meta for Macs)
		if ( !event.metaKey && event.ctrlKey )
			try { event.metaKey = event.ctrlKey; } catch(e){};

		// Add which for click: 1 == left; 2 == middle; 3 == right
		// Note: button is not normalized, so don't use it
		if ( !event.which && event.button )
			event.which = (event.button & 1 ? 1 : ( event.button & 2 ? 3 : ( event.button & 4 ? 2 : 0 ) ));

		return event;
    }
});

if (root.attachEvent) {
    root.attachEvent('onunload', function() {
        uki.each(uki.dom.bound, function(id, types) {
            uki.each(types, function(type, handlers) {
                try {
                    uki.dom.handles[id].elem.detachEvent('on' + type, uki.dom.handles[id]);
                } catch (e) {};
            });
        });
    });
};(function() {
    var self;

    if ( doc.documentElement["getBoundingClientRect"] ) {
    	self = uki.dom.offset = function( elem ) {
    		if ( !elem || elem == root ) return new Point();
    		if ( elem === elem.ownerDocument.body ) return self.bodyOffset( elem );
    		self.boxModel === undefined && self.initializeBoxModel();
    		var box  = elem.getBoundingClientRect(),
    		    doc = elem.ownerDocument,
    		    body = doc.body,
    		    docElem = doc.documentElement,
    			clientTop = docElem.clientTop || body.clientTop || 0,
    			clientLeft = docElem.clientLeft || body.clientLeft || 0,
    			top  = box.top  + (self.pageYOffset || self.boxModel && docElem.scrollTop  || body.scrollTop ) - clientTop,
    			left = box.left + (self.pageXOffset || self.boxModel && docElem.scrollLeft || body.scrollLeft) - clientLeft;

    		return new Point(left, top);
    	};
    } else {
    	self = uki.dom.offset = function( elem ) {
    		if ( !elem || elem == root ) return new Point();
    		if ( elem === elem.ownerDocument.body ) return self.bodyOffset( elem );
    		self.initialized || self.initialize();

    		var offsetParent = elem.offsetParent,
    		    prevOffsetParent = elem,
    			doc = elem.ownerDocument,
    			computedStyle,
    			docElem = doc.documentElement,
    			body = doc.body,
    			defaultView = doc.defaultView,
    			prevComputedStyle = defaultView.getComputedStyle(elem, null),
    			top = elem.offsetTop,
    			left = elem.offsetLeft;

    		while ( (elem = elem.parentNode) && elem !== body && elem !== docElem ) {
    			computedStyle = defaultView.getComputedStyle(elem, null);
    			top -= elem.scrollTop;
    			left -= elem.scrollLeft;

    			if ( elem === offsetParent ) {
    				top += elem.offsetTop;
    				left += elem.offsetLeft;

    				if ( self.doesNotAddBorder && !(self.doesAddBorderForTableAndCells && (/^t(able|d|h)$/i).test(elem.tagName)) ) {
    					top  += parseInt( computedStyle.borderTopWidth,  10) || 0;
    					left += parseInt( computedStyle.borderLeftWidth, 10) || 0;
    				}
    				prevOffsetParent = offsetParent;
    				offsetParent = elem.offsetParent;
    			}
    			if ( self.subtractsBorderForOverflowNotVisible && computedStyle.overflow !== "visible" ) {
    				top  += parseInt( computedStyle.borderTopWidth,  10) || 0;
    				left += parseInt( computedStyle.borderLeftWidth, 10) || 0;
    			}
    			prevComputedStyle = computedStyle;
    		}

    		if ( prevComputedStyle.position === "relative" || prevComputedStyle.position === "static" ) {
    			top  += body.offsetTop;
    			left += body.offsetLeft;
    		}

    		if ( prevComputedStyle.position === "fixed" ) {
    			top  += MAX(docElem.scrollTop, body.scrollTop);
    			left += MAX(docElem.scrollLeft, body.scrollLeft);
    		}

    		return new Point(left, top);
    	};
    }

    uki.extend(self, {
    	initialize: function() {
    		if ( this.initialized ) return;
    		var body = doc.body,
    		    container = doc.createElement('div'),
    		    innerDiv, checkDiv, table, td, rules, prop,
    		    bodyMarginTop = body.style.marginTop,
    			html = '<div style="position:absolute;top:0;left:0;margin:0;border:5px solid #000;padding:0;width:1px;height:1px;"><div></div></div><table style="position:absolute;top:0;left:0;margin:0;border:5px solid #000;padding:0;width:1px;height:1px;" cellpadding="0" cellspacing="0"><tr><td></td></tr></table>';

    		rules = { position: 'absolute', top: 0, left: 0, margin: 0, border: 0, width: '1px', height: '1px', visibility: 'hidden' };
    		for ( prop in rules ) container.style[prop] = rules[prop];

    		container.innerHTML = html;
    		body.insertBefore(container, body.firstChild);
    		innerDiv = container.firstChild;
    		checkDiv = innerDiv.firstChild;
    		td = innerDiv.nextSibling.firstChild.firstChild;

    		this.doesNotAddBorder = (checkDiv.offsetTop !== 5);
    		this.doesAddBorderForTableAndCells = (td.offsetTop === 5);

    		innerDiv.style.overflow = 'hidden';
    		innerDiv.style.position = 'relative';
    		this.subtractsBorderForOverflowNotVisible = (checkDiv.offsetTop === -5);

    		body.style.marginTop = '1px';
    		this.doesNotIncludeMarginInBodyOffset = (body.offsetTop === 0);
    		body.style.marginTop = bodyMarginTop;

    		body.removeChild(container);
    		this.boxModel === undefined && this.initializeBoxModel();
    		this.initialized = true;
    	},

    	initializeBoxModel: function() {
    	    if (this.boxModel !== undefined) return;
    		var div = doc.createElement("div");
    		div.style.width = div.style.paddingLeft = "1px";

    		doc.body.appendChild( div );
    		this.boxModel = div.offsetWidth === 2;
    		doc.body.removeChild( div ).style.display = 'none';
    	},

    	bodyOffset: function(body) {
    		self.initialized || self.initialize();
    		var top = body.offsetTop, left = body.offsetLeft;
    		if ( uki.dom.doesNotIncludeMarginInBodyOffset ) {
    			top  += parseInt( uki.dom.elem.currentStyle(body).marginTop, 10 ) || 0;
    			left += parseInt( uki.dom.elem.currentStyle(body).marginLeft, 10 ) || 0;
    		}
    		return new Point(left, top);
    	}
    });
})();
(function() {
    var controller = uki.dom.drag = {
        draggable: null,
        pos: null,

        start: function(draggable, e) {
            this.draggable = draggable;
            this.pos = new Point(e.pageX, e.pageY);
            bind();
        }
    };
    var doc = document;


    function bind() {
        uki.dom.bind(doc, 'mousemove scroll', dragging);
        uki.dom.bind(doc, 'mouseup', drop);
        uki.dom.bind(doc.body, 'selectstart mousedown', preventSelectionHandler);
    }

    function unbind() {
        uki.dom.unbind(doc, 'mousemove scroll', dragging);
        uki.dom.unbind(doc, 'mouseup', drop);
        uki.dom.unbind(doc.body, 'selectstart mousedown', preventSelectionHandler);
    }

    function dragging(e) {
        if (controller.draggable) controller.draggable._drag(e, offset(e));
    }

    function drop(e) {
        if (controller.draggable) controller.draggable._drop(e, offset(e));
        controller.draggable = null;
        unbind();
    }

    function preventSelectionHandler(e) {
		e.preventDefault ? e.preventDefault() : e.returnValue = false;
    }

    function offset (e) {
        return controller.pos.clone().offset(-e.pageX, -e.pageY);
    }
})();

uki.initNativeLayout = function() {
    if (uki.supportNativeLayout === undefined) {
        uki.dom.probe(
            uki.createElement(
                'div',
                'position:absolute;width:100px;height:100px;left:-999em;',
                '<div style="position:absolute;left:0;right:0"></div>'
            ),
            function(div) {
                uki.supportNativeLayout = div.childNodes[0].offsetWidth == 100 && !root.opera;
            }
        );
    }
};

// uki.supportNativeLayout = false;
/**
 * Creates image element from url
 *
 * @param {string} url Image url
 * @param {string=} dataUrl Data url representation of image, used if supported
 * @param {string=} alphaUrl Gif image url for IE6
 *
 * @namespace
 *
 * @returns {Element}
 */
uki.image = function(url, dataUrl, alphaUrl) {
    var result = new Image();
    result.src = uki.imageSrc(url, dataUrl, alphaUrl);
    return result;
};

/**
 * Selects image src depending on browser
 *
 * @param {string} url Image url
 * @param {string=} dataUrl Data url representation of image, used if supported
 * @param {string=} alphaUrl Gif image url for IE6
 *
 * @returns {string}
 */
uki.imageSrc = function(url, dataUrl, alphaUrl) {
    if (uki.image.dataUrlSupported && dataUrl) return dataUrl;
    if (alphaUrl && uki.image.needAlphaFix) return alphaUrl;
    return url;
};

/**
 * Image html representation: '<img src="">'
 *
 * @param {string} url Image url
 * @param {string=} dataUrl Data url representation of image, used if supported
 * @param {boolean=} alphaUrl Gif image url for IE6
 * @param {string=} html Additional html
 *
 * @returns {string} html
 */
uki.imageHTML = function(url, dataUrl, alphaUrl, html) {
    if (uki.image.needAlphaFix && alphaUrl) {
        url = alphaUrl;
    } else if (uki.image.dataUrlSupported) {
        url = dataUrl;
    }
    return '<img' + (html || '') + ' src="' + url + '" />';
};

/**
 * Loads given images, callbacks after all of them loads
 *
 * @param {Array.<Element>} images Images to load
 * @param {function()} callback
 */
uki.image.load = function(images, callback) {
    var imagesToLoad = images.length;

    uki.each(images, function(i, img) {
        if (!img || img.width) {
            if (!--imagesToLoad) callback();
            return;
        }

        var handler = function() {
                img.onload = img.onerror = img.onabort = null; // prevent mem leaks
                if (!--imagesToLoad) callback();
            };
		img.onload  = handler;
		img.onerror = handler;
		img.onabort = handler;
    });
};

/**
 * @type boolean
 */
uki.image.dataUrlSupported = doc.createElement('canvas').toDataURL || (/MSIE (8)/).test(ua);

/**
 * @type boolean
 */
uki.image.needAlphaFix = /MSIE 6/.test(ua);
if(uki.image.needAlphaFix) doc.execCommand("BackgroundImageCache", false, true);
(function() {
    var nullRegexp = /^\s*null\s*$/,
        themeRegexp  = /theme\s*\(\s*([^)]*\s*)\)/,
        rowsRegexp = /rows\s*\(\s*([^)]*\s*)\)/,
        cssBoxRegexp = /cssBox\s*\(\s*([^)]*\s*)\)/;

    /**
     * <p>Transforms a bg string into a background object</p>
     * <p>Supported strings:<br />
     *   theme(bg-name) Takes background with bg-name from uki.theme<br />
     *   rows(30, #CCFFFF, #FFCCFF, #FFFFCC) Creates Rows background with 30px rowHeight and 3 colors<br />
     *   cssBox(border:1px solid red;background:blue) Creates CssBox background with given cssText<br />
     *   url(i.png) or #FFFFFF Creates Css background with single property</p>
     *
     * @param {String} bg
     * @name uki.background
     * @namespace
     * @returns {uki.background.Base} created background
     */
    var self = uki.background = function(bg) {
        if (typeof(bg) === 'string') {
            var match;
            /*jsl:ignore*/
            if ( match = bg.match(nullRegexp) ) return new self.Null();
            if ( match = bg.match(themeRegexp) ) return uki.theme.background( match[1] );
            if ( match = bg.match(rowsRegexp) ) return new self.Rows( match[1].split(',', 2)[0], match[1].split(',', 2)[1] );
            if ( match = bg.match(cssBoxRegexp) ) return new self.CssBox( match[1] );
            /*jsl:end*/
            return new self.Css(bg);
        }
        return bg;
    };
})();
uki.background.Base = uki.background.Null = uki.newClass({
    init: uki.F,
    attachTo: uki.F,
    detach: uki.F
});/**
 * Adds a div with 9 sliced images in the corners, sides and center
 *
 * @class
 */
uki.background.Sliced9 = uki.newClass(new function() {
    var nativeCss = ['MozBorderImage', 'WebkitBorderImage', 'borderImage'],
        dom = uki.dom;

    var LEFT = 'left:',
        TOP  = 'top:',
        RIGHT = 'right:',
        BOTTOM = 'bottom:',
        WIDTH = 'width:',
        HEIGHT = 'height:',
        PX = 'px',
        P100 = '100%';

    var cache = {},
        proto = this;

    proto.init = function(partSettings, inset, options) {
        this._settings  = uki.extend({}, partSettings);
        this._inset     = Inset.create(inset);
        this._size      = null;
        this._inited    = false;

        options = options || {};
        this._fixedSize = Size.create(options.fixedSize) || new Size();
        this._bgInset   = Inset.create(options.inset)    || new Inset();
        this._zIndex    = options.zIndex || -1;

        this._container = this._getContainer();
        this._container.style.zIndex = this._zIndex;
    };

    function makeDiv (name, style, setting, imgStyle, bgSettings) {
        var inner = setting[3] ? img(setting, imgStyle) : '';
        if (!setting[3]) style += bgStyle(setting, bgSettings);
        return '<div class="' +  name + '" style="position:absolute;overflow:hidden;' + style + '">' + inner + '</div>';
    }

    function bgStyle (setting, bgSettings) {
        return ';background: url(' + uki.imageSrc(setting[0], setting[1], setting[2]) + ') ' + bgSettings;
    }

    function img (setting, style) {
        return uki.imageHTML(setting[0], setting[1], setting[2], ' galleryimg="no" style="-webkit-user-drag:none;position:absolute;' + style + '"');
    }

    proto._getContainer = function() {
        var key = this._getKey();
        if (!cache[key]) {
            return cache[key] = this._createContainer();
        }
        return cache[key].cloneNode(true);
    };

    proto._createContainer = function() {
        var inset = this._inset,
            bgInset = this._bgInset,
            settings = this._settings,
            width = inset.left + inset.right,
            height = inset.top + inset.bottom,
            css = [LEFT + bgInset.left + PX, RIGHT + bgInset.right + PX, TOP + bgInset.top + PX, BOTTOM + bgInset.bottom + PX].join(';'),
            html = [];

        if (inset.top && inset.left) {
            html[html.length] = makeDiv('tl',
                [LEFT + 0, TOP + 0, WIDTH + inset.left + PX, HEIGHT + inset.top + PX].join(';'),
                settings.c, [LEFT + 0, TOP + 0, WIDTH + width + PX, HEIGHT + height + PX].join(';'), 'top left'
            );
        }
        if (inset.top) {
            html[html.length] = makeDiv('t',
                [LEFT + inset.left + PX, TOP + 0, HEIGHT + inset.top + PX, RIGHT + inset.right + PX].join(';'),
                settings.h, [LEFT + 0, TOP + 0, WIDTH + P100, HEIGHT + height + PX].join(';'), 'repeat-x top'
            );
        }
        if (inset.top && inset.right) {
            html[html.length] = makeDiv('tr',
                [RIGHT + 0, TOP + 0, WIDTH + inset.right + PX, HEIGHT + inset.top + PX].join(';'),
                settings.c, [LEFT + '-' + inset.left + PX, TOP + 0, WIDTH + width + PX, HEIGHT + height + PX].join(';'), 'top right'
            );
        }

        if (inset.left) {
            html[html.length] = makeDiv('l',
                [LEFT + 0, TOP + inset.top + PX, WIDTH + inset.left + PX, BOTTOM + inset.bottom + PX].join(';'),
                settings.v, [LEFT + 0, TOP + 0, HEIGHT + P100, WIDTH + width + PX].join(';'), 'repeat-y left'
            );
        }
        if (settings.m) {
            html[html.length] = makeDiv('m',
                [LEFT + inset.left + PX, TOP + inset.top + PX, RIGHT + inset.left + PX, BOTTOM + inset.bottom + PX].join(';'),
                settings.m, [LEFT + 0, TOP + 0, HEIGHT + P100, WIDTH + P100].join(';'), ''
            );
        }
        if (inset.right) {
            html[html.length] = makeDiv('r',
                [RIGHT + 0, TOP + inset.top + PX, WIDTH + inset.right + PX, BOTTOM + inset.bottom + PX].join(';'),
                settings.v, [LEFT + '-' + inset.left + PX, TOP + 0, HEIGHT + P100, WIDTH + width + PX].join(';'), 'repeat-y right'
            );
        }

        if (inset.bottom && inset.left) {
            html[html.length] = makeDiv('bl',
                [LEFT + 0, BOTTOM + 0, WIDTH + inset.left + PX, HEIGHT + inset.bottom + PX].join(';'),
                settings.c, [LEFT + 0, TOP + '-' + inset.top + PX, WIDTH + width + PX, HEIGHT + height + PX].join(';'), 'left -' + inset.top + PX
            );
        }
        if (inset.bottom) {
            html[html.length] = makeDiv('b',
                [LEFT + inset.left + PX, BOTTOM + 0, HEIGHT + inset.bottom + PX, RIGHT + inset.right + PX].join(';'),
                settings.h, [LEFT + 0, TOP + '-' + inset.top + PX, WIDTH + P100, HEIGHT + height + PX].join(';'), 'repeat-x 0 -' + inset.top + PX
            );
        }
        if (inset.bottom && inset.right) {
            html[html.length] = makeDiv('br',
                [RIGHT + 0, BOTTOM + 0, WIDTH + inset.right + PX, HEIGHT + inset.bottom + PX].join(';'),
                settings.c, [LEFT + '-' + inset.left + PX, TOP + '-' + inset.top + PX, WIDTH + width + PX, HEIGHT + height + PX].join(';'), 'right -' + inset.top + PX
            );
        }
        return uki.createElement('div', 'position:absolute;overflow:hidden;' + css, html.join(''));
    };

    proto._getKey = function() {
        return uki.map(['v', 'h', 'm', 'c'], function(x) {
            return this._settings[x] && this._settings[x][0] || '';
        }, this).concat([this._inset, this._bgInset, this._fixedSize]).join(',');
    };

    proto.attachTo = function(comp) {
        this._comp = comp;

        this._container.style.visibility = 'visible';
        this._comp.dom().appendChild(this._container);

        if (!uki.supportNativeLayout) {
            var _this = this;
            this._layoutHandler = function(e) {
                if (_this._size && _this._size.eq(e.rect)) return;
                _this._size = e.rect;
                _this.layout();
            };
            this._comp.bind('layout', this._layoutHandler);
            this.layout();
        }
    };

    proto.detach = function() {
        if (this._comp) {
            // this._comp.dom().removeChild(this._container);
            this._container.style.visibility = 'hidden';
            if (!uki.supportNativeLayout) this._comp.unbind('layout', this._layoutHandler);
            this._size = this._comp = null;
            this._attached = this._inited = false;
        }
    };

    proto.layout = function() {
        var size = this._comp.rect(),
            parts = this._parts,
            inset = this._inset,
            bgInset = this._bgInset,
            fixedSize = this._fixedSize,
            width = FLOOR(fixedSize.width || size.width - bgInset.left - bgInset.right),
            height = FLOOR(fixedSize.height || size.height - bgInset.top - bgInset.bottom),
            insetWidth = inset.left + inset.right,
            insetHeight = inset.top + inset.bottom;

        if (!parts) {
            parts = {};
            uki.each(this._container.childNodes, function() {
                if (this.className) parts[this.className] = this;
            });
            this._parts = parts;
        }
        // parts.b.style.bottom = ''
        // parts.b.style.top = '100%';
        // parts.b.style.marginTop = - inset.bottom + 'px';
        if (parts.t) dom.layout(parts.t.style, { width: width - insetWidth });
        if (parts.b) dom.layout(parts.b.style, { width: width - insetWidth });
        if (parts.l) dom.layout(parts.l.style, { height: height - insetHeight });
        if (parts.r) dom.layout(parts.r.style, { height: height - insetHeight });
        if (parts.m) dom.layout(parts.m.style, {
            height: height - insetHeight,
            width: width - insetWidth
        });
        dom.layout(this._container.style, {
            width: width,
            height: height
        });
    };
});/**
 * Writes css properties to targets dom()
 *
 * @class
 */
uki.background.Css = uki.newClass(new function() {

    this.init = function(options) {
        this._options = typeof options == 'string' ? {background: options} : options;
    };

    this.attachTo = function(comp) {
        this._comp = comp;
        this._originalValues = {};
        var dom = comp.dom();

        uki.each(this._options, function(name, value) {
            this._originalValues[name] = dom.style[name];
            dom.style[name] = value;
        }, this);
    };

    this.detach = function() {
        if (this._comp) {
            var dom = this._comp.dom();
            uki.each(this._options, function(name, value) {
                dom.style[name] = this._originalValues[name];
            }, this);
        }

    };
});/**
 * Adds a div with given cssText to dom()
 *
 * @class
 */
uki.background.CssBox = uki.newClass(new function() {

    var cache = {};

    function getInsets(options) {
        if (!cache[options]) {
            var _this = this;
            uki.dom.probe(
                uki.createElement('div', options + ';position:absolute;overflow:hidden;left:-999em;width:10px;height:10px;'),
                function(c) {
                    cache[options] = new Inset(
                        c.offsetHeight - 10,
                        c.offsetWidth - 10
                    );
                }
            );
        }
        return cache[options];
    }

    this.init = function(options, ext) {
        this._options = options;
        ext = ext || {};
        this._inset = inset = Inset.create(ext.inset) || new Inset();
        this._insetWidth  = getInsets(options).left + inset.left + inset.right;
        this._insetHeight = getInsets(options).top + inset.top + inset.bottom;

        this._container = uki.createElement(
            'div',
            options + ';position:absolute;overflow:hidden;z-index:' + (ext.zIndex || '-1') + ';' +
            'left:' + inset.left + ';top:' + inset.top + 'px;right:' + inset.right + 'px;bottom:' + inset.bottom + 'px'
        );
        this._attached = false;
    };

    this.attachTo = function(comp) {
        var _this = this;
        this._comp = comp;
        this._comp.dom().appendChild(this._container);

        if (uki.supportNativeLayout) return;

        this._layoutHandler = function(e) {
            _this.layout(e.rect);
        };
        this._comp.bind('layout', this._layoutHandler);
        this.layout(this._comp.rect());
    };

    this.layout = function(size) {
        this._prevLayout = uki.dom.layout(this._container.style, {
            width: size.width - this._insetWidth,
            height: size.height - this._insetHeight
        }, this._prevLayout);
    };

    this.detach = function() {
        if (this._comp) {
            this._comp.dom().removeChild(this._container);
            if (!uki.supportNativeLayout) this._comp.unbind('layout', this._layoutHandler);
            this._attached = false;
        }
    };
});/**
 * Adds a div with colored rows to dom
 *
 * @class
 */
uki.background.Rows = uki.newClass(new function() {
    var proto = this,
        cache = [],
        packSize = 100;

    proto.init = function(height, colors) {
        this._height = height || 20;
        this._colors = uki.isArray(colors) ? colors : colors.split(' ');
        this._packSize = CEIL(packSize/this._colors.length)*this._colors.length;
        this._renderedHeight = 0;
        this._visibleExt = 200;
        if (this._colors.length == 1) this._colors = this._colors.concat(['#FFF']);
    };

    proto.attachTo = function(comp) {
        this._comp && this.detach();
        this._comp = comp;
        if (!this._container) {
            this._container = uki.createElement(
                'div',
                'position:absolute;left:0;top:0;width:100%;z-index:-1'
            );
        }
        var _this = this;
        this._layoutHandler = function(e) { _this.layout(e.rect, e.visibleRect); };
        this._comp.dom().appendChild(this._container);
        this._comp.bind('layout', this._layoutHandler);
    };

    proto.layout = function(rect, visibleRect) {
        var height = visibleRect ? visibleRect.height + this._visibleExt*2 : rect.maxY();
        while (this._renderedHeight < height) {
            var h = packSize * this._height,
                c = uki.createElement('div', 'height:' + h + 'px;overflow:hidden;width:100%;', getPackHTML(this._height, this._colors));
            this._renderedHeight += h;
            this._container.appendChild(c);
        }
        if (visibleRect) {
            this._container.style.top = CEIL((visibleRect.y - this._visibleExt)/this._height/this._colors.length)*this._height*this._colors.length + 'px';
        }
    };

    proto.detach = function() {
        this._comp.dom().removeChild(this._container);
        this._comp.unbind('layout', this._layoutHandler);
        this._comp = null;
    };

    function getPackHTML (height, colors) {
        var key = height + ' ' + colors.join(' '),
            rows = [],
            html = [],
            i, l = colors.length;
        if (!cache[key]) {
            for (i=0; i < l; i++) {
                rows[i] = ['<div style="height:', height, 'px;width:100%;overflow:hidden;',
                            (colors[i] ? 'background:' + colors[i] : ''),
                            '"></div>'].join('');
            };
            for (i=0; i < packSize; i++) {
                html[i] = rows[i%l];
            };
            cache[key] = html.join('');
        }
        return cache[key];
    }
});/**
 * @class
 */
uki.theme = {
    themes: [],

    register: function(theme) {
        uki.theme.themes.push(theme);
    },

    background: function(name, params) {
        return uki.theme._namedResource(name, 'background', params) || new uki.background.Null();
    },

    image: function(name, params) {
        return uki.theme._namedResource(name, 'image', params) || new Image();
    },

    imageSrc: function(name, params) {
        return uki.theme._namedResource(name, 'imageSrc', params) || '';
    },

    style: function(name, params) {
        return uki.theme._namedResource(name, 'style', params) || '';
    },

    dom: function(name, params) {
        return uki.theme._namedResource(name, 'dom', params) || uki.createElement('div');
    },

    template: function(name, params) {
        return uki.theme._namedResource(name, 'template', params) || '';
    },

    _namedResource: function(name, type, params) {
        for (var i = uki.theme.themes.length - 1; i >= 0; i--){
            var result = uki.theme.themes[i][type](name, params);
            if (result) return result;
        };
        return null;

    }
};/**
 * @class
 */
uki.theme.Base = {
    images: [],
    imageSrcs: [],
    backgrounds: [],
    doms: [],
    styles: [],
    templates: [],

    background: function(name, params) {
        return this.backgrounds[name] && this.backgrounds[name](params);
    },

    image: function(name, params) {
        if (this.images[name]) return this.images[name](params);
        return this.imageSrcs[name] && uki.image.apply(uki, this.imageSrcs[name](params));
    },

    imageSrc: function(name, params) {
        if (this.imageSrcs[name]) return uki.imageSrc.apply(uki, this.imageSrcs[name](params));
        return this.images[name] && this.images[name](params).src;
    },

    dom: function(name, params) {
        return this.doms[name] && this.doms[name](params);
    },

    style: function(name, params) {
        return this.styles[name] && this.styles[name](params);
    },

    template: function(name, params) {
        return this.templates[name] && this.templates[name](params);
    }
};/**
 * Simple and fast (2x–15x faster than regexp) html template
 * @example
 *   var t = new uki.theme.Template('<p class="${className}">${value}</p>')
 *   t.render({className: 'myClass', value: 'some html'})
 */
uki.theme.Template = function(code) {
    var parts = code.split('${'), i, l, tmp;
    this.parts = [parts[0]];
    this.names = [];
    for (i=1, l = parts.length; i < l; i++) {
        tmp = parts[i].split('}');
        this.names.push(tmp[0]);
        this.parts.push('');
        this.parts.push(tmp[1]);
    };
};

uki.theme.Template.prototype.render = function(values) {
    for (var i=0, names = this.names, l = names.length; i < l; i++) {
        this.parts[i*2+1] = values[names[i]] || '';
    };
    return this.parts.join('');
};/**
 * @class
 */
uki.view.utils = new function() {
    var proto = this;

    /** @exports proto as uki.view.utils */
    /**#@+ @memberOf uki.view.utils */

    proto.visibleRect = function (from, upTo) {
        var queue = [],
            rect, i, tmpRect, c = from;

        do {
            queue[queue.length] = c;
            c = c.parent();
        } while (c && c != upTo);

        if (upTo && upTo != from) queue[queue.length] = upTo;

        for (i = queue.length - 1; i >= 0; i--){
            c = queue[i];
            tmpRect = visibleRect(c);
            rect = rect ? rect.intersection(tmpRect) : tmpRect;
            rect.x -= c.rect().x;
            rect.y -= c.rect().y;

        };
        return rect;
    };

    proto.top = function(c) {
        while (c.parent()) c = c.parent();
        return c;
    };

    proto.offset = function(c, upTo) {
        var offset = new Point(),
            rect;

        while (c && c != upTo) {
            rect = c.rect();
            offset.x += rect.x;
            offset.y += rect.y;
            if (c.scrollTop) {
                offset.x -= c.scrollLeft();
                offset.y -= c.scrollTop();
            }
            c = c.parent();
        }
        return offset;
    };

    proto.scrollableParent = function(c) {
        do {
            if (uki.isFunction(c.scrollTop)) return c;
            c = c.parent();
        } while (c);
        return null;
    };

    /** @inner */
    function visibleRect (c) {
        return c.visibleRect ? c.visibleRect() : c.rect().clone();
    }
    /**#@-*/
};

uki.extend(uki.view, uki.view.utils);/**
 * @class
 */
uki.view.Focusable = {
    // dom: function() {
    //     return null; // should implement
    // },

    _focusable: true, // default value

    focusable: uki.newProp('_focusable', function(v) {
        this._focusable = v;
        if (this._focusableInput) this._focusableInput.style.display = v ? '' : 'none';
    }),

    _initFocusable: function(preCreatedInput) {
        if (!this._focusable) return;

        var input = preCreatedInput;
        if (!input) {
            input = uki.createElement(
                'input',
                uki.view.Base[PROTOTYPE].defaultCss + "border:none;padding:0;overflow:hidden;width:1px;height:1px;padding:1px;" +
                "font-size:1px;left:-9999em;top:50%;background:transparent;outline:none;opacity:0;"
            );
            this.dom().appendChild(input);
        }
        this._focusableInput = input;
        this._hasFocus = false;
        this._firstFocus = true;
        var _this = this,
            needsRefocus = doc.attachEvent;

        uki.dom.bind(input, 'focus', function(e) {
            if (_this._hasFocus) return;
            _this._hasFocus = true;
            _this._focus(e);
            _this._firstFocus = false;
            _this.trigger('focus', {domEvent: e, source: _this});
        });

        uki.dom.bind(input, 'blur', function(e) {
            if (!_this._hasFocus) return;
            _this._hasFocus = false;
            _this._blur(e);
            _this.trigger('blur', {domEvent: e, source: this});
        });

        if (!preCreatedInput) this.bind('mousedown', function(e) {
            if (!_this.hasFocus()) {
                _this._focusableInput.focus();
            } else {
                needsRefocus && setTimeout(function() {_this._focusableInput.focus();}, 1);
            }
            e.domEvent.preventDefault ? e.domEvent.preventDefault() : e.domEvent.returnValue = false;
        });
    },

    _focus: function(e) {
    },

    _blur: function(e) {
    },

    focus: function() {
        this._focusableInput.focus();
    },

    blur: function() {
        this._focusableInput.blur()
    },

    hasFocus: function() {
        return this._hasFocus;
    },

    _bindToDom: function(name) {
        if (!this._focusableInput || 'keyup keydown keypress'.indexOf(name) == -1) return false;

        return uki.view.Observable._bindToDom.call(this, name, this._focusableInput);
    }


};var ANCHOR_TOP    = 1,
    ANCHOR_RIGHT  = 2,
    ANCHOR_BOTTOM = 4,
    ANCHOR_LEFT   = 8,
    ANCHOR_WIDTH  = 16,
    ANCHOR_HEIGHT = 32;

uki.view.Base = uki.newClass(uki.view.Observable, new function() {

    /** @exports proto as uki.view.Base# */
    var proto = this,
        layoutId = 1;

    proto.defaultCss = 'position:absolute;z-index:100;-moz-user-focus:none;'
                     + 'font-family:Arial,Helvetica,sans-serif;';



    /**
     * Base class for all uki views.
     *
     * <p>View creates and layouts dom nodes. uki.view.Base defines basic API for other views.
     * It also defines common layout algorithms.</p>
     *
     * Layout
     *
     * <p>View layout is defined by rectangle and anchors.
     * Rectangle is passed to constructor, anchors are set through the #anchors attribute.</p>
     *
     * <p>Rectangle defines initial position and size. Anchors specify how view will move and
     * resize when its parent is resized.</p>
     *
     * 2 phases of layout
     *
     * <p>Layout process has 2 phases.
     * First views rectangles are recalculated. This may happen several times before dom
     * is touched. This is done either explictly through #rect attribute or through
     * #parentResized callbacks.
     * After rectangles are set #layout is called. This actually updates dom styles.</p>
     *
     * @example
     * uki({ view: 'Base', rect: '10 20 100 50', anchors: 'left top right' })
     * // Creates Base view with initial x = 10px, y = 20px, width = 100px, height = 50px.
     * // When parent resizes x, y and height will stay the same. Width will resize with parent.
     *
     *
     * @see uki.view.Base#anchors
     * @constructor
     *
     * @name uki.view.Base
     * @implements uki.view.Observable
     * @param {uki.geometry.Rect} rect initial position and size
     */
    proto.init = function(rect) {
        this._parentRect = this._rect = Rect.create(rect);
        this._setup();
        uki.initNativeLayout();
        this._createDom();
    };

    /**#@+ @memberOf uki.view.Base# */

    proto._setup = function() {
        uki.extend(this, {
           _anchors: 0,
           _parent: null,
           _visible: true,
           _needsLayout: true,
           _selectable: false,
           _styleH: 'left',
           _styleV: 'top',
           _firstLayout: true
        });
    };

    /**
     * Get views container dom node.
     * @returns {Element} dom
     */
    proto.dom = function() {
        return this._dom;
    };

    /* ------------------------------- Common settings --------------------------------*/
    /**
     * Full type name of a view. Used by uki.Selector
     * @return {String}
     */
    proto.typeName = function() {
        return 'uki.view.Base';
    };

    /**
     * Used for fast (on hash lookup) view searches: uki('#view_id');
     *
     * @param {string=} id New id value
     * @returns {string|uki.view.Base} current id or self
     */
    proto.id = function(id) {
        if (id === undefined) return this._dom.id;
        if (this._dom.id) uki.unregisterId(this);
        this._dom.id = id;
        uki.registerId(this);
        return this;
    };

    /**
     * Accessor for dom().className
     * @param {string=} className
     * @returns {string|uki.view.Base} className or self
     */
    uki.delegateProp(proto, 'className', '_dom');

    /**
     * Accessor for view visibility.
     *
     * @param {boolean=} state
     * @returns {boolean|uki.view.Base} current visibility state of self
     */
    proto.visible = function(state) {
        if (state === undefined) return this._dom.style.display != 'none';

        this._dom.style.display = state ? 'block' : 'none';
        return this;
    };

    proto.toggle = function() {
        this.visible(!this.visible());
    };

    /**
     * @todo move to a separate interface
     * Sets wherether text of the view can be selected.
     *
     * @param {boolean=} state
     * @returns {boolean|uki.view.Base} current selectable state of self
     */
    proto.selectable = function(state) {
        if (state === undefined) return this._selectable;

        this._selectable = state;
        this._dom.unselectable = state ? '' : 'on';
        var style = this._dom.style;
        style.userSelect = state ? 'text' : 'none';
        style.MozUserSelect = state ? 'text' : '-moz-none';
        style.WebkitUserSelect = state ? 'text' : 'none';
        style.cursor = state ? 'text' : 'default';
        return this;
    };

    /**
     * Accessor for background attribute.
     * @name background
     * @function
     * @param {string|uki.background.Base=} background
     * @returns {uki.background.Base|uki.view.Base} current background or self
     */
    /**
     * Accessor for shadow background attribute.
     * @name shadow
     * @function
     * @param {string|uki.background.Base=} background
     * @returns {uki.background.Base|uki.view.Base} current background or self
     */
    /**
     * Accessor for default background attribute.
     * @name defaultBackground
     * @function
     * @returns {uki.background.Base} default background if not overriden through attribute
     */
    /**
     * Accessor for default shadow background attribute.
     * @name defaultShadow
     * @function
     * @returns {uki.background.Base} default background if not overriden through attribute
     */
    uki.each(['background', 'shadow'], function(i, name) {
        var field = '_' + name,
            defaultAttr = 'default' + name.substr(0, 1).toUpperCase() + name.substr(1),
            defaultField = '_' + defaultAttr;

        proto[name] = function(val) {
            if (val === undefined && !this[field] && this[defaultAttr]) this[field] = this[defaultAttr]();
            if (val === undefined) return this[field];
            val = uki.background(val);

            if (val == this[field]) return this;
            if (this[field]) this[field].detach(this);
            val.attachTo(this);

            this[field] = val;
            return this;
        };

        proto[defaultAttr] = function() {
            return this[defaultField] && uki.theme.background(this[defaultField]);
        };
    });

    /* ----------------------------- Container api ------------------------------*/

    /**
     * Accessor attibute for parent view. When parent is set view appends its #dom
     * to parents #domForChild
     *
     * @param {?uki.view.Base=} parent
     * @returns {uki.view.Base} parent or self
     */
    proto.parent = function(parent) {
        if (parent === undefined) return this._parent;

        if (this._parent) this._dom.parentNode.removeChild(this._dom);
        this._parent = parent;
        if (this._parent) this._parent.domForChild(this).appendChild(this._dom);
        return this;
    };

    /**
     * Accessor for childViews. @see uki.view.Container for implementation
     * @returns {Array.<uki.view.Base>}
     */
    proto.childViews = function() {
        return [];
    };


    /* ----------------------------- Layout ------------------------------*/

    /**
     * Sets or retrieves view's position and size.
     *
     * @param {string|uki.geometry.Rect} newRect
     * @returns {uki.view.Base} self
     */
    proto.rect = function(newRect) {
        if (newRect === undefined) return this._rect;

        newRect = Rect.create(newRect);
        this._parentRect = newRect;
        // if (newRect.eq(this._rect)) return false;
        this._rect = this._normalizeRect(newRect);
        this._needsLayout = this._needsLayout || layoutId++;
        return this;
    };

    /**
     * Set or get sides which the view should be attached to.
     * When a view is attached to a side the distance between this side and views border
     * will remain constant on resize. Anchor can be any combination of
     * "top", "right", "bottom", "left", "width" and "height".
     * If you set both "right" and "left" than "width" is assumed.
     *
     * Anchors are stored as a bitmask. Though its easier to set them using strings
     *
     * @function
     * @param {string|number} anchors
     * @returns {number|uki.view.Base} anchors or self
     */
    proto.anchors = uki.newProp('_anchors', function(anchors) {
        if (anchors.indexOf) {
            var tmp = 0;
            if (anchors.indexOf('right'  ) > -1) tmp |= ANCHOR_RIGHT;
            if (anchors.indexOf('bottom' ) > -1) tmp |= ANCHOR_BOTTOM;
            if (anchors.indexOf('top'    ) > -1) tmp |= ANCHOR_TOP;
            if (anchors.indexOf('left'   ) > -1) tmp |= ANCHOR_LEFT;
            if (anchors.indexOf('width'  ) > -1 || (tmp & ANCHOR_LEFT && tmp & ANCHOR_RIGHT)) tmp |= ANCHOR_WIDTH;
            if (anchors.indexOf('height' ) > -1 || (tmp & ANCHOR_BOTTOM && tmp & ANCHOR_TOP)) tmp |= ANCHOR_HEIGHT;
            anchors = tmp;
        }
        this._anchors = anchors;
        this._styleH = anchors & ANCHOR_LEFT ? 'left' : 'right';
        this._styleV = anchors & ANCHOR_TOP ? 'top' : 'bottom';
    });

    /**
     * Returns rectangle for child layout. Usualy equals to #rect. Though in some cases,
     * client rectangle my differ from #rect. Example uki.view.ScrollPane.
     *
     * @param {uki.view.Base} child
     * @returns {uki.geometry.Rect}
     */
    proto.rectForChild = function(child) {
        return this.rect();
    };

    /**
     * Updates dom to match #rect property.
     *
     * Layout is designed to minimize dom writes. If view is anchored to the right then
     * style.right is used, style.left for left anchor, etc. If browser supports this
     * both style.right and style.left are used. Otherwise style.width will be updated
     * manualy on each resize.
     *
     * @see uki.dom.initNativeLayout
     */
    proto.layout = function() {
        this._layoutDom(this._rect);
        this._needsLayout = false;
        this.trigger('layout', {rect: this._rect, source: this});
        this._firstLayout = false;
    };

    proto.minSize = uki.newProp('_minSize', function(s) {
        this._minSize = Size.create(s);
        this.rect(this._parentRect);
        if (this._minSize.width) this._dom.style.minWidth = this._minSize.width + 'px';
        if (this._minSize.height) this._dom.style.minHeight = this._minSize.height + 'px';
    });

    /**
     * Resizes view when parent changes size acording to anchors.
     * Called from parent view. Usualy after parent's #rect is called.
     *
     * @param {uki.geometry.Rect} oldSize
     * @param {uki.geometry.Rect} newSize
     */
    proto.parentResized = function(oldSize, newSize) {
        var newRect = this._parentRect.clone(),
            dX = (newSize.width - oldSize.width) /
                ((this._anchors & ANCHOR_LEFT ^ ANCHOR_LEFT ? 1 : 0) +   // flexible left
                (this._anchors & ANCHOR_WIDTH ? 1 : 0) +
                (this._anchors & ANCHOR_RIGHT ^ ANCHOR_RIGHT ? 1 : 0)),   // flexible right
            dY = (newSize.height - oldSize.height) /
                ((this._anchors & ANCHOR_TOP ^ ANCHOR_TOP ? 1 : 0) +      // flexible top
                (this._anchors & ANCHOR_HEIGHT ? 1 : 0) +
                (this._anchors & ANCHOR_BOTTOM ^ ANCHOR_BOTTOM ? 1 : 0)); // flexible right

        if (this._anchors & ANCHOR_LEFT ^ ANCHOR_LEFT) newRect.x += dX;
        if (this._anchors & ANCHOR_WIDTH) newRect.width += dX;

        if (this._anchors & ANCHOR_TOP ^ ANCHOR_TOP) newRect.y += dY;
        if (this._anchors & ANCHOR_HEIGHT) newRect.height += dY;
        this.rect(newRect);
    };

    /**
     * Resizes view to its contents. Contents size is determined by view.
     * View can be resized by width, height or both. This is specified throght
     * autosizeStr param.
     * View will grow shrink acording to its #anchors.
     *
     * @param {autosizeStr} autosize
     * @returns {uki.view.Base} self
     */
    proto.resizeToContents = function(autosizeStr) {
        var autosize = decodeAutosize(autosizeStr);
        if (0 == autosize) return this;

        var oldRect = this.rect(),
            newRect = this._calcRectOnContentResize(autosize);
        // if (newRect.eq(oldRect)) return this;
        // this.rect(newRect);
        this._rect = this._parentRect = newRect;
        this._needsLayout = true;
        return this;
    };

    /**
     * Calculates view's contents size. Redefined in subclasses.
     *
     * @param {number} autosize Bitmask
     * @returns {uki.geometry.Rect}
     */
    proto.contentsSize = function(autosize) {
        return this.rect();
    };

    proto._normalizeRect = function(rect) {
        if (this._minSize) {
            rect = new Rect(rect.x, rect.y, MAX(this._minSize.width, rect.width), MAX(this._minSize.height, rect.height));
        }
        return rect;
    };



    function decodeAutosize (autosizeStr) {
        if (!autosizeStr) return 0;
        var autosize = 0;
        if (autosizeStr.indexOf('width' ) > -1) autosize = autosize | ANCHOR_WIDTH;
        if (autosizeStr.indexOf('height') > -1) autosize = autosize | ANCHOR_HEIGHT;
        return autosize;
    }


    proto._initBackgrounds = function() {
        if (this.background()) this.background().attachTo(this);
        if (this.shadow()) this.shadow().attachTo(this);
    };

    proto._calcRectOnContentResize = function(autosize) {
        var newSize = this.contentsSize( autosize ),
            oldSize = this.rect();

        if (newSize.eq(oldSize)) return oldSize; // nothing changed

        // calculate where to resize
        var newRect = this.rect().clone(),
            dX = newSize.width - oldSize.width,
            dY = newSize.height - oldSize.height;

        if (autosize & ANCHOR_WIDTH) {
            if (this._anchors & ANCHOR_LEFT ^ ANCHOR_LEFT && this._anchors & ANCHOR_RIGHT ^ ANCHOR_RIGHT) {
                newRect.x -= dX/2;
            } else if (this._anchors & ANCHOR_LEFT ^ ANCHOR_LEFT) {
                newRect.x -= dX;
            }
            newRect.width += dX;
        }

        if (autosize & ANCHOR_HEIGHT) {
            if (this._anchors & ANCHOR_TOP ^ ANCHOR_TOP && this._anchors & ANCHOR_BOTTOM ^ ANCHOR_BOTTOM) {
                newRect.y -= dY/2;
            } else if (this._anchors & ANCHOR_TOP ^ ANCHOR_TOP) {
                newRect.y -= dY;
            }
            newRect.height += dY;
        }

        return newRect;
    };

    uki.each(['width', 'height', 'minX', 'maxX', 'minY', 'maxY', 'left', 'top'], function(index, attr) {
        proto[attr] = function() {
            var rect = this.rect();
            return rect[attr].apply(rect, arguments);
        };
    });

    /* ---------------------------------- Dom --------------------------------*/
    /**
     * Called through a second layout pass when _dom should be created
     */
    proto._createDom = function() {
        this._dom = uki.createElement('div', this.defaultCss);
    };

    /**
     * Called through a second layout pass when _dom is allready created
     */
    proto._layoutDom = function(rect) {
        var l = {}, s = uki.supportNativeLayout, relativeRect = this.parent().rectForChild(this);
        if (s && this._anchors & ANCHOR_LEFT && this._anchors & ANCHOR_RIGHT) {
            l.left = rect.x;
            l.right = relativeRect.width - rect.x - rect.width;
        } else {
            l.width = rect.width;
            l[this._styleH] = this._styleH == 'left' ? rect.x : relativeRect.width - rect.x - rect.width;
        }

        if (s && this._anchors & ANCHOR_TOP && this._anchors & ANCHOR_BOTTOM) {
            l.top = rect.y;
            l.bottom = relativeRect.height - rect.y - rect.height;
        } else {
            l.height = rect.height;
            l[this._styleV] = this._styleV == 'top'  ? rect.y : relativeRect.height - rect.y - rect.height;
        }
        this._lastLayout = uki.dom.layout(this._dom.style, l, this._lastLayout);
        if (this._firstLayout) this._initBackgrounds();
        return true;
    };

    proto._bindToDom = function(name) {
        if ('resize layout'.indexOf(name) > -1) return true;
        return uki.view.Observable._bindToDom.call(this, name);
    };

    /**#@-*/
});uki.view.Container = uki.newClass(uki.view.Base, new function() {
    var Base = uki.view.Base[PROTOTYPE],
        /**
         * @class
         * @name uki.view.Container
         */
        proto = this;

    /** @exports proto as uki.view.Container# */

    /**#@+ @memberOf uki.view.Container# */

    proto._setup = function() {
        this._childViews = [];
        Base._setup.call(this);
    };

    proto.typeName = function() {
        return 'uki.view.Container';
    };

    function maxProp (c, prop) {
        var val = 0, i, l;
        for (i = c._childViews.length - 1; i >= 0; i--){
            val = MAX(c._childViews[i].rect()[prop]());
        };
        return val;
    }

    proto.contentsWidth = function() {
        return maxProp(this, 'maxX') - maxProp(this, 'minX');
    };

    proto.contentsHeight = function() {
        return maxProp(this, 'maxY') - maxProp(this, 'minY');
    };

    proto.contentsSize = function() {
        return new Size(this.contentsWidth(), this.contentsHeight());
    };

    /**
     * Sets or retrievs view child view.
     * @param anything uki.build can parse
     *
     * Note: if setting on view with child views, all child view will be removed
     */
    proto.childViews = function(val) {
        if (val === undefined) return this._childViews;
        uki.each(this._childViews, function(i, child) {
            this.removeChild(child);
        }, this);
        uki.each(uki.build(val), function(tmp, child) {
            this.appendChild(child);
        }, this);
        return this;
    };

    /**
     * Remove particular child
     */
    proto.removeChild = function(child) {
        child.parent(null);
        var index = child._viewIndex,
            i, l;
        for (i=index+1, l = this._childViews.length; i < l; i++) {
            this._childViews[i]._viewIndex--;
        };
        this._childViews = uki.grep(this._childViews, function(elem) { return elem != child; });
    };

    /**
     * Adds a child.
     */
    proto.appendChild = function(child) {
        child._viewIndex = this._childViews.length;
        this._childViews.push(child);
        child.parent(this);
    };

    /**
     * Should return a dom node for a child.
     * Child should append itself to this dom node
     */
    proto.domForChild = function(child) {
        return this._dom;
    };



    proto._layoutDom = function(rect) {
        Base._layoutDom.call(this, rect);
        this._layoutChildViews(rect);
    };

    proto._layoutChildViews = function() {
        for (var i=0, childViews = this.childViews(); i < childViews.length; i++) {
            if (childViews[i]._needsLayout && childViews[i].visible()) {
                childViews[i].layout(this._rect);
            }
        };
    };


    proto.rect = function(newRect) {
        if (newRect === undefined) return this._rect;

        newRect = Rect.create(newRect);
        this._parentRect = newRect;
        var oldRect = this._rect;
        if (!this._resizeSelf(newRect)) return this;
        this._needsLayout = true;

        if (oldRect.width != newRect.width || oldRect.height != newRect.height) this._resizeChildViews(oldRect);
        this.trigger('resize', {oldRect: oldRect, newRect: this._rect, source: this});
        return this;
    };

    proto._resizeSelf = function(newRect) {
        // if (newRect.eq(this._rect)) return false;
        this._rect = this._normalizeRect(newRect);
        return true;
    };

    /**
     * Called to notify all intersted parties: childViews and observers
     */
    proto._resizeChildViews = function(oldRect) {
        for (var i=0, childViews = this.childViews(); i < childViews.length; i++) {
            childViews[i].parentResized(oldRect, this._rect);
        };
    };

   /**#@-*/
});uki.view.Box = uki.newClass(uki.view.Container, {
    typeName: function() { return 'uki.view.Box'; }
});
uki.view.Image = uki.newClass(uki.view.Base, new function() {
    var proto = this;

    proto.typeName = function() { return 'uki.view.Image'; };

    uki.delegateProp(proto, 'src', '_dom');

    proto._createDom = function() {
        this._dom = uki.createElement('img', this.defaultCss)
    };

});
uki.view.Label = uki.newClass(uki.view.Base, new function() {
    var Base = uki.view.Base[PROTOTYPE],
        proto = this;


    proto._setup = function() {
        Base._setup.call(this);
        uki.extend(this, {
            _scrollable: false,
            _selectable: false,
            _inset: new Inset()
        });
    };

    proto.typeName = function() {
        return 'uki.view.Label';
    };

    proto.selectable = function(state) {
        if (state !== undefined) {
            this._label.unselectable = state ? '' : 'on';
        }
        return Base.selectable.call(this, state);
    };

    proto.fontSize = function(size) {
        if (size === undefined) return this._label.style.fontSize;

        this._label.style.fontSize = this._label.style.lineHeight = size + 'px';
        return this;
    };

    /**
     * Warning! this operation is expensive
     */
    proto.contentsSize = function(autosize) {
        var clone = this._createLabelClone(autosize), inset = this.inset(), size;

        uki.dom.probe(clone, function() {
            size = new Size(clone.offsetWidth + inset.width(), clone.offsetHeight + inset.height());
        });

        return size;
    };

    proto.text = function(text) {
        return text === undefined ? this.html() : this.html(uki.escapeHTML(text));
    };

    proto.html = function(html) {
        if (html === undefined) return this._label.innerHTML;
        this._label.innerHTML = html;
        return this;
    };

    proto._css = function(name, value) {
        if (value === undefined) return this._label.style[name];
        this._label.style[name] = value;
        return this;
    };

    uki.each(['fontSize', 'textAlign', 'color', 'fontFamily', 'fontWeight', 'lineHeight'], function(i, name) {
        proto[name] = function(value) {
            return this._css(name, value);
        };
    });

    proto.inset = uki.newProp('_inset', function(inset) {
        this._inset = Inset.create(inset);
    });

    proto.scrollable = uki.newProp('_scrollable', function(state) {
        this._scrollable = state;
        this._label.style.overflow = state ? 'auto' : 'hidden';
    });

    proto.multiline = uki.newProp('_multiline', function(state) {
        this._multiline = state;
        this._label.style.whiteSpace = state ? '' : 'nowrap';
    });

    proto._createLabelClone = function(autosize) {
        var clone = this._label.cloneNode(true),
            inset = this.inset(), rect = this.rect();

        if (autosize & ANCHOR_WIDTH) {
            clone.style.width = clone.style.right = '';
        } else if (uki.supportNativeLayout) {
            clone.style.right = '';
            clone.style.width = rect.width - inset.width() + 'px';
        }
        if (autosize & ANCHOR_HEIGHT) {
            clone.style.height = clone.style.bottom = '';
        } else if (uki.supportNativeLayout) {
            clone.style.bottom = '';
            clone.style.height = rect.height - inset.height() + 'px';
        }
        clone.style.paddingTop = 0;
        clone.style.visibility = 'hidden';
        return clone;
    };

    proto._createDom = function() {
        Base._createDom.call(this);
        this._label = uki.createElement('div', Base.defaultCss +
            "font-size:12px;line-height:12px;white-space:nowrap;"); // text-shadow:0 1px 0px rgba(255,255,255,0.8);
        this._dom.appendChild(this._label);
        this.selectable(this.selectable());
    };

    proto._layoutDom = function() {
        Base._layoutDom.apply(this, arguments);

        var inset = this._inset;
        if (!this.multiline()) {
            var fz = parseInt(this._css('fontSize'), 10) || 12;
            this._label.style.lineHeight = (this._rect.height - inset.top - inset.bottom) + 'px';
            // this._label.style.paddingTop = MAX(0, this._rect.height/2 - fz/2) + 'px';
        }
        var l;

        if (uki.supportNativeLayout) {
            l = {
                left: inset.left,
                top: inset.top,
                right: inset.right,
                bottom: inset.bottom
            };
        } else {
            l = {
                left: inset.left,
                top: inset.top,
                width: this._rect.width - inset.width(),
                height: this._rect.height - inset.height()
            };
        }
        this._lastLabelLayout = uki.dom.layout(this._label.style, l, this._lastLabelLayout);
    };
});
uki.view.Button = uki.newClass(uki.view.Label, uki.view.Focusable, new function() {
    var proto = this,
        Base = uki.view.Label[PROTOTYPE];

    proto._setup = function() {
        Base._setup.call(this);
        uki.extend(this, {
            _inset: new Inset(0, 4),
            _backgroundPrefix: '',
            defaultCss: Base.defaultCss + "cursor:default;-moz-user-select:none;-webkit-user-select:none;" //text-shadow:0 1px 0px rgba(255,255,255,0.8)
        });
    };

    proto.typeName = function() {
        return 'uki.view.Button';
    };

    uki.addProps(proto, ['backgroundPrefix', 'icon']);

    uki.each(['normal', 'hover', 'down', 'focus'], function(i, name) {
        var property = name + '-background';
        proto[property] = function(bg) {
            if (bg) this['_' + property] = bg;
            return this['_' + property] = this['_' + property] ||
                uki.theme.background(this._backgroundPrefix + 'button-' + name, {height: this.rect().height, view: this});
        };
    });

    proto._createLabelClone = function(autosize) {
        var clone = Base._createLabelClone.call(this, autosize);
        clone.style.fontWeight = 'bold';
        return clone;
    };

    proto._layoutDom = function(rect) {
        Base._layoutDom.call(this, rect);
        if (this._firstLayout) {
            this['hover-background']();
            this['down-background']();

            this._backgroundByName('normal');
            this._initFocusable();
        }
    };

    proto._createDom = function() {
        // dom
        this._dom = uki.createElement('div', this.defaultCss);
        this._label = uki.createElement('div', Base.defaultCss +
            "font-size:12px;line-height:12px;white-space:nowrap;text-align:center;font-weight:bold;color:#333;"); // text-shadow:0 1px 0px rgba(255,255,255,0.8);
        this._dom.appendChild(this._label);
        if (this._dom.attachEvent) {
            // click handler for ie
            this._dom.appendChild(uki.createElement('div', 'width:100%;height:100%;position:absolute;background:url(' + uki.theme.imageSrc('x') + ');'));
        }

        // events
        this._down = false;

        var _this = this;

        this._mouseup = function(e) {
            _this._backgroundByName(_this._over ? 'hover' : 'normal');
            _this._down = false;
        };

        this._mousedown = function(e) {
            _this._backgroundByName('down');
            _this._down = true;
        };

        var supportMouseEnter = this._dom.attachEvent && !root.opera;

        this._mouseover = function(e) {
            if (!supportMouseEnter && uki.dom.contains(_this._dom, e.relatedTarget) || _this._over) return;
            _this._backgroundByName(_this._down ? 'down' : 'hover');
            _this._over = true;
        };

        this._mouseout = function(e) {
            if (!supportMouseEnter && uki.dom.contains(_this._dom, e.relatedTarget) || !_this._over) return;
            _this._backgroundByName('normal');
            _this._over = false;
        };

        uki.dom.bind(document, 'mouseup', this._mouseup);
        uki.dom.bind(this._dom, 'mousedown', this._mousedown);
        uki.dom.bind(this._dom, supportMouseEnter ? 'mouseenter' : 'mouseover', this._mouseover);
        uki.dom.bind(this._dom, supportMouseEnter ? 'mouseleave' : 'mouseout', this._mouseout);
        this.selectable(this.selectable());
    };

    proto._focus = function() {
        this['focus-background']().attachTo(this);
        if (this._firstFocus) {
            var _this = this;
            uki.dom.bind(this._focusableInput, 'keydown', function(e) {
                if ((e.which == 32 || e.which == 13) && !_this._down) _this._mousedown();
            });
            uki.dom.bind(this._focusableInput, 'keyup', function(e) {
                if ((e.which == 32 || e.which == 13) && _this._down) {
                    _this._mouseup();
                    _this.trigger('click', {domEvent: e, source: _this});
                }
                if (e.which == 27 && _this._down) {
                    _this._mouseup();
                }
            });
        }
    };

    proto._blur = function() {
       this['focus-background']().detach();
    };

    proto._backgroundByName = function(name) {
        var bg = this[name + '-background']();
        if (this._background == bg) return;
        if (this._background) this._background.detach();
        bg.attachTo(this);
        this._background = bg;
        this._backgroundName = name;
    };

    proto._bindToDom = function(name) {
        return uki.view.Focusable._bindToDom.call(this, name) || uki.view.Label[PROTOTYPE]._bindToDom.call(this, name);
    };
});(function() {

var Base = uki.view.Base[PROTOTYPE],
self = uki.view.Checkbox = uki.newClass(uki.view.Base, uki.view.Focusable, {

    _bindToDom: function(name) {
        if (' change'.indexOf(name) > -1) return;
        Base._bindToDom.apply(this, arguments);
    },

    _imageSize: 18,

    _setup: function() {
        Base._setup.call(this);
        uki.extend(this, {
            _checked: false,
            _selectable: false,
            _disabled: false
        });
    },

    disabled: uki.newProp('_disabled', function(d) {
        this._disabled = d;
        if (d) this.blur();
        this._focusableInput.disabled = d;
        this._updateBg();
    }),

    checked: function(state) {
        if (state === undefined) return this._checked;
        this._checked = !!state;
        this._updateBg();
        return this;
    },

    _updateBg: function() {
        var position = this._checked ? 0 : this._imageSize;
        position += this._disabled ? this._imageSize*4 : this._over ? this._imageSize*2 : 0;
        this._image.style.top = - position + PX;
    },

    _createDom: function() {
        var w = this._imageSize + PX,
            hw = this._imageSize / 2 + PX;

        this._image = uki.theme.image('checkbox');
        this._dom = uki.createElement('div', Base.defaultCss + 'overflow:visible');
        this._box = uki.createElement('div', Base.defaultCss + 'overflow:hidden;left:50%;top:50%;margin-left:-9px;margin-top:-9px;width:' + w + ';height:' + w);
        this._image.style.position = 'absolute';
        this._image.style.WebkitUserDrag = 'none';
        this._focusImage = uki.theme.image('checkbox-focus');
        this._focusImage.style.cssText += 'display:block;-webkit-user-drag:none;position:absolute;z-index:-1;left:50%;top:50%;';

        this._dom.appendChild(this._box);
        this._box.appendChild(this._image);

        var _this = this;
        this._click = function() {
            if (_this._disabled) return;
            _this.checked(!_this._checked);
            _this.trigger('change', {checked: _this._checked, source: _this});
        };

        uki.dom.bind(this._box, 'click', this._click);
        uki.dom.bind(this._box, 'mouseover', function() {
            _this._over = true;
            _this._updateBg();
        });
        uki.dom.bind(this._box, 'mouseout', function() {
            _this._over = false;
            _this._updateBg();
        });
        uki.image.load([this._focusImage], function() {
            _this._focusImage.style.cssText += ';margin-left:-' + _this._focusImage.width/2 + PX + ';margin-top:-' + _this._focusImage.height/2 + PX
        });
        this._initFocusable();
        this.disabled(this.disabled());
    },

    _focus: function() {
        this._dom.appendChild(this._focusImage);
        if (this._firstFocus) {
            var _this = this;
            uki.dom.bind(this._focusableInput, 'keyup', function(e) {
                if (e.which == 32 || e.which == 13) {
                    _this._click();
                    _this.trigger('click', {domEvent: e, source: _this});
                }
            });
        }
    },

    _blur: function() {
        this._dom.removeChild(this._focusImage);
    },

    typeName: function() {
        return 'uki.view.Checkbox';
    },

    _bindToDom: function(name) {
        return uki.view.Focusable._bindToDom.call(this, name) || Base._bindToDom.call(this, name);
    }

});

})();uki.view.TextField = uki.newClass(uki.view.Base, uki.view.Focusable, new function() {
    var Base = uki.view.Base[PROTOTYPE],
        emptyInputHeight = {},
        proto = this;

    function getEmptyInputHeight (fontSize) {
        if (!emptyInputHeight[fontSize]) {
            var node = uki.createElement('input', Base.defaultCss + "border:none;padding:0;border:0;overflow:hidden;font-size:"+fontSize+";left:-999em;top:0");
            uki.dom.probe(
                node,
                function(probe) {
                    emptyInputHeight[fontSize] = probe.offsetHeight;
                }
            );
        }
        return emptyInputHeight[fontSize];
    }

    function nativePlaceholder (node) {
        return typeof node.placeholder == 'string';
    }


    proto._setup = function() {
        Base._setup.apply(this, arguments);
        uki.extend(this, {
            _value: '',
            _multiline: false,
            _placeholder: '',
            _backgroundPrefix: '',
            defaultCss: Base.defaultCss + "margin:0;border:none;outline:none;padding:0;font-size:11px;left:2px;top:0;z-index:100;resize:none;background: url(" + uki.theme.imageSrc('x') + ")"
        });
    };

    proto.disabled = uki.newProp('_disabled', function(d) {
        this._disabled = d;
        if (d) this.blur();
        this._input.disabled = d;
    });

    proto.value = function(value) {
        if (value === undefined) return this._input.value;

        this._input.value = value;
        this._updatePlaceholderVis();
        return this;
    };

    proto.placeholder = uki.newProp('_placeholder', function(v) {
        this._placeholder = v;
        if (!this._multiline && nativePlaceholder(this._input)) {
            this._input.placeholder = v;
        } else {
            if (!this._placeholderDom) {
                this._placeholderDom = uki.createElement('div', this.defaultCss + 'z-input:101;color:#999;cursor:text', v);
                this._dom.appendChild(this._placeholderDom);
                this._updatePlaceholderVis();
                uki.each(['fontSize', 'fontFamily', 'fontWeight'], function(i, name) {
                    this._placeholderDom.style[name] = this[name]();
                }, this)

                var _this = this;
                uki.dom.bind(this._placeholderDom, 'mousedown', function() { _this.focus() });
            } else {
                this._placeholderDom.innerHTML = v;
            }
        }
    });

    var cssProps = ['fontSize', 'textAlign', 'color', 'fontFamily', 'fontWeight'];
    uki.each(cssProps, function(i, name) {
        proto[name] = function(v) {
            if (v === undefined) return this._input.style[name];
            this._input.style[name] = v;
            if (this._placeholderDom) {
                this._placeholderDom.style[name] = v;
            }
            return this;
        };
    });

    uki.addProps(proto, ['backgroundPrefix']);

    proto.defaultBackground = function() {
        return uki.theme.background(this._backgroundPrefix + 'input');
    };

    proto.typeName = function() {
        return 'uki.component.TextField';
    };

    proto._createDom = function() {
        var tagName = this._multiline ? 'textarea' : 'input';
        this._dom = uki.createElement('div', Base.defaultCss + ';cursor:text;overflow:visible;');
        this._input = uki.createElement(tagName, this.defaultCss + (this._multiline ? '' : ';overflow:hidden;'));
        this._inputStyle = this._input.style;

        this._input.value = this._value;
        this._dom.appendChild(this._input);

        this._input.value = this.value();

        this._initFocusable(this._input);
    };

    proto._layoutDom = function(rect) {
        Base._layoutDom.apply(this, arguments);
        uki.dom.layout(this._input.style, {
            width: this._rect.width - 4
        });
        var padding;
        if (this._multiline) {
            this._input.style.height = this._rect.height - 4 + 'px';
            this._input.style.top = 2 + 'px'
            padding = '2px 0';
        } else {
            var o = (this._rect.height - getEmptyInputHeight(this.fontSize())) / 2;
            padding = FLOOR(o) + 'px 0 ' + CEIL(o) + 'px 0';
            this._input.style.padding = padding;
        }
        if (this._placeholderDom) this._placeholderDom.style.padding = padding;
        if (this._firstLayout) this._initFocusable(this._input);
    };

    proto._recalcOffset = function() {
        if (this._multiline) return;
    };

    proto._updatePlaceholderVis = function() {
        if (this._placeholderDom) this._placeholderDom.style.display = this.value() ? 'none' : 'block';
    };

    proto._focus = function() {
        this._focusBackground = this._focusBackground || uki.theme.background(this._backgroundPrefix + 'input-focus');
        this._focusBackground.attachTo(this);
        if (this._placeholderDom) this._placeholderDom.style.display = 'none';
    };

    proto._blur = function() {
        this._focusBackground.detach();
        this._updatePlaceholderVis();
    };

    proto._bindToDom = function(name) {
        return uki.view.Focusable._bindToDom.call(this, name) || Base._bindToDom.call(this, name);
    };
});

uki.view.MultilineTextField = uki.newClass(uki.view.TextField, {
    _setup: function() {
        uki.view.TextField[PROTOTYPE]._setup.call(this);
        this._multiline = true;
    }
});uki.view.list = {};

uki.view.List = uki.newClass(uki.view.Base, uki.view.Focusable, new function() {
    var Base = uki.view.Base[PROTOTYPE],
        proto = this;

    proto.typeName = function() {
        return 'uki.view.List';
    };

    proto._throttle = 5; // do not try to render more often than every 5ms
    proto._visibleRectExt = 300;
    proto._packSize = 20;
    proto._defaultBackground = 'list';

    proto._setup = function() {
        Base._setup.call(this);
        uki.extend(this, {
            _rowHeight: 30,
            _render: new uki.view.list.Render(),
            _data: [],
            _selectedIndex: -1
        });
    };

    proto.defaultBackground = function() {
        return uki.theme.background('list', this._rowHeight);
    };

    uki.addProps(proto, ['rowHeight', 'render', 'packSize', 'visibleRectExt', 'throttle']);

    proto.rowHeight = uki.newProp('_rowHeight', function(val) {
        this._rowHeight = val;
        if (this._background) this._background.detach();
        this._background = null;
        if (this.background()) this.background().attachTo(this);
    });

    proto.data = function(d) {
        if (d === undefined) return this._data;
        this.selectedIndex(-1);
        this._data = d;
        this._updateRectOnDataChnage();
        return this;
    };

    // data api
    proto.addRow = function(position, data) {
        this.selectedIndex(-1);
        this._data.splice(position, 0, data);
        if (position < this._packs[0].itemFrom) {
            this._movePack(this._packs[0], 1);
            this._movePack(this._packs[1], 1);
        } else if (position < this._packs[1].itemTo && position >= this._packs[1].itemFrom) {
            this._addToPack(this._packs[1], position, data);
        } else if (position < this._packs[0].itemTo && position >= this._packs[0].itemFrom) {
            this._addToPack(this._packs[0], position, data);
            this._movePack(this._packs[1], 1);
            this._packs[1].itemTo++;
        }
        this._updateRectOnDataChnage();
    };

    proto.removeRow = function(position) {
        this.selectedIndex(-1);
        this._data.splice(position, 1);
        if (position < this._packs[0].itemFrom) {
            this._movePack(this._packs[0], -1);
            this._movePack(this._packs[1], -1);
        } else if (position < this._packs[1].itemTo && position >= this._packs[1].itemFrom) {
            this._removeFromPack(this._packs[1], position);
            this._layoutDom(this.rect());
        } else if (position < this._packs[0].itemTo && position >= this._packs[0].itemFrom) {
            this._removeFromPack(this._packs[0], position);
            this._movePack(this._packs[1], -1);
            this._packs[1].itemTo--;
            this._layoutDom(this.rect());
        }
        this._updateRectOnDataChnage();
    };

    proto.selectedIndex = function(position) {
        if (position === undefined) return this._selectedIndex;
        var nextIndex = MAX(0, MIN((this._data || []).length - 1, position));
        if (this._selectedIndex > -1) this._setSelected(this._selectedIndex, false);
        if (nextIndex == position) this._setSelected(this._selectedIndex = nextIndex, true);
        return this;
    };

    proto.layout = function() {
        this._layoutDom(this._rect);
        this._needsLayout = false;
        // send visibleRect with layout
        this.trigger('layout', { rect: this._rect, source: this, visibleRect: this._visibleRect });
        this._firstLayout = false;
    };


    proto._updateRectOnDataChnage = function() {
        if (this.rect()) {
            var newRect = this.rect().clone(),
                oldRect = this.rect(),
                h = this._data.length * this._rowHeight;

            if (!this._minSize || h > this._minSize.height) {
                newRect.height = h;
                this.rect(newRect);
            }
        }
    };

    proto._createDom = function() {
        this._dom = uki.createElement('div', this.defaultCss + 'overflow:hidden');

        var packDom = uki.createElement('div', 'position:absolute;left:0;top:0px;width:100%;overflow:hidden');
        this._packs = [
            {
                dom: packDom,
                itemTo: 0,
                itemFrom: 0
            },
            {
                dom: packDom.cloneNode(false),
                itemTo: 0,
                itemFrom: 0
            }
        ];
        this._dom.appendChild(this._packs[0].dom);
        this._dom.appendChild(this._packs[1].dom);

        var _this = this;

        this.bind('mousedown', function(e) {
            var o = uki.dom.offset(this._dom),
                y = e.domEvent.pageY - o.y,
                p = FLOOR(y / this._rowHeight);
            this.selectedIndex(p);
        });
    };

    proto._setSelected = function(position, state) {
        var item = this._itemAt(position);
        if (!item) return;
        if (state) this._scrollToPosition(position);
        this._render.setSelected(item, this._data[position], state, this.hasFocus());
    };

    proto._scrollToPosition = function(position) {
        var maxY, minY;
        maxY = (position+1)*this._rowHeight;
        minY = position*this._rowHeight;
        if (maxY >= this._visibleRect.maxY()) {
            this._scrollableParent.scroll(0, maxY - this._visibleRect.maxY());
        } else if (minY < this._visibleRect.y) {
            this._scrollableParent.scroll(0, minY - this._visibleRect.y);
        }
        this.layout();
    };

    proto._itemAt = function(position) {
        if (position < this._packs[1].itemTo && position >= this._packs[1].itemFrom) {
            return this._packs[1].dom.childNodes[position - this._packs[1].itemFrom];
        } else if (position < this._packs[0].itemTo && position >= this._packs[0].itemFrom) {
            return this._packs[0].dom.childNodes[position - this._packs[0].itemFrom];
        }
        return null;
    };

    proto._focus = function() {
        this._selectedIndex = this._selectedIndex > -1 ? this._selectedIndex : 0;
        this._setSelected(this._selectedIndex, true);
        if (this._firstFocus) {
            var _this = this,
                userAgent = navigator.userAgent.toLowerCase(),
                useKeyPress = /mozilla/.test( userAgent ) && !(/(compatible|webkit)/).test( userAgent );
            uki.dom.bind(this._focusableInput, useKeyPress ? 'keypress' : 'keydown', function(e) {
                if (e.which == 38 || e.keyCode == 38) { // UP
                    _this.selectedIndex(MAX(0, _this.selectedIndex() - 1));
                    uki.dom.preventDefault(e);
                } else if (e.which == 40 || e.keyCode == 40) { // DOWN
                    _this.selectedIndex(MIN(_this._data.length-1, _this.selectedIndex() + 1));
                    uki.dom.preventDefault(e);
                }
            });
        }

    };

    proto._blur = function() {
        if (this._selectedIndex > -1) { this._setSelected(this._selectedIndex, true); }
    };

    proto._rowCss = function() {
        return ['width:100%;height:', this._rowHeight, 'px;overflow:hidden'].join('');
    };

    proto._renderPack = function(pack, itemFrom, itemTo) {
        var html = [], position,
            rect = new Rect(0, itemFrom*this._rowHeight, this.rect().width, this._rowHeight);
        for (i=itemFrom; i < itemTo; i++) {
            html[html.length] = [
                '<div style="', this._rowCss(), '">',
                this._render.render(this._data[i], rect, i),
                '</div>'
            ].join('');
            rect.y += this._rowHeight;
        };
        pack.dom.innerHTML = html.join('');
        pack.itemFrom = itemFrom;
        pack.itemTo   = itemTo;
        pack.dom.style.top = itemFrom*this._rowHeight + 'px';
        if (this._selectedIndex > pack.itemFrom && this._selectedIndex < pack.itemTo) {
            position = this._selectedIndex - pack.itemFrom;
            this._render.setSelected(this._itemAt(position), this._data[position], true, this.hasFocus());
        }
    };

    proto._addToPack = function (pack, position, data) {
        var row = this._createRow(data),
            nextChild = pack.dom.childNodes[position - pack.itemFrom];
        nextChild ? pack.dom.insertBefore(row, nextChild) : pack.dom.appendChild(row);
        pack.itemTo++;
    };

    proto._removeFromPack = function(pack, position) {
        pack.dom.removeChild(pack.dom.childNodes[position - pack.itemFrom]);
        pack.itemTo--;
    };

    proto._movePack = function(pack, offset) {
        pack.itemFrom += offset;
        pack.dom.style.top = pack.itemFrom * this._rowHeight + 'px';
    };

    proto._createRow = function(data) {
        return uki.createElement('div', this._rowCss(), this._render.render(data));
    };

    proto._swapPacks = function() {
        var tmp = this._packs[0];
        this._packs[0] = this._packs[1];
        this._packs[1] = tmp;
    };

    proto._normalizeRect = function(rect) {
        rect = Base._normalizeRect.call(this, rect);
        if (rect.height < this._rowHeight * this._data.length) {
            rect = new Rect(rect.x, rect.y, rect.width, this._rowHeight * this._data.length);
        }
        return rect;
    };

    proto._layoutDom = function(rect) {
        if (this._firstLayout) {
            var _this = this;
            this._scrollableParent = uki.view.scrollableParent(this);

            this._scrollableParent.bind('scroll', function() {
                if (_this._throttle) {
                    if (_this._throttleStarted) return;
                    _this._throttleStarted = true;
                    setTimeout(function() {
                        _this._throttleStarted = false;
                        _this.layout();
                    }, _this._throttle);
                } else {
                    _this.layout();
                }
            });

            this._initFocusable();
        }


        this._visibleRect = uki.view.visibleRect(this, this._scrollableParent);
        Base._layoutDom.call(this, rect);

        var totalHeight = this._rowHeight * this._data.length,
            prefferedPackSize = CEIL((this._visibleRect.height + this._visibleRectExt*2) / this._rowHeight),

            minVisibleY  = MAX(0, this._visibleRect.y - this._visibleRectExt),
            maxVisibleY  = MIN(totalHeight, this._visibleRect.maxY() + this._visibleRectExt),
            minRenderedY = this._packs[0].itemFrom * this._rowHeight,
            maxRenderedY = this._packs[1].itemTo * this._rowHeight,

            itemFrom, itemTo, startAt;

        if (
            maxVisibleY <= minRenderedY || minVisibleY >= maxRenderedY || // both packs below/above visible area
            (maxVisibleY > maxRenderedY && this._packs[1].itemFrom * this._rowHeight > this._visibleRect.y) || // need to render below, and pack 2 is not enough to cover
            (minVisibleY < minRenderedY && this._packs[0].itemTo * this._rowHeight < this._visibleRect.maxY()) // need to render above, and pack 1 is not enough to cover the area
            // || prefferedPackSize is not enougth to cover the area above/below, can this actualy happen?
        ) {
            // this happens a) on first render b) on scroll jumps c) on container resize
            // render both packs, move them to be at the center of visible area
            startAt = minVisibleY + (maxVisibleY - minVisibleY - prefferedPackSize*this._rowHeight*2) / 2;
            itemFrom = MAX(0, Math.round(startAt / this._rowHeight));
            itemTo = MIN(this._data.length, itemFrom + prefferedPackSize);

            this._renderPack(this._packs[0], itemFrom, itemTo);
            this._renderPack(this._packs[1], itemTo, MIN(this._data.length, itemTo + prefferedPackSize));
        } else if (maxVisibleY > maxRenderedY) { // we need to render below current area
            // this happens on normal scroll down
            // rerender bottom, swap
            itemFrom = this._packs[1].itemTo;
            itemTo   = MIN(this._data.length, this._packs[1].itemTo + prefferedPackSize);

            this._renderPack(this._packs[0], itemFrom, itemTo);
            this._swapPacks();
        } else if (minVisibleY < minRenderedY) { // we need to render above current area
            // this happens on normal scroll up
            // rerender top, swap
            itemFrom = MAX(this._packs[0].itemFrom - prefferedPackSize, 0);
            itemTo   = this._packs[0].itemFrom;

            this._renderPack(this._packs[1], itemFrom, itemTo);
            this._swapPacks();
        }

        if (this._focusableInput) this._focusableInput.style.top = this._visibleRect.y + 'px'; // move to reduce on focus jump

    };

    proto._bindToDom = function(name) {
        return uki.view.Focusable._bindToDom.call(this, name) || Base._bindToDom.call(this, name);
    };
});

uki.fn.data = function( value ) { return this.attr( 'data', value ); };


/**
 * Flyweight view rendering
 * Used in lists, tables, grids
 */
uki.view.list.Render = uki.newClass({
    init: function() {},

    /**
     * Renders data to an html string
     * @param Object data Data to render
     * @return String html
     */
    render: function(data, rect, i) {
        return '<div style="line-height: 30px; text-align: center; font-size: 12px">' + data + '</div>';
    },

    setSelected: function(container, data, state, focus) {
        container.style.backgroundColor = state && focus ? '#3875D7' : state ? '#CCC' : '';
        container.style.color = state && focus ? '#FFF' : '#000';
    }

    /**
     * Layouts a component with a given container
     * Only one instance of component should reside within the container
     *
     * @param uki.geometry.Rect rect
     * @param Node container
     */
    // layout: function(rect, container) {
    //
    // }
});
uki.view.table = {};

uki.view.Table = uki.newClass(uki.view.Container, new function() {
    var proto = this,
        Base = uki.view.Container[PROTOTYPE],
        propertiesToDelegate = ['rowHeight', 'data', 'packSize', 'visibleRectExt'];

    proto.typeName = function() { return 'uki.view.Table'; };
    proto._rowHeight = 17;
    proto._headerHeight = 21;
    proto.defaultCss = Base.defaultCss + 'overflow:hidden;';

    uki.each(propertiesToDelegate, function(i, name) { uki.delegateProp(proto, name, '_list'); });

    proto.columns = uki.newProp('_columns', function(c) {
        this._columns = uki.build(c);
        this._totalWidth = 0;
        for (var i=0; i < this._columns.length; i++) {
            this._columns[i].position(i);
            this._totalWidth += this._columns[i].width();
        };
        this._list.minSize(new Size(this._totalWidth, 0));
        this._header.minSize(new Size(this._totalWidth, 0));
        this._header.columns(this._columns);
    });

    proto._createDom = function() {
        Base._createDom.call(this);
        var scrollPaneRect = new Rect(0, this._headerHeight, this.rect().width, this.rect().height - this._headerHeight),
            listRect = scrollPaneRect.clone().normalize(),
            headerRect = new Rect(0, 0, this.rect().width, this._headerHeight),
            listML = { view: 'List', rect: listRect, anchors: 'left top bottom right', render: new uki.view.table.Render(this), className: 'table-list' },
            paneML = { view: 'ScrollPane', rect: scrollPaneRect, anchors: 'left top right bottom', scrollableH: true, childViews: [listML], className: 'table-scroll-pane'},
            headerML = { view: 'table.Header', rect: headerRect, anchors: 'top left right', className: 'table-header' };

        uki.each(propertiesToDelegate, function(i, name) {
            if (this['_' + name] !== undefined) listML[name] = this['_' + name];
        }, this);
        this._scrollPane = uki.build(paneML)[0];
        this._list = this._scrollPane.childViews()[0];
        this._header = uki.build(headerML)[0];
        this._scrollPane.resizeToContents();
        this.appendChild(this._header);
        this.appendChild(this._scrollPane);

        var _this = this;
        this._scrollPane.bind('scroll', function() {
            // this is kinda wrong but faster than colling rect() + layout()
            _this._header.dom().style.left = -_this._scrollPane.scrollLeft() + 'px';
        });
    };


});

uki.view.table.Render = uki.newClass(uki.view.list.Render, new function() {

    var proto = this;

    proto.init = function(table) {
        this._table = table;
    };

    proto.render = function(row, rect, i) {
        if (!this._template) this._template = this._buildTemplate(rect);
        var table = this._table,
            columns = table.columns();
        this._template[1] = uki.map(columns, function(val, j) {
            return columns[j].render(row, rect, i);
        }).join('');
        return this._template.join('');
    };

    proto._buildTemplate = function(rect) {
        var tagOpening = ['<div style="position:relwidth:100%;height:', rect.height, 'px">'].join(',');
        return ['', '', '']
    };
});uki.view.table.Column = uki.newClass(new function() {
    var proto = this;

    proto._width = 100;
    proto._offset = 0;
    proto._position = 0;
    proto._css = 'overflow:hidden;float:left;font-size:11px;line-height:11px;white-space:nowrap;text-overflow:ellipsis;';
    proto._inset = new Inset(3, 5);

    proto.init = function() {};

    uki.addProps(proto, ['width', 'position', 'css', 'formatter', 'label']);

    proto.inset = uki.newProp('_inset', function(i) {
        this._inset = Inset.create(i);
    });

    proto.render = function(row, rect, i) {
        if (!this._template) this._template = this._buildTemplate(rect);
        this._template[1] = this._formatter ? this._formatter(row[this._position]) : row[this._position];
        return this._template.join('')
    };

    proto.renderHeader = function(height) {
        if (!this._headerTemplate) this._headerTemplate = this._buildHeaderTemplate(height);
        var template = this._headerTemplate;
        template[1] = this.label();
        return template.join('');
    };

    proto._buildHeaderTemplate = function(headerHeight) {
        uki.dom.offset.initializeBoxModel();
        var border = 'border:1px solid #CCC;border-top: none;border-left:none;',
            inset  = this._inset,
            padding = ['padding:', inset.top, 'px ', inset.right, 'px ', inset.bottom, 'px ', inset.left, 'px;'].join(''),
            height = 'height:' + (headerHeight - (uki.dom.offset.boxModel ? inset.height() + 1 : 0)) + 'px;',
            width = 'width:' + (this._width - (uki.dom.offset.boxModel ? inset.width() + 1 : 0)) + 'px;',
            tagOpening = ['<div style="', width, border, padding, height, this._css, '">'].join('');

        return [tagOpening, '', '</div>'];
    };

    proto._buildTemplate = function(rect) {
        uki.dom.offset.initializeBoxModel();
        var border = 'border-right:1px solid #CCC;',
            inset = this._inset,
            padding = ['padding:', inset.top, 'px ', inset.right, 'px ', inset.bottom, 'px ', inset.left, 'px;'].join(''),
            height = 'height:' + (rect.height - (uki.dom.offset.boxModel ? inset.height() : 0)) + 'px;',
            width = 'width:' + (this._width - (uki.dom.offset.boxModel ? inset.width() + 1 : 0)) + 'px;',
            // left = 'left:' + this._offset + 'px;',
            tagOpening = ['<div style="', width, border, padding, height, this._css, '">'].join('');
        return [tagOpening, '', '</div>'];
    };
});

uki.view.table.NumberColumn = uki.newClass(uki.view.table.Column, new function() {
    var Base = uki.view.table.Column[PROTOTYPE],
        proto = this;

    proto._css = Base._css + 'text-align:right;';
});

uki.view.table.CustomColumn = uki.view.table.Column;uki.view.table.Header = uki.newClass(uki.view.Label, new function() {
    var Base = uki.view.Label[PROTOTYPE],
        proto = this;

    proto._setup = function() {
        Base._setup.call(this);
        this._multiline = true;
    };

    proto.columns = uki.newProp('_columns', function(v) {
        this._columns = v;
        this.html(this._createColumns());
    });

    proto._template = ['<div style="width:',0,'px;height:',0,'px;float:left;overflow:hidden;">',0,'</div>'];
    proto._createColumns = function() {
        var html = [],
            template = this._template;
        for(var i = 0, offset = 0, columns = this._columns, l = columns.length; i < l; i++) {
            html[html.length] = columns[i].renderHeader(this.rect().height);
        }
        return html.join('')
    };
});(function() {

var Base = uki.view.Base[PROTOTYPE],
self = uki.view.Slider = uki.newClass(uki.view.Base, uki.view.Focusable, {

    _setup: function() {
        Base._setup.call(this);
        uki.extend(this, {
            _min: 0,
            _max: 1,
            _value: 0,
            _values: null,
            _keyStep: 0.01
        });
    },

    min: uki.newProp('_min'),
    max: uki.newProp('_max'),
    values: uki.newProp('_values'),
    keyStep: uki.newProp('_keyStep'),

    value: function(val) {
        if (val === undefined) return this._value;
        this._value = MAX(this._min, MIN(this._max, val));
        this._position = this._val2pos(this._value);
        this._moveHandle();
        this.trigger('change', {source: this, value: this._value});
        return this;
    },

    _pos2val: function(pos) {
        return pos / this._rect.width * (this._max - this._min);
    },

    _val2pos: function(val) {
        return val / (this._max - this._min) * this._rect.width;
    },

    _createDom: function() {
        this._dom = uki.createElement('div', Base.defaultCss + 'height:18px;-moz-user-select:none;-webkit-user-select:none;overflow:visible;');
        this._handle = uki.createElement('div', Base.defaultCss + 'overflow:hidden;cursor:default;background:url(' + uki.theme.image('x').src + ')');
        this._bg = uki.theme.image('slider-handle');
        this._focusBg = uki.theme.image('slider-focus');
        this._focusBg.style.cssText += this._bg.style.cssText += Base.defaultCss + 'top:0;left:0;z-index:-1;position:absolute;';
        this._handle.appendChild(this._bg);

        var _this = this;
        uki.image.load([this._bg, this._focusBg], function() { _this._afterHandleLoad(); });

        uki.theme.background('slider-bar').attachTo(this);
    },

    _afterHandleLoad: function() {
        this._focusBg.style.cssText += ';z-index:10;margin-left:-' + (this._focusBg.width) / 2 + 'px;margin-top:-' + (this._focusBg.height-(this._bg.height/2)+1)/2 + 'px;';
        this._handle.style.cssText += ';margin-left:-' + (this._bg.width / 2) + 'px;width:' +this._bg.width + 'px;height:' + (this._bg.height / 2) + 'px;';
        this._dom.appendChild(this._handle);
        var _this = this;
        uki.dom.bind(this._handle, 'dragstart', function(e) { e.returnValue = false; });

        uki.dom.bind(this._handle, 'mousedown', function(e) {
            _this._dragging = true;
            _this._initialPosition = new Point(parseInt(_this._handle.style.left, 10), parseInt(_this._handle.style.top, 10));
            uki.dom.drag.start(_this, e);
        });

        uki.dom.bind(this._handle, 'mouseover', function() {
            _this._over = true;
            _this._bg.style.top = - _this._bg.height / 2 + 'px';
        });

        uki.dom.bind(this._handle, 'mouseout', function() {
            _this._over = false;
            _this._bg.style.top = _this._dragging ? (- _this._bg.height / 2 + 'px') : 0;
        });

        uki.dom.bind(this._dom, 'click', function(e) {
            var x = e.pageX - uki.dom.offset(_this._dom).x;
            _this.value(_this._pos2val(x));
        });
    },

    _moveHandle: function() {
        // this._focusBg.style.left = this._handle.style.left = 100.0*this._position/this._rect.width + '%';
        this._focusBg.style.left = this._handle.style.left = this._position + 'px';
    },

    _drag: function(e, offset) {
        this._position = MAX(0, MIN(this._rect.width, this._initialPosition.x - offset.x));
        this._value = this._pos2val(this._position);
        this._moveHandle();
        this.trigger('change', {source: this, value: this._value});
    },

    _drop: function(e, offset) {
        this._dragging = false;
        this._initialPosition = null;
        if (!this._over) this._bg.style.top = 0;
        this._value = this._pos2val(this._position);
    },

    _focus: function() {
        this._dom.appendChild(this._focusBg);
        this._focusBg.style.left = this._handle.style.left;
        if (this._firstFocus) {
            var _this = this;
            uki.dom.bind(this._focusableInput, 'keydown', function(e) {
                if (e.which == 39) {
                    _this.value(_this.value() + _this._keyStep * (_this._max - _this._min));
                } else if (e.which == 37) {
                    _this.value(_this.value() - _this._keyStep * (_this._max - _this._min));
                }
            });
        }
    },

    _blur: function() {
        this._dom.removeChild(this._focusBg);
    },

    _layoutDom: function(rect) {
        var fixedRect = rect.clone();
        fixedRect.height = 18;
        Base._layoutDom.call(this, fixedRect);
        this._position = this._val2pos(this._value);
        this._moveHandle();
        if (this._firstLayout) this._initFocusable();
        return true;
    },

    typeName: function() {
        return 'uki.view.Slider';
    },

    _bindToDom: function(name) {
        if (name == 'change') return true;
        return uki.view.Focusable._bindToDom.call(this, name) || Base._bindToDom.call(this, name);
    }

});

})();uki.view.SplitPane = uki.newClass(uki.view.Container, new function() {
    var Base = uki.view.Container[PROTOTYPE],
        proto = this;

    proto.typeName = function() {
        return 'uki.view.SplitPane';
    };

    proto._setup = function() {
        Base._setup.call(this);
        this._originalRect = this._rect;
        uki.extend(this, {
            _vertical: false,
            _handlePosition: 200,
            _autogrowLeft: false,
            _autogrowRight: true,
            _handleWidth: 7,
            _leftMin: 100,
            _rightMin: 100,

            _panes: []
        });
    };

    proto.handlePosition = uki.newProp('_handlePosition', function(val) {
        this._handlePosition = this._normalizePosition(val);
        this.trigger('handleMove', {source: this, handlePosition: this._handlePosition, dragValue: val });
        this._resizeChildViews();
    });

    proto.handleWidth = uki.newProp('_handleWidth', function(val) {
        if (this._handleWidth != val) {
            this._handleWidth = val;
            var handle = this._createHandle();
            this._dom.insertBefore(handle, this._handle);
            this._removeHandle()
            this._handle = handle;
            this._resizeChildViews();
        }
    });


    proto._normalizePosition = function(val) {
        var prop = this._vertical ? 'height' : 'width';
        return MAX(
                this._leftMin,
                MIN(
                    this._rect[prop] - this._rightMin - this._handleWidth,
                    MAX(0, MIN(this._rect ? this._rect[prop] : 1000, val * 1))
                ));
    };


    uki.addProps(proto, ['leftMin', 'rightMin', 'autogrowLeft', 'autogrowRight']);
    proto.topMin = proto.leftMin;
    proto.bottomMin = proto.rightMin;

    proto._removeHandle = function() {
        this._dom.removeChild(this._handle);
    };

    proto._createHandle = function() {
        var handle, _this = this;
        if (this._vertical) {
            handle = uki.theme.dom('splitPane-vertical', {handleWidth: this._handleWidth});
            handle.style.top = this._handlePosition + 'px';
        } else {
            handle = uki.theme.dom('splitPane-horizontal', {handleWidth: this._handleWidth});
            handle.style.left = this._handlePosition + 'px';
        }
        if (!handle.style.cursor || window.opera) handle.style.cursor = this._vertical ? 'n-resize' : 'e-resize';

        uki.dom.bind(handle, 'dragstart', function(e) { e.returnValue = false; });

        uki.dom.bind(handle, 'mousedown', function(e) {
            var offset = uki.dom.offset(_this.dom());
            _this._posWithinHandle = (e[_this._vertical ? 'pageY' : 'pageX'] - offset[_this._vertical ? 'y' : 'x']) - _this._handlePosition;
            uki.dom.drag.start(_this, e);
        });

        return handle;
    };

    proto._createDom = function() {
        this._dom = uki.createElement('div', Base.defaultCss);
        for (var i=0, paneML; i < 2; i++) {
            paneML = { view: 'Container' };
            paneML.anchors = i == 1         ? 'left top bottom right' :
                             this._vertical ? 'left top right' :
                                              'left top bottom';
            paneML.rect = i == 0 ? this._leftRect() : this._rightRect();
            this._panes[i] = uki.build(paneML)[0];
            this.appendChild(this._panes[i]);
        };
        this._dom.appendChild(this._handle = this._createHandle());
    };

    proto._normalizeRect = function(rect) {
        rect = Base._normalizeRect.call(this, rect);
        var newRect = rect.clone();
        if (this._vertical) {
            newRect.height = MAX(newRect.height, this._leftMin + this._rightMin); // force min width
        } else {
            newRect.width = MAX(newRect.width, this._leftMin + this._rightMin); // force min width
        }
        return newRect;
    };

    proto._resizeSelf = function(newRect) {
        var oldRect = this._rect,
            dx, prop = this._vertical ? 'height' : 'width';
        if (!Base._resizeSelf.call(this, newRect)) return false;
        if (this._autogrowLeft) {
            dx = newRect[prop] - oldRect[prop];
            this._handlePosition = this._normalizePosition(this._handlePosition + (this._autogrowRight ? dx / 2 : dx));
        }
        if (this._vertical) {
            if (newRect.height - this._handlePosition < this._rightMin) {
                this._handlePosition = MAX(this._leftMin, newRect.height - this._rightMin);
            }
        } else {
            if (newRect.width - this._handlePosition < this._rightMin) {
                this._handlePosition = MAX(this._leftMin, newRect.width - this._rightMin);
            }
        }
        return true;
    };

    proto._drag = function(e) {
        var offset = uki.dom.offset(this.dom());
        this.handlePosition(e[this._vertical ? 'pageY' : 'pageX'] - offset[this._vertical ? 'y' : 'x'] - this._posWithinHandle);
        e.preventDefault ? e.preventDefault() : e.returnValue = false;
        this.layout();
    };

    proto._drop = function(e, offset) {
    };

    proto.topPane = proto.leftPane = function(pane) {
        return this._paneAt(0, pane);
    };

    proto.bottomPane = proto.rightPane = function(pane) {
        return this._paneAt(1, pane);
    };

    proto.topChildViews = proto.leftChildViews = function(views) {
        return this._childViewsAt(0, views);
    };

    proto.bottomChildViews = proto.rightChildViews = function(views) {
        return this._childViewsAt(1, views);
    };

    proto._childViewsAt = function(i, views) {
        if (views === undefined) return this._panes[i].childViews();
        this._panes[i].childViews(views);
        return this;
    };

    proto._paneAt = function(i, pane) {
        if (pane === undefined) return this._panes[i];
        uki.build.copyAttrs(this._panes[i], pane);
        return this;
    };

    proto._leftRect = function() {
        if (this._vertical) {
            return new Rect(this._rect.width, this._handlePosition);
        } else {
            return new Rect(this._handlePosition, this._rect.height);
        }
    };

    proto._rightRect = function() {
        if (this._vertical) {
            return new Rect(
                0, this._handlePosition + this._handleWidth,
                this._rect.width, this._rect.height - this._handleWidth - this._handlePosition
            );
        } else {
            return new Rect(
                this._handlePosition + this._handleWidth, 0,
                this._rect.width - this._handleWidth - this._handlePosition, this._rect.height
            );
        }
    };

    proto._resizeChildViews = function() {
        this._panes[0].rect(this._leftRect());
        this._panes[1].rect(this._rightRect());
    };

    proto._layoutDom = function(rect) {
        Base._layoutDom.call(this, rect);
        this._handle.style[this._vertical ? 'top' : 'left'] = this._handlePosition + 'px';
    };

    proto._bindToDom = function(name) {
        if (name == 'handleMove') return true;
        return Base._bindToDom.call(this, name);
    };

});

uki.view.HorizontalSplitPane = uki.view.SplitPane;
uki.view.VerticalSplitPane = uki.newClass(uki.view.SplitPane, {
    _setup: function() {
        uki.view.SplitPane[PROTOTYPE]._setup.call(this);
        this._vertical = true;
    }
});

uki.fn.handlePosition = function( value ) { return this.attr( 'handlePosition', value ); };uki.view.ScrollPane = uki.newClass(uki.view.Container, new function() {
    /**
     * Scroll pane. Pane with scrollbars with content overflowing the borders.
     * Works consisently across all supported browsers.
     */
    var Base = uki.view.Container[PROTOTYPE],
        proto = this,
        scrollWidth,
        widthIncludesScrollBar;

    function initScrollWidth () {
        if (!scrollWidth) {
            uki.dom.probe(
                uki.createElement(
                    'div',
                    'position:absolute;left:-99em;width:100px;height:100px;overflow:scroll;',
                    '<div style="position:absolute;left:0;width:100%;"></div>'
                ),
                function( probe ) {
                    scrollWidth = probe.offsetWidth - probe.clientWidth;
                    widthIncludesScrollBar = probe.firstChild.offsetWidth == 100;
                }
            );
        }
    }

    proto.typeName = function() {
        return 'uki.view.ScrollPane';
    };

    proto._setup = function() {
        Base._setup.call(this);

        uki.extend(this, {
            _clientRect: this.rect().clone(), // rect with scroll bars excluded
            _rectForChild: this.rect().clone(),
            _clientRectValid: false,
            _scrollableV: true,
            _scrollableH: false,
            _scrollV: false,
            _scrollH: false
        });
    };

    uki.addProps(proto, ['scrollableV', 'scrollableH']);

    this.rectForChild = function() { return this._rectForChild; };
    this.clientRect = function() { return this._clientRect; };

    this.scroll = function(dx, dy) {
        if (dx) this.scrollTop(this.scrollLeft() + dy);
        if (dy) this.scrollTop(this.scrollTop() + dy);
    };

    uki.each(['appendChild', 'removeChild', 'childViews'], function(i, name) {
        proto[name] = function(arg) {
            this._clientRectValid = false;
            return Base[name].call(this, arg);
        };
    });

    uki.delegateProp(proto, 'scrollTop', '_dom');
    uki.delegateProp(proto, 'scrollLeft', '_dom');

    proto.visibleRect = function() {
        var tmpRect = this._clientRect.clone();
        tmpRect.x = this.rect().x + this.scrollLeft();
        tmpRect.y = this.rect().y + this.scrollTop();
        return tmpRect;
    };

    proto.rect = function(newRect) {
        if (newRect === undefined) return this._rect;

        newRect = Rect.create(newRect);
        var oldRect = this._rect;
        this._parentRect = newRect;
        if (!this._resizeSelf(newRect)) return this;
        this._clientRectValid = false;
        this._updateClientRects();
        this._needsLayout = true;
        this.trigger('resize', {oldRect: oldRect, newRect: this._rect, source: this});
        return this;
    };

    proto._recalcClientRects = function() {
        initScrollWidth();
        this._clientRectValid = true;

        var cw = this.contentsWidth(),
            ch = this.contentsHeight(),
            sh = this._scrollableH ? cw > this._rect.width : false,
            sv = this._scrollableV ? ch > this._rect.height : false;

        this._scrollH = sh;
        this._scrollV = sv;
        this._clientRect = new Rect( this._rect.width +  (sv ? -1 : 0) * scrollWidth,
                                     this._rect.height + (sh ? -1 : 0) * scrollWidth );
        this._rectForChild = new Rect( this._rect.width +  (sv && !widthIncludesScrollBar ? -1 : 0) * scrollWidth,
                                       this._rect.height + (sh && !widthIncludesScrollBar ? -1 : 0) * scrollWidth );
    };

    proto._updateClientRects = function() {
        if (this._clientRectValid) return;

        var oldClientRect = this._clientRect;
        this._recalcClientRects();

        if (oldClientRect.width != this._clientRect.width || oldClientRect.height != this._clientRect.height) {
            this._resizeChildViews(oldClientRect);
        }
    };

    proto._resizeChildViews = function(oldClientRect) {
        for (var i=0, childViews = this.childViews(); i < childViews.length; i++) {
            childViews[i].parentResized(oldClientRect, this._clientRect);
        };
    };

    proto._layoutChildViews = function() {
        for (var i=0, childViews = this.childViews(); i < childViews.length; i++) {
            if (childViews[i]._needsLayout && childViews[i].visible()) {
                childViews[i].layout();
            }
        };
    };

    proto._layoutDom = function(rect) {
        this._updateClientRects();

        Base._layoutDom.call(this, rect);

        if (this._layoutScrollH !== this._scrollH) {
            this._dom.style.overflowX = this._scrollH ? 'scroll' : 'hidden';
            this._layoutScrollH = this._scrollH;
        }

        if (this._layoutScrollV !== this._scrollV) {
            this._dom.style.overflowY = this._scrollV ? 'scroll' : 'hidden';
            this._layoutScrollV = this._scrollV;
        }

        // force redraw in ie
        if (this._dom.attachEvent) this._dom.className += '';
    };
});
uki.view.Popup = uki.newClass(uki.view.Container, new function() {
    var Base = uki.view.Container[PROTOTYPE],
        proto = this;

    proto._setup = function() {
        Base._setup.call(this);
        uki.extend(this, {
            _offset: 2,
            _relativeTo: null,
            _horizontal: false,
            _flipOnResize: true,
            _shadow: null,
            _defaultBackground: 'popup-normal',
            _defaultShadow: 'popup-shadow'
        });
    };

    proto._createDom = function() {
        Base._createDom.call(this);
        this.hideOnClick(true);
    };

    uki.addProps(proto, ['offset', 'relativeTo', 'horizontal', 'flipOnResize']);

    this.hideOnClick = function(state) {
        if (state === undefined) return this._clickHandler;
        if (state != !!this._clickHandler) {
            if (state) {
                var _this = this;
                this._clickHandler = function(e) {
                    if (uki.dom.contains(_this._relativeTo.dom(), e.target)) return;
                    if (uki.dom.contains(_this.dom(), e.target)) return;
                    _this.hide();
                };
                uki.dom.bind(doc.body, 'mousedown', this._clickHandler);
                uki.dom.bind(root, 'resize', this._clickHandler);
            } else {
                uki.dom.unbind(doc.body, 'mousedown', this._clickHandler);
                uki.dom.unbind(root, 'resize', this._clickHandler);
                this._clickHandler = false;
            }
        }
        return this;
    };

    proto.typeName = function() {
        return 'uki.view.Popup';
    };

    proto.toggle = function() {
        if (this.parent() && this.visible()) {
            this.hide();
        } else {
            this.show();
        }
    };

    proto.show = function() {
        this.visible(true);
        if (!this.parent()) {
            new uki.Attachment( root, this );
        } else {
            this.rect(this._recalculateRect());
            this.layout(this._rect);
        }
    };

    proto.hide = function() {
        this.visible(false);
    };

    proto.parentResized = function() {
        this.rect(this._recalculateRect());
    };

    proto._resizeSelf = function(newRect) {
        this._rect = this._normalizeRect(newRect);
        return true;
    };

    proto._layoutDom = function(rect) {
        return Base._layoutDom.call(this, rect);
    };

    proto._recalculateRect = function() {
        if (!this.visible()) return this._rect;
        var relativeOffset = uki.dom.offset(this._relativeTo.dom()),
            relativeRect = this._relativeTo.rect(),
            rect = this.rect().clone(),
            attachment = uki.view.top(this),
            attachmentRect = attachment.rect(),
            attachmentOffset = uki.dom.offset(attachment.dom()),
            position = new Point(),
            hOffset = this._horizontal ? this._offset : 0,
            vOffset = this._horizontal ? 0 : this._offset;

        relativeOffset.offset(-attachmentOffset.x, -attachmentOffset.y);

        if (this._anchors & ANCHOR_RIGHT) {
            position.x = relativeOffset.x + relativeRect.width - (this._horizontal ? 0 : rect.width) + hOffset;
        } else {
            position.x = relativeOffset.x - (this._horizontal ? rect.width : 0) - hOffset;
        }

        if (this._anchors & ANCHOR_BOTTOM) {
            position.y = relativeOffset.y + (this._horizontal ? relativeRect.height : 0) - rect.height - vOffset;
        } else {
            position.y = relativeOffset.y + (this._horizontal ? 0 : relativeRect.height) + vOffset;
        }

        return new Rect(position.x, position.y, rect.width, rect.height);
    };
});

uki.each(['show', 'hide', 'toggle'], function(i, name) {
    uki.fn[name] = function() {
        this.each(function() { this[name](); });
    };
});uki.view.Flow = uki.newClass(uki.view.Container, new function() {
    var Base = uki.view.Container[PROTOTYPE],
        proto = this;


    proto._setup = function() {
        Base._setup.call(this);
        uki.extend(this, {
            _horizontal: false,
            _dimension: 'height',
            _containers: [],
            _containerSizes: [],
            defaultCss: this.defaultCss + 'overflow:hidden;'
        });
    };

    proto.typeName = function() {
        return 'uki.view.Flow';
    };

    proto.horizontal = uki.newProp('_horizontal', function(h) {
        this._dimension = h ? 'width' : 'height';
    });

    proto.appendChild = function(view) {
        var container = this._createContainer(view),
            v = view.rect()[this._dimension],
            viewIndex = this._childViews.length;
        this._containers[viewIndex] = container;
        this._containerSizes[viewIndex] = 0;
        this._initContainer(container);
        container.style[this._dimension] = v + 'px';
        this._dom.appendChild(container);
        this._containerSizes[viewIndex] = v;
        // this._updateSize(v);

        Base.appendChild.call(this, view);
    };

    proto.removeChild = function(view) {
        // this._updateSize(-view.rect()[this._dimension]);
        var container = this._containers[view._viewIndex];
        Base.removeChild.call(this, view);

        this._dom.removeChild(container);
        this._containers = uki.grep(this._containers, function(c) { return c != container; });
    };

    proto.domForChild = function(child) {
        return this._containers[child._viewIndex];
    };

    proto.contentsSize = function() {
        var d = this._dimension,
            value = uki.reduce(0, this._childViews, function(sum, e) { return sum + e.rect()[d]; } );
        if (this._horizontal) {
            return new Size( value, this.contentsHeight() );
        } else {
            return new Size( this.contentsWidth(), value );
        }
    };

    proto._updateSize = function(dv) {
        var rect = this.rect(),
            width, height;

        if (this._horizontal) {
            width = rect.width + dv;
            height = rect.height;
        } else {
            width = rect.width;
            height = rect.height + dv;
        }
        this.rect( new Rect(rect.x, rect.y, width, height) );
    };

    proto._createDom = function() {
        Base._createDom.call(this);
        for (var i=0; i < this._containers.length; i++) {
            this._initContainer(this._containers[i]);
            this._dom.appendChild(this._containers[i]);
        };
    };

    proto._initContainer = function(c) {
        if (this._horizontal) {
            c.style.cssText += ';float:left';
            c.style.height = this.rect().height + 'px';
        }
    };

    proto._layoutDom = function(rect) {
        Base._layoutDom.call(this, rect);
        var i, l, c, v;
        for (i=0, l = this._containers.length; i < l; i++) {
            c = this._containers[i];
            v = this._childViews[i].rect()[this._dimension];
            if (v != this._containerSizes[i]) {
                c.style[this._dimension] = v + 'px';
                this._containerSizes[i] = v;
            }
        };
    };

    proto._createContainer = function(view) {
        return uki.createElement('div', 'position:relative;top:0;left:0;width:100%;padding:0;');
    };
});

uki.view.VerticalFlow = uki.view.Flow;
uki.view.HorizontalFlow = uki.newClass(uki.view.Flow, {
    _setup: function() {
        uki.view.Flow[PROTOTYPE]._setup.call(this);
        this._horizontal = true;
    }
});
uki.view.toolbar = {};

uki.view.Toolbar = uki.newClass(uki.view.Container, new function() {
    var Base = uki.view.Container[PROTOTYPE],
        proto = this;

    proto.typeName = function() { return 'uki.view.Toolbar'; };

    proto._moreWidth = 30;

    proto._setup = function() {
        Base._setup.call(this);
        this._buttons = [];
        this._widths = [];
    };

    this.buttons = uki.newProp('_buttons', function(b) {
        this._buttons = b;
        var buttons = uki.build(uki.map(this._buttons, this._createButton, this)).resizeToContents('width');
        this._flow.childViews(buttons);
        this._totalWidth = uki.reduce(0, this._flow.childViews(), function(s, v) { return s + v.rect().width; });
    });

    uki.moreWidth = uki.newProp('_moreWidth', function(v) {
        this._moreWidth = v;
        this._updateMoreVisible();
    });

    proto._createDom = function() {
        Base._createDom.call(this);

        var rect = this.rect(),
            flowRect = rect.clone().normalize(),
            moreRect = new Rect(rect.width - this._moreWidth, 0, this._moreWidth, rect.height),
            flowML = { view: 'HorizontalFlow', rect: flowRect, anchors: 'left top right', className: 'toolbar-flow', horizontal: true },
            moreML = { view: 'Button', rect: moreRect, anchors: 'right top', className: 'toolbar-button',  visible: false, backgroundPrefix: 'toolbar-more-', text: '>>', focusable: false },
            popupML = { view: 'Popup', rect: '0 0', anchors: 'right top', className: 'toolbar-popup', background: 'theme(toolbar-popup)', shadow: 'theme(toolbar-popup-shadow)',
                childViews: { view: 'VerticalFlow', rect: '0 0', anchors: 'right top left bottom' }
            };

        this._flow = uki.build(flowML)[0];
        this._more = uki.build(moreML)[0];
        this.appendChild(this._flow);
        this.appendChild(this._more);
        popupML.relativeTo = this._more;
        this._popup = uki.build(popupML)[0];

        var _this = this;
        this._more.bind('click', function() {
            _this._showMissingButtons();
        });
    };

    proto._showMissingButtons = function() {
        var maxWith = this._flow.rect().width,
            currentWidth = 0,
            missing = [], _this = this;
        for (var i=0, childViews = this._flow.childViews(), l = childViews.length; i < l; i++) {
            currentWidth += childViews[i].rect().width;
            if (currentWidth > maxWith) missing.push(i);
        };
        var newButtons = uki.map(missing, function(i) {
            var descr = { html: childViews[i].html(), backgroundPrefix: 'toolbar-popup-' };
            uki.each(['fontSize', 'fontWeight', 'color', 'textAlign', 'inset'], function(j, name) {
                descr[name] = uki.attr(childViews[i], name);
            })
            return _this._createButton(descr);
        });
        uki('Flow', this._popup).childViews(newButtons).resizeToContents('width height');
        this._popup.resizeToContents('width height').toggle();
    };

    proto._updateMoreVisible = function() {
        var rect = this._rect;
        if (this._more.visible() != rect.width < this._totalWidth) {
            this._more.visible(rect.width < this._totalWidth);
            var flowRect = this._flow.rect();
            flowRect.width += (rect.width < this._totalWidth ? -1 : 1)*this._moreWidth;
            this._flow.rect(flowRect);
        }
    };

    proto.rect = function(rect) {
        var result = Base.rect.call(this, rect);
        if (rect) this._updateMoreVisible();
        return result;
    };

    proto._createButton = function(descr) {
        return uki.extend({
                view: 'Button', rect: new Rect(100, this.rect().height), focusable: false, align: 'left',
                anchors: 'left top', backgroundPrefix: 'toolbar-', autosizeToContents: 'width', focusable: false
            }, descr);
    };
});}());(function() {
    var defaultCss = 'position:absolute;z-index:100;-moz-user-focus:none;font-family:Arial,Helvetica,sans-serif;';

    function u(url) {
        return uki.theme.airport.imagePath + url;
    }

    function shadowData() {
        var prefix = "shadow/large-";
        return {
            c: [u(prefix + "c.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAC4AAAAuCAYAAABXuSs3AAACzklEQVRo3t2a63KiUBCEPYCX1U2Ixvd/Qm/kYjRBWd2aTjW950CS3fyYtaprULl8p2kGAcMg/QqR6SDTsXk/8moi041Mx+bt3WAKVDVIDOQj0ArcROCbFHzoAGbYTICzLwygC/jc8T62bGccFDKLKLUXeH2625sIpCo2mBa8bkiBWbkpo5oaQMrxFPCJ6ikxkNYAQg90Tiqk5h0DiDmeAoZqqTqIFrxuSB0uSENTQVUHkHJdnVbgN6qYrmkQ6n7U6VygRwY6Eg1pHiyDdcQcx0YZGLCvInxWyx44q+Nwi6Hh8Ng0kTqieTQ2QcCbSDzeCPB40UHqUfYAlvu9Lu0aDD0i0B+iiQnup1wfdLgNdw+mFxEG8CrwZziuB6JCT00zqQyfcn3Q4TZD7y96lrqPwL9HJkiLKygecPcK+tN0Y3VG348lMlnC8bNE5EjuXmGfLnq0+mSf4fujuh6kM8DtCUHfmG6pMry63uc4u83QDwaO+kjwB3U9SD45InD61lSS4PzU4GNxUXCNyYvFAU5XpAcTnOfI/AFeiNuIxhX0TgT3pxKXoge8lpjsyeWdqKLosOs1wIcEzgck3L6Czk0Le1/ad7O/BH826MpgNxdtTTtynQ/UFngh4DNym6HvbfqO4oKcfwYc+UZMdga7FviKss7gdbB45NJNAA637wl8QXFBzsfSz7vAccLZ00EJt9dU4TofpOgup0AbLKSbICYAZiEu3NM/6zh6NmKyFm0oLtxdWo5z/8ZJpiTwpYDPxfGvgsPxrUCvCLyik9J7P1dw7igAB+zStDDw8h+BVwa+MeAVDQDg3FmS4NxR5gTN9TvA1wS9opxrZ+kFL6mbLEnfDb6iqGzJ8f8f3F1UXB6cLtuhyxOQy1O+2x9Zbn/Wur2QcHvp5vZi2e3tCbc3hNzegnN709P1bWaXN/bdPkpx/fDK9eNCtw9oXT8Sd/MnhF+iLpLibpmRrgAAAABJRU5ErkJggg==", u(prefix + "c.gif")],
            v: [u(prefix + "v.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAC4AAAAECAYAAADxjg1nAAAAWklEQVQYGdXBWwpAQAAAwLEeSUqy9z/hSkpSnh9OsTMFGlSo0aJDjwEjJkREREwYMaBHhxY1KpQIKPxePLhx4cSBHRtWLJiRkJAwY8GKDTsOnLiCTAWZCjL1AeihFg5/1kytAAAAAElFTkSuQmCC", u(prefix + "v.gif")],
            h: [u(prefix + "h.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAuCAYAAAAPxrguAAAAe0lEQVQoz5XSWwtAQBCG4XEMOST+/y8kOYScKRe8WzZbc7FPX7PNtLaIuPI49l0vUBIewT/LuO/7BRETMRMpExkh/w9KD+WVhBASAu20jnZjFsEkGAQh7ISNsBIWwkwYCT2hI9SEilASiv+g9KgEH6ZhomVi0E47fW7sAEmnGr/QVlzBAAAAAElFTkSuQmCC", u(prefix + "h.gif")],
            m: [u(prefix + "m.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAYAAACp8Z5+AAAAEUlEQVQIHWNgYGD4i4ZJFQAAAkoP0RsgosoAAAAASUVORK5CYII=", u(prefix + "m.gif"), true]
        };
    };

    uki.theme.airport = uki.extend({}, uki.theme.Base, {
        imagePath: 'http://static.ukijs.org/pkg/0.0.5/uki-theme/airport/i/',

        backgrounds: {
            'button-normal': function() {
                var prefix = "button/normal-";
                return new uki.background.Sliced9({
                    c: [u(prefix + "c.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAYAAAAGCAYAAADgzO9IAAAAS0lEQVQIW2NgAILy8vL/yJgBJrh+/fr/MABigyVBxN9//1EwXGLGrDn/j5++9P/G7Qf/t+/YBZEA6k5LTU39j4xBYmB7QAxkDBIDALKrX9FN99pwAAAAAElFTkSuQmCC", u(prefix + "c.gif")],
                    v: [u(prefix + "v.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAYAAAATCAYAAACz13xgAAAAT0lEQVQYlZXPMQ6AQAhE0b9m78zZFca1sdEwxZLQ8MIQiIh1XuvTEbEmQOnmXxNAVT2UB5komY1MA5KNys3jHlyUtv+wNzhGDwMDzfyFRh7wcj5EWWRJUgAAAABJRU5ErkJggg==", false, true],
                    h: [u(prefix + "h.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAH8AAAAGCAYAAADqkEEaAAAAMklEQVRIie3DUQ0AIAxDwZodFmaVhB+MjIeQ9pJTd5OeRdjSPEjP2ueSnlVVpGcBKz1/kUWrDOOOWIQAAAAASUVORK5CYII=", u(prefix + "h.gif")],
                    m: [u(prefix + "m.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAH8AAAATCAYAAAC5i9IyAAAAXklEQVQYGe3BgRHDMAACsXeP/fdtDUkHAUnf3/syleQ8TCfFZjrJNtNJdphOSsx00r1mOikJ00nJZTrJDtNJdphOci7TSXGYTkrMdJIdppP4HKaTDofpJA5TSnCYTn/FLC2twbqbSQAAAABJRU5ErkJggg==", null, true]
                }, '3 3 3 3');
            },

            'button-hover': function() {
                var prefix = "button/hover-";
                return new uki.background.Sliced9({
                    c: [u(prefix + "c.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAYAAAAGCAYAAADgzO9IAAAAS0lEQVQIW2NgAILy8vL/yJgBJrh+/fr/MABigyVBxN9//1EwXGL+wqX/b9579v/Ji3f/9+w9AJEA6m5ITU39j4xBYmB7QAxkDBIDAN/zYPRpDtd1AAAAAElFTkSuQmCC", u(prefix + "c.gif")],
                    v: [u(prefix + "v.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAYAAAATCAYAAACz13xgAAAAT0lEQVQY062PMQ7AIAwDj8Kf83gw7tKlhQxItZTp5EtCRLh3vyYi3AA0J980gJmBoayh31S290DS4Q4pUzlTjdOr0j9KLXvAanrAWuAiyQ2Hqz+Eaxa7lwAAAABJRU5ErkJggg==", false, true],
                    h: [u(prefix + "h.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAH8AAAAGCAYAAADqkEEaAAAAM0lEQVRIS+3DsREAIAwDMS+bGdIyLAUVG4RnEFt3UneTnkXY0jxIz9rnkp5VVaRnASs9f4uJy0upJnsYAAAAAElFTkSuQmCC", u(prefix + "h.gif")],
                    m: [u(prefix + "m.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAH8AAAATCAYAAAC5i9IyAAAAXElEQVQYGe3BgRHAMAABQLnaf+EEHYR/3ptgKlE2phNtYzrxyZhOtIzpRNuYTnwyphOTYDpREqYTbWM6UQqmExVhOtHPmE5MgunEJ2M68XwH04kIphRxMKWIqfUDGFEu5jKnhiUAAAAASUVORK5CYII=", null, true]
                }, "3 3 3 3");
            },

            'button-down': function() {
                var prefix = "button/down-";
                return new uki.background.Sliced9({
                    c: [u(prefix + "c.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAYAAAAGCAYAAADgzO9IAAAAR0lEQVQIW2NgAIKGhob/yJgBJlhQUPg/JTUDjEFssCSIyC8o+l9b1wjGIDZcoq9v4v9tO/aDMYiNYhyGHSDw////NGQMEgMAouBOxXrB3FIAAAAASUVORK5CYII=", u(prefix + "c.gif")],
                    v: [u(prefix + "v.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAYAAAATCAYAAACz13xgAAAAkUlEQVQYV42Nuw4CMQwEHT9ojvvnNPTQ8LfIeH3BmKuwNMomI294zulz3vz+eCbIeGOK2a4b7fueIGNSmF1IRBPkEqxMYpIgl1A2UllE/m5IbCyQS4hEjS4iN6FHXYDcBCokkV7FrYp7lcXFVA+6oME0xkiQS3weS9YGj19q48QfVbQ+zY+b4BMlXu7kcfrKmDdNVhnN3VjMVQAAAABJRU5ErkJggg==", false, true],
                    h: [u(prefix + "h.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAH8AAAAGCAYAAADqkEEaAAAARklEQVQYGe3BsQ2AQAwDQNvD0oCYIQ0UbIVExVDxDxLfsaqMGIn7cVoSYpbuBq/7sSTELN0Nvt9vgohZDINVZcRItL0hRloovBiO+VNuegAAAABJRU5ErkJggg==", u(prefix + "h.gif")],
                    m: [u(prefix + "m.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAH8AAAATCAYAAAC5i9IyAAAAa0lEQVQYGe3BsREDQQwDseWJF/n7n3lH7lQuhAT8fn8rUWF2wc/zQRKVZXexfalMnjtUJnsulckzQ2XyeKhM9lwqk2eGyuTxUJl8bSqTpUNlsiQqkzmiMllUKkuiMhkdKpMPlcoLSKKy7C5/du0Mt289U6QAAAAASUVORK5CYII=", null, true]
                }, "3 3 3 3");
            },

            'button-focus': function() {
                if (uki.image.needAlphaFix) return new uki.background.CssBox('filter:progid:DXImageTransform.Microsoft.Blur(pixelradius=3);background:#7594D2;', { inset: '-5 -5 -4 -5', zIndex: -2 } );

                var prefix = "button/focusRing-";
                return new uki.background.Sliced9({
                    c: [u(prefix + "c.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAwAAAAMCAYAAABWdVznAAAAtUlEQVQokWNgQABGBof9LNqhq9hkQo9xgjCIDRIDy6GA0FXMKp7b2NX9jvCqJB4S1Y47IgfCIDZYDCgHUgM3GSSgkLBfQCfxoKxO3Ak93fijdiAMYoPEQHJgTWCbgFaCTAFJ6MafMNZNPOGvl3AiC4RBbJAYSA6kBuw8kDtBVoNNBis+WQWzGsQGiYHkwE4F+QnsOaB7QU4AmcqABsA2AeVAakBqSddAspNI9jTpwUpGxJGUNADqMZr1BXNgDAAAAABJRU5ErkJggg=="],
                    v: [u(prefix + "v.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAwAAAAUCAYAAAC58NwRAAAAf0lEQVQoU2MQj93JrZlwQlUn/qS3TsLJegY0ABIDyYHUgNQyqPsd4dWJPa6pl3giRDfxeB+6BpAYSA6kBqSWQSl0N79m7FEdvcSTkUA8DV0DSAwkB1IDUgvWoBN3Qk83/ni0buKJGegaQGIgOZCaUQ2jGgZeA0nJm+QMRGoWBQCeEP1BW4HCpgAAAABJRU5ErkJggg=="],
                    h: [u(prefix + "h.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAIEAAAAMCAYAAABFjt5WAAAAlElEQVRYw+2YMQoCMRBFpxextBcWCUxykQUrLUTibk7iDbYUryK2SpJbubN7jLwHr0n7msmXfXxvjqfv9nD57LAtrbv1FzeWTmN2Lv5U7yVgGy69rfvcX3SofUjlHFK9+iHfsA2tt3W3/qJjffiUp/nx6VN5YRuuvfNk/QUAAADA4DDkMOSLyBexZyxiLOqE2ZjZ+A+dZWjNi3C4GwAAAABJRU5ErkJggg=="]
                }, "6 6 6 6", { inset: '-4 -4 -3 -4', zIndex: 2 });
            },

            'toolbar-button-normal': function() {
                return new uki.background.Css('#CCC');
            },

            'toolbar-button-hover': function() {
                return new uki.background.Css('#E0E0E0');
            },

            'toolbar-button-down': function() {
                return new uki.background.Css('#AAA');
            },

            'toolbar-button-focus': function() {
                return new uki.background.Css('#CCC');
            },

            'panel': function() {
                var prefix = "panel/dark-";
                return new uki.background.Sliced9({
                    v: [u(prefix + "v.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAYAAABtCAYAAACC/xDRAAAAZklEQVRIx+3UwQ3AIAgF0E/jjaW6rY7TTeyNG3QBOJi0sRr+9SV80USqtRmckIi4UHq/4YKIuHAgSAjFzKaNmlq+/ah/li8G8YKq+nnH+B7RqbZ/qISE/ABWu8SEF4CY+XIBwImRPBYi1Esk9DsnAAAAAElFTkSuQmCC", u(prefix + "v.gif"), true],
                    m: [u(prefix + "m.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAABtCAYAAACGCsDsAAAAYElEQVQ4y2NcvnzFfwYkwPj9+3cUAZb37z8woAgAVaAIMDGgAQwBlv///1NdC00MHbRa6GPoAAlgOuzfv38Um0HYHei2DNoAGhX4P5pwB9xzw1qAkZub+wKKABA7MOADAEid1EugsXRPAAAAAElFTkSuQmCC", u(prefix + "m.gif"), true]
                }, "0 3 0 3");
            },
            'input': function() {
                return new uki.background.CssBox(
                   'background:white;border: 1px solid #787878;border-top-color:#555;-moz-border-radius:2px;-webkit-border-radius:2px;border-radius:2px;-moz-box-shadow:0 1px 0 rgba(255, 255, 255, 0.4);-webkit-box-shadow:0 1px 0 rgba(255, 255, 255, 0.4);box-shadow:0 1px 0 rgba(255, 255, 255, 0.4)',
                   { inset: '0 0 1 0' }
               );
            },
            'slider-bar': function() {
                var prefix = "slider/bar-";
                return new uki.background.Sliced9({
                    v: [u(prefix + "v.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAYAAAASCAYAAAB4i6/FAAAASUlEQVQY02NgGHqgvLz8PzKGC7a0tP1ftnwNGIPYYEkQsW//0f/Hjp8FYxAbLjFjxiy4BIgNlvj//38auh0gMbA9IAYyHvDQAACE3VpNVzKSLwAAAABJRU5ErkJggg==", u(prefix + "v.gif")],
                    m: [u(prefix + "m.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAARMAAAASCAYAAAB4gjqpAAAAUUlEQVQYGe3BwRFAMBAAwItJIVTAR0taSE9eVHiKOA9jdjcCAAAAAAAAgJ9rY4wMgKK+bnsAVPV5XgKgagqAF/T7OgOgqmXmEQAAAAAAAHzTAx6DCNiUJps4AAAAAElFTkSuQmCC", u(prefix + "m.gif"), true]
                }, "0 3 0 3", {fixedSize: '0 18'});
            },
            list: function(rowHeight) {
                return new uki.background.Rows(rowHeight, '#EDF3FE');
            },
            'popup-normal': function() {
                return new uki.background.CssBox('opacity:0.95;background:#ECEDEE;-moz-border-radius:5px;-webkit-border-radius:5px;border-radius:5px;border:1px solid #CCC');
            },
            'shadow-big': function() {
                return new uki.background.Sliced9(shadowData(), "23 23 23 23", {zIndex: -2, inset: '-4 -10 -12 -10'});
            },
            'shadow-medium': function() {
                return new uki.background.Sliced9(shadowData(), "23 23 23 23", {zIndex: -2, inset: '-1 -6 -6 -6'});
            }
        },

        images: {
            'slider-handle': function() {
                return uki.image(u("slider/handle.gif"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAoAAAAkCAYAAACwlKv7AAABcUlEQVQ4T7WSy0sCYRRH758b0aKFCwmjB1kwlGEIJQiVvRe6iajIVpGbJArKjWK0GCgq7DE2muJt7kf3cmeaBlz0wVl8Z34Dw3AAvLNbqmIUtIHNo2sk/jr8HNb2K/jV60dCG8gVy9jq9IThsXmDdrQxw2arKySzRYN2mcIpQma7hE/vHYHuYQ5SG8doN9sC3cMcWKsHeP/sCnQPc5BcKWDtoSXwN/qct4GZzA5WbUcgSWhHG5hczP+SwZdpAwkrhxeNN2EivWXQjjYQn13Gcu1VSKTWDdrRBmJTS/6h9zahHW1gdHwBT25fIqENjMTnkND/TcPPTWpDsWmMAgY+AxXedZxQfIXffWIkUriWXLh2UvjlIwpcj3ZSuE/+FB50pvAzGwUuPOhM4aVGX+DCg84UfljvCvyNPseFF6oocOHaSeF7N22BC9dOCs9fuQIXrp0Unq24AheunRSeLjsCF66dFG6df0TiK7xid0L5v8K/AYNKQJdGv2S4AAAAAElFTkSuQmCC");
            },
            'slider-focus': function() {
                if (uki.image.needAlphaFix) {
                    var i = new uki.createElement('div', 'width:12px;height:18px;filter:progid:DXImageTransform.Microsoft.Blur(pixelradius=4);background:#7594D2;');
                    i.width = 20;
                    i.height = 28;
                    return i;
                }
                return uki.image(u("slider/focus.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABIAAAAZCAYAAAA8CX6UAAABUUlEQVQ4y+2VsUoDQRCGVxRFRLQRGxUJETmyd4292IoQm1wRwuVuWx/BxjewDHmGPIAgFkFJvL23cr45MVhli20EFxaW2f//b+afvV1jVmPD3My3evls+yT/3D0uXvcu+4v9Tv52wGRNjD0wYJXza+Szze7tyw7grvs46o0XZ0nlL2xRJ0mxtExdS4w9MIoVDtyfTAicV/ND695P7dhnabm8tmVzlzk/yFwzbKcfENM9wYCFo2KamaSIOhtp6a9S5++zyj/YqnlKXf0sIhMma2LsgQELB66WSb2kqpmoSPNo1gwwYOGoFXim5kndpMzXTODQzIQDFw1DJ9RYqZ/UQ4XAwoGLhlF/pCOYiQ+hQq1/0gDhqk+cEdr73Z1JcGltE4Zw0VChtuX1SAychmfkp3Dg/gv9aaEo5yjayY72r0X7+6PdR9FuyHh3dsRXJMq79gUgPopCCBOTpwAAAABJRU5ErkJggg==");
            },
            checkbox: function() {
                return uki.image(u("checkbox/normal.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABIAAABsCAYAAABn5uLmAAAF+ElEQVRYw+WY7U9TZxjG/b4v+xP2J5iZ7MM+mX3QfZuZyeIHEoxMnbEOnDMqMIcKE3HgGy+uQgRFVApqoVAKRZCVWqSFymtBQLSUN4FCKX2jeO3cR9vx9Bz0PE22hO0kv6R9nvu6OM85yc11d8uWOK/Ptm799PMvv8qQbGhtM+Dh4Ils7E5UYfvX3+6NmmjMr0C8fQtFlFVU4cjpS0jNLcXuvapdUaM7LcMIrq4p4nGbGft/zsTF0nrk3W4Ec6ybDc/hDYRFXrx0Rj/HYu8fRlLKrzhXVAV1VSuKdTbWSK3tgHtlFa1mG3Z+swfWHof4fT3OGTe+T05Het4tXCpvxPNhJ0jHGOVXtuGNJ4gr6nLsO5qBfaqTeDk5L65FSEnNxLHMQuSW6VH3pFtcIx1jRGeddAcwOulGbslDpJy5hv3JqRh2zonr2Vdu4NCpHFwoqUXhvSZxjZA8o+ziGrya84t0DjqRVaTB4bSLSPrxFMofGpB4JA1nCyuFu6kX9yO1pGOMzglFIzO+KE2WQWRcrUDy6cs4eDwL6bllyCmpwcNmG1NHOsboF+EBDk56GSrqzcgU7ixLeEPZNx6h6J5BUkM6xujEeTV6nMsSCu42CM9Fi6zrGnSOzEn2SccYHT1zFbZxjwTz0CzOF93Hgxa77D7pGKPDqTmwjC5yQzrG6MCxsyD6XV7FRDSMEYBte1Wp4IV0kjZCi7xs2RxXu60PvEhMmtutINbW3iomomGM9C1mhNfWuCEd2/gbHmM1HOaGdIxRpVaP0OoqN6RjjMo12riMSMcYFd++j2AoxA3pGKOiklsIBIPckI4xulx4A75AkBvSMUYX8vKx4g9wQzq2Z5//HV6fnxvSMUbpZ7KwvOLjhnSM0fHU0/B4V7ghHWOUfOwkCJ5XH9FImtohVQp4+Y92yH8nQyq9Ppghw0LIVEKryfLhDBkS+jAx7nRFP8fSPzSCpKPvM2T1E/kM6Q2uwdTRhZ279oihk76vZ3p+CftT/s6Q/aMT8hlyWUit+SV3xAyZdOQkpuYWxbUIP6VlRTOkwWR/Vy+XIZf8YUwtLOPSzXcZ8mBKGlxvFsX13IKSaIa8XmkU1wjZDDnvDYn0j7rEYEUZ8kBKKjS6pmiGzBMyJO1HamUz5KwQdyOYuodkM6S+rYupk82Qk4sBhgdNFiZDFmuaJDWyGfL1vF+CurJRzJC/CUcddrkl+7IZ8tWcT8LQxAIuqivR9LRXdl82Q4698XKzYYZ0uf2K2ShDJsSZIRPkOmQCL5skQ/J2SEJiwjtlExENY8QzZa+HdBtO2TyQTnbK5mXDKZuXDadsXj44ZfPw0SlbKYqmbCUonrI/huIp+2Nsgimbrng65CZpbLzXjh07PklMTPhOsjHhGgUPj7QaZJxN16hUP2yPmoyNDYAICjOYEoaGHKiprYLRWIeMjLQvokaOoeeKTaanpwQDAwYGemCxxLSRnt7OdcWBDU2Wlz1oaKiD3W7FyIgD3XYLa2S1touFs7NTKC8vw+LSgqyRXq/Ds2ftGBzshc/nBekYI7O5VSzs7raiXl8Lna5GYmI0NuJPUwv6+uyYmpoQ10jHGLW0GKKCnh4bWluNqKurFb77xbWODjOajHpxz+Hoi9aSjjEyNOqim0vCsaxWi3AHeuHOHmF4eFB4Qw+Etafo7e0S9yO1pGOManXV7x/yO6ZnJoS7MKG52QB9g044Qptw7E44nWNMHekYo6rqu8KGl4HeSqfwMK1Ws3iH9Gxia0jHGFVU3BLegkcCHcVqe4pnnSZ4PPOSfdIxRqWlxfB6FyS4F2fFI46Pj8juk479T6suEP7iHDekY4yu5V8GEQh4FRPRSAbj3Lwc8PJ/+x3S5XoNXiQmY2MvQISFX/KUEtEwRg7HAJdJBNKxHbKnOy4j0sV0yI64jEgX0yFNcRmRLqZDNsdlRDrGqLFRH5cR6RgjnU4blxHpGKPq6qq4jEgX0yHvxGVEupgOeTMuI9LFdMg/4jIiHWNUUJAPgsckomGMQqHQtry8XPBCOkkHoEVe/pF+9hfUKDhR04w/OgAAAABJRU5ErkJggg==", u("checkbox/normal.gif"));
            },
            'checkbox-focus': function() {
                if (uki.image.needAlphaFix) {
                    var i = new uki.createElement('div', 'width:18px;height:18px;filter:progid:DXImageTransform.Microsoft.Blur(pixelradius=3);background:#7594D2;');
                    i.width = 26;
                    i.height = 26;
                    return i;
                }
                return uki.image(u("checkbox/focus.png"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAABdklEQVRIie2Wv0rDUBSHUxRFRHQRFxWRioTeZHEXVxHqYgaRNrmrj+Ai+AAdi6trH0AQh6K05uYpfBQ934m0EnHQcIuDB+5y7/l9J+dPchMEU2sER8P5VjJY2ExeljY6D8v77dHKbvK4+t3iHD/80aFXzhdLBnPN4/tFBE37vN7qjrbDzO2ZTh6GnbExXRdXl+5zLn74o1O9cOB9xjfY3MmGa8Y+bSGO0vGhSYuT2Lqz2BbnUZpfVBf7nOOn/gQWPRwNMslE0iIyh1HqDiLrTuPMXZqsuI5s3hNQX/Zuq6vcz3v44Y8OPRx4H+UKAmpHevrkCi+ubu5e33660KGHo2WmJ5g2SGpImjzJb+DTIJKJcODB1QBMgTZUakm6dQKghwMPrgbQ+ss00DBqWidA2TNpvPC0DxjzzMiVU1H0a5VI9HDgwZ0EKEdTxk+mo14GMmHCgfcf4I8F8DpF3t8D72+y92+R96+p9/vA/402gzt5Bn8VHu0d2HhIetPffvAAAAAASUVORK5CYII=");
            }
        },

        imageSrcs: {
            x: function() {
                return [u("x.gif"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVQIHWNgAAIAAAUAAY27m/MAAAAASUVORK5CYII="];
            },
            'splitPane-horizontal': function() {
                return [u("splitPane/horizontal.gif"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAMAAAAICAYAAAA870V8AAAAFUlEQVQIW2MoLy//zwAEYJq6HGQAAJuVIXm0sEPnAAAAAElFTkSuQmCC"];
            },
            'splitPane-vertical': function() {
                return [u("splitPane/vertical.gif"), "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAcAAAADCAYAAABfwxXFAAAAE0lEQVQIHWMsLy//z0AOYMSnEwAIngTLoazFLgAAAABJRU5ErkJggg=="];
            }
        },

        doms: {
            'splitPane-vertical': function(params) {
                var commonVerticalStyle = 'cursor:row-resize;cursor:ns-resize;z-index:200;overflow:hidden;';

                return params.handleWidth == 1 ?
                    uki.createElement('div',
                        defaultCss + 'width:100%;height:5px;margin-top:-2px;' +
                        commonVerticalStyle + 'background: url(' + uki.theme.imageSrc('x') + ')',
                        '<div style="' +
                        defaultCss + 'background:#999;width:100%;height:1px;left:0px;top:2px;overflow:hidden;' +
                        '"></div>') :
                    uki.createElement('div',
                        defaultCss + 'width:100%;height:' + (params.handleWidth - 2) + 'px;' +
                        'border: 1px solid #CCC;border-width: 1px 0;' + commonVerticalStyle +
                        'background: url(' + uki.theme.imageSrc('splitPane-vertical') + ') 50% 50% no-repeat;');

            },

            'splitPane-horizontal': function(params) {
                var commonHorizontalStyle = 'cursor:col-resize;cursor:ew-resize;z-index:200;overflow:hidden;';

                return params.handleWidth == 1 ?
                    uki.createElement('div',
                        defaultCss + 'height:100%;width:5px;margin-left:-2px;' +
                        commonHorizontalStyle + 'background: url(' + uki.theme.imageSrc('x') + ')',
                        '<div style="' +
                        defaultCss + 'background:#999;height:100%;width:1px;top:0px;left:2px;overflow:hidden;' +
                        '"></div>') :
                    uki.createElement('div',
                        defaultCss + 'height:100%;width:' + (params.handleWidth - 2) + 'px;' +
                        'border: 1px solid #CCC;border-width: 0 1px;' + commonHorizontalStyle +
                        'background: url(' + uki.theme.imageSrc('splitPane-horizontal') + ') 50% 50% no-repeat;');

            }
        }
    });

    uki.theme.airport.backgrounds['button-disabled'] = uki.theme.airport.backgrounds['button-normal'];
    uki.theme.airport.backgrounds['input-focus'] = uki.theme.airport.backgrounds['button-focus'];
    uki.theme.airport.backgrounds['popup-shadow'] = uki.theme.airport.backgrounds['shadow-medium'];
    uki.theme.airport.backgrounds['toolbar-popup'] = uki.theme.airport.backgrounds['popup-normal'];
    uki.theme.airport.backgrounds['toolbar-popup-shadow'] = uki.theme.airport.backgrounds['shadow-medium'];

    uki.theme.register(uki.theme.airport);
})();
