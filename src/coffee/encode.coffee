# Public Domain (-) 2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

# This package implements an Argonought encoder to run in a JavaScript
# environment.

#TYPES = require("./argo").TYPES

#real = require("./argo/util").real
#imag = require("./argo/util").imag
#int64 = require("./argo/util").int64
#uint64 = require("./argo/util").uint64

define 'argo', (exports, root) ->

  # Type encoding constants
  ENC =
    encAny: new Uint8Array([TYPES.Any])
    encBigDecimal:  new Uint8Array([TYPES.BigDecimal])
    encBool: new Uint8Array([TYPES.Bool])
    encByte:  new Uint8Array([TYPES.Byte])
    encByteSlice: new Uint8Array([TYPES.ByteSlice])
    encComplex64: new Uint8Array([TYPES.Complex64])
    encComplex128: new Uint8Array([TYPES.Complex128])
    encDict: new Uint8Array([TYPES.Dict])
    encFloat32: new Uint8Array([TYPES.Float32])
    encFloat64: new Uint8Array([TYPES.Float64])
    encInt32: new Uint8Array([TYPES.Int32])
    encInt64: new Uint8Array([TYPES.Int64])
    encMap: new Uint8Array([TYPES.Map])
    encNil: new Uint8Array([TYPES.Nil])
    encSlice: new Uint8Array([TYPES.Slice])
    encString: new Uint8Array([TYPES.String])
    encStringSlice: new Uint8Array([TYPES.StringSlice])
    encStruct: new Uint8Array([TYPES.Struct])
    encStructInfo: new Uint8Array([TYPES.StructInfo])
    encUint32: new Uint8Array([TYPES.Uint32])
    encUint64: new Uint8Array([TYPES.Uint64])
    # End native types
    encDictAny: new Uint8Array([TYPES.Dict, TYPES.Any])
    encSliceAny: new Uint8Array([TYPES.Slice, TYPES.Any])
    encTrue: new Uint8Array([1])
    encFalse: new Uint8Array([0])
    encBoolTrue: new Uint8Array([TYPES.Bool, 1])
    encBoolFalse: new Uint8Array([TYPES.Bool, 0])

  exports.Encoder = class Encoder
    constructor: (@value) ->

    b: new ByteBuffer()
    scratch: new ByteBuffer(new ArrayBuffer(11))  
    maxInt32: -1 >>> 1

    baseId: 64

    writeSize: (value) ->
      i = 0
      while value >= 128
        @scratch[i] = value | 128
        value >>= 7
        i += 1
      @scratch[i] = value
      @b.write(@scratch.slice(0, i+1))

    encode: (value) ->
      switch value.type
        when TYPES.Bool
          if value
            @b.write(ENC.encBoolTrue)
          else
            @b.write(ENC.encBoolFalse)
        when TYPES.Byte then @b.write(ENC.encByte)
        when TYPES.ByteSlice
          @b.write(ENC.encByteSlice)
          @b.writeByteSlice(value)
        when TYPES.Complex64
          @b.write(ENC.encComplex64)
          @b.writeFloat32(real(value))
          @b.writeFloat32(imag(value))
        when TYPES.Complex128
          @b.write(ENC.encComplex128)
          @b.writeFloat64(real(value))
          @b.writeFloat64(imag(value))
        when TYPES.Float32
          @b.write(ENC.encFloat32)
          @b.writeFloat32(value)
        when TYPES.Float64
          @b.write(ENC.encFloat64)
          @b.writeFloat64(value)
        when TYPES.Int32, TYPES.Int64
          if is64bit
            @b.write(ENC.encInt64)
          else
            @b.write(ENC.encInt32)
          @b.write(int64(value))
        when TYPES.Int32, TYPES.Int64
          if is64bit
            @b.write(ENC.encUint64)
          else
            @b.write(ENC.encUint32)
          @b.write(uint64(value))
