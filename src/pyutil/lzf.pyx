# Public Domain (-) 2009-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

"""Wrapper extension for the super fast LZF library."""

cimport python_exc

cdef extern from "stdlib.h":
    ctypedef unsigned long size_t
    void free(void *ptr)
    void *malloc(size_t size)
    void *realloc(void *ptr, size_t size)
    size_t strlen(char *s)
    char *strcpy(char *dest, char *src)

cdef extern from "lzf.h":
    unsigned int lzf_compress (char *in_data, unsigned int in_len, void *out_data, unsigned int out_len)
    unsigned int lzf_decompress (char *in_data,  unsigned int in_len, void *out_data, unsigned int out_len)
    
def compress(bytes data):
    """Compress the given data."""

    cdef int in_len = len(data)
    cdef int out_len = in_len - 4

    cdef char *out_data = <char *>malloc(in_len)
    if out_data is NULL:
        python_exc.PyErr_NoMemory()

    try:
        retval = lzf_compress(data, in_len, out_data+4, out_len)
        if retval == 0:
            return
        out_data[0] = (in_len >> 24) & 0xff
        out_data[1] = (in_len >> 16) & 0xff
        out_data[2] = (in_len >> 8) & 0xff
        out_data[3] = (in_len >> 0) & 0xff
        return out_data[:retval + 4]
    finally:
        free(out_data)

def decompress(bytes data):
    """Decompress the given data."""

    cdef unsigned int in_len = len(data) - 4
    cdef unsigned int out_len = (ord(data[0]) << 24) | (ord(data[1]) << 16) | (ord(data[2]) << 8) | ord(data[3])
    cdef char *in_data = <char *>data

    cdef char *out_data = <char *>malloc(out_len)
    if out_data is NULL:
        python_exc.PyErr_NoMemory()

    try:
        retval = lzf_decompress(in_data+4, in_len, out_data, out_len)
        if retval == 0:
            return
        return out_data[:out_len]
    finally:
        free(out_data)
