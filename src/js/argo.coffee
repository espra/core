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

  exports.types =
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
    Nil:            14
    Slice:          15
    String:         16
    StringSlice:    17
    Struct:         18
    StructInfo:     19
    Uint32:         20
    Uint64:         21
    sentinel:       22

define()

class ArgoType
  constructor: (@value) ->

class Any extends ArgoType

class BigDecimal extends ArgoType

class Bool extends ArgoType

class Byte extends ArgoType

class Complex64 extends ArgoType

class Complex128 extends ArgoType

class Float32 extends ArgoType

class Float64 extends ArgoType

class Int32 extends ArgoType

class Int64 extends ArgoType

class Nil extends ArgoType

class String extends ArgoType

class Uint32 extends ArgoType

class Uint64 extends ArgoType

class Dict extends ArgoType

class Slice extends ArgoType

class ByteSlice extends ArgoType
  
class StringSlice extends ArgoType

class Struct extends ArgoType

class Map extends ArgoType


