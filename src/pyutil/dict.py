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
        '_cache_size', '_garbage_collector', '_buffer_size', 'itersort',
        '_clock'
        )

    def __init__(
        self, cache_size=10000, buffer_size=None, garbage_collector=None, *args,
        **kwargs
        ):

        self._cache_size = cache_size
        self._garbage_collector = garbage_collector
        self._buffer_size = buffer_size or cache_size / 2
        self._clock = 0

        for key, value in args:
            self.__setitem__(key, value)

        for key, value in kwargs.iteritems():
            self.__setitem__(key, value)

    def __setitem__(self, key, value):
        excess = len(self) - self._cache_size - self._buffer_size + 1
        if excess > 0:
            garbage_collector = self._garbage_collector
            # @@ time against : heapq.nsmallest()
            excess = sorted(self.itersort())[:excess + self._buffer_size]
            for ex_value, ex_key in excess:
                if garbage_collector:
                    garbage_collector(ex_key, ex_value)
                del self[ex_key]

        self._clock += 1
        return dict.__setitem__(self, key, [self._clock, value])

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

    def itervalues(self):
        getitem = self.__getitem__
        for key in self:
            yield getitem(key)

    def values(self):
        return list(self.itervalues())

    def iteritems(self):
        getitem = self.__getitem__
        for key in self:
            yield key, getitem(key)

    def items(self):
        return list(self.iteritems())

    def set_cache_size(self, cache_size):
        if not isinstance(cache_size, (int, long)):
            raise ValueError("Cache size must be an integer.")
        self._cache_size = cache_size

    def get_cache_byte_size(self):
        getitem = self.__getitem__
        return sum(len(str(getitem(key))) for key in self)
