# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

# Argonought
# ==========
#
# Argonought is meant as a complementary serialisation format to JSON.

define 'argo', (exports, root) ->

  exports.encode = (val) ->
    stream

  exports.decode = (stream) ->
    val

  exports.TYPES = TYPES =
    Nil:            0
    Any:            1
    BigDecimal:     2
    Bool:           3
    Byte:           4
    ByteSlice:      5
    Complex64:      6
    Complex128:     7
    Dict:           8
    Float32:        9
    Float64:        10
    Int32:          11
    Int64:          12
    Map:            13
    Slice:          14
    String:         15
    StringSlice:    16
    Struct:         17
    StructInfo:     18
    Uint32:         19
    Uint64:         20
    sentinel:       21

  exports.ArgoType = class ArgoType
    constructor: (@value) ->

  exports.Nil = class Nil extends ArgoType
    type: TYPES.Nil

  exports.Any = class Any extends ArgoType
    type: TYPES.Any

  exports.BigDecimal = class BigDecimal extends Any
    type: TYPES.BigDecimal

  exports.Bool = class Bool extends Any
    type: TYPES.Bool

  exports.Byte = class Byte extends Any
    type: TYPES.Bool

  exports.Complex64 = class Complex64 extends Any
    type: TYPES.Complex64

  exports.Complex128 = class Complex128 extends Any
    type: TYPES.Complex128

  exports.Float32 = class Float32 extends Any
    type: TYPES.Float32

  exports.Float64 = class Float64 extends Any
    type: TYPES.Float64

  exports.Int32 = class Int32 extends Any
    type: TYPES.Int32

  exports.Int64 = class Int64 extends Any
    type: TYPES.Int64

  exports.String = class String extends Any
    type: TYPES.String

  exports.Uint32 = class Uint32 extends Any
    type: TYPES.Uint32

  exports.Uint64 = class Uint64 extends Any
    type: TYPES.Uint64

  exports.Dict = class Dict extends Any
    type: TYPES.Dict

  exports.Slice = class Slice extends Any
    type: TYPES.Slice

  exports.ByteSlice = class ByteSlice extends Any
    type: TYPES.ByteSlice

  exports.StringSlice = class StringSlice extends Any
    type: TYPES.StringSlice

  exports.Struct = class Struct extends Any
    type: TYPES.Struct

  exports.Map = class Map extends Any
    type: TYPES.Map
