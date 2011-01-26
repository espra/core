# Public Domain (-) 2005-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

"""Dictionary Subclasses and Functions."""

# ------------------------------------------------------------------------------
# CachingDict
# ------------------------------------------------------------------------------

Blank = object()

class CachingDict(dict):
    """A caching dict that discards its least recently used items."""

    __slots__ = (
        '_buffer_size', '_cache_size', '_clock', '_garbage_collector'
        )

    def __init__(
        self, cache_size=10000, buffer_size=None, garbage_collector=None, *args,
        **kwargs
        ):

        if buffer_size is None:
            self._buffer_size = cache_size / 2
        else:
            self._buffer_size = buffer_size

        self._cache_size = cache_size
        self._clock = 0
        self._garbage_collector = garbage_collector

        for key, value in args:
            self.__setitem__(key, value)

        for key, value in kwargs.iteritems():
            self.__setitem__(key, value)

    def __setitem__(self, key, value):
        excess = len(self) - self._cache_size - self._buffer_size + 1
        if excess > 0:
            excess = sorted(self.itersort())[:excess + self._buffer_size]
            garbage_collector = self._garbage_collector
            if garbage_collector:
                for ex_value, ex_key in excess:
                    garbage_collector(ex_key, ex_value)
                    del self[ex_key]
            else:
                for _, ex_key in excess:
                    del self[ex_key]
        self._clock += 1
        dict.__setitem__(self, key, [self._clock, value])

    def __getitem__(self, key):
        if key in self:
            access = dict.__getitem__(self, key)
            self._clock += 1
            access[0] = self._clock
            return access[1]
        raise KeyError(key)

    def itersort(self):
        getitem = dict.__getitem__
        for key in self:
            yield getitem(self, key), key

    def get(self, key, default=None):
        if key in self:
            return self.__getitem__(key)
        return default

    def pop(self, key, default=Blank):
        if key in self:
            value = dict.__getitem__(self, key)[1]
            del self[key]
            return value
        if default is not Blank:
            return default
        raise KeyError(key)

    def setdefault(self, key, default):
        if key in self:
            return self.__getitem__(key)
        self.__setitem__(key, default)
        return default

    iterkeys = dict.iterkeys
    keys = dict.keys

    def itervalues(self):
        getitem = self.__getitem__
        for key in self:
            yield getitem(key)

    def values(self):
        getitem = self.__getitem__
        return [getitem(key) for key in self]

    def iteritems(self):
        getitem = self.__getitem__
        for key in self:
            yield key, getitem(key)

    def items(self):
        getitem = self.__getitem__
        return [(key, getitem(key)) for key in self]

    def set_cache_size(self, cache_size, buffer_size=None):
        if buffer_size is None:
            self._buffer_size = cache_size / 2
        else:
            self._buffer_size = buffer_size
        self._cache_size = cache_size

    def get_cache_byte_size(self):
        getitem = self.__getitem__
        return sum(len(str(getitem(key))) for key in self)
