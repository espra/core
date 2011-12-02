# Public Domain (-) 2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

# This package implements an Argonought encoder to run in a JavaScript
# environment.

define 'argo', (exports, root) ->

  exports.encode = (val) ->
    stream

  exports.decode = (stream) ->
    val

# Read operations
  READOP =
    opInvalid: 0
    opReadRune: 1
    opRead: 2

  exports.ByteBuffer = class ByteBuffer
    # Based on Go's bytes.Buffer
    constructor: (@buf, @off, @lastRead, @bootstrap = new ArrayBuffer(64)) ->

    # Get bytes from buffer
    bytes: -> @buf[@off...]

    # Get length from offset
    len: -> @buf.byteLength - @off

    # Truncate buffer
    truncate: (n) ->
      @lastRead = READOP.opInvalid
      if n is 0
        @off = 0
      @buf = @buff.slice(0, @off+n)

    reset: -> @truncate(0)

    _grow: (n) ->
      m = @len()
      if m is 0 and @off isnt 0
        @truncate(0)
      # Go has slices built in, here we check the buffer backing the Int8Array 
      if @buf.byteLength+n > @buf.buffer.byteLength
        if !@buf and n <=  @bootstrap.byteLength # Avoid reallocation for small buffers
          @buf.buffer = @bootstrap.buffer
        else
          old = new Int8Array(@buf)
          buf = new ArrayBuffer(2*@buf.byteLength+n)
          temp = new Int8Array(buf)
          temp.set(old.subarray())
          @buf = temp
          @off = 0
      @buf = @buf.slice(0, @off+m+n)
      @off + m

    write: (p) ->
      @lastRead = READOP.opInvalid
      m = @_grow(p.byteLength)
      @buf.set(p, m)
      p.byteLength


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
