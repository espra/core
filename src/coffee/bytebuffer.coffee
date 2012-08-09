# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

# ByteBuffer
# ==========
#
# ByteBuffer is a layer around typed arrays with automatic extension built in.
 # Based on Go's bytes.Buffer

define 'argo.util', (exports, root) ->

  # Read operations
  READOP =
    opInvalid: 0
    opReadRune: 1
    opRead: 2

  exports.ByteBuffer = class ByteBuffer
    constructor: (@buf, @off=0, @lastRead=0, @bootstrap=new ArrayBuffer(64)) ->

    # Get bytes from buffer
    bytes: -> @buf.slice(@off)

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
      # Go has slices built in, here we check the buffer backing the Uint8Array 
      if @buf.byteLength+n > @buf.buffer.byteLength
        if !@buf and n <=  @bootstrap.byteLength # Avoid reallocation for small buffers
          @buf.buffer = @bootstrap.buffer
        else
          old = new Uint8Array(@buf)
          buf = new ArrayBuffer(2*@buf.byteLength+n)
          temp = new Uint8Array(buf)
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
