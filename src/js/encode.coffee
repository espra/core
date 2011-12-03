# Public Domain (-) 2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

# This package implements an Argonought encoder to run in a JavaScript
# environment.

define 'argo', (exports, root) ->

  exports.Encoder = class Encoder
    constructor: (@value) ->

    @b: new ByteBuffer()
    @scratch: new ByteBuffer(new ArrayBuffer(11))  
    maxInt32: -1 >>> 1

    baseId: 64

    writeSize: (value) ->
      i = 0
      while value >= 128
        @scratch[i] = value | 128
        value >>= 7
        i += 1
      @scratch[i] = value
      @b.write(@scratch[0..i+1])
