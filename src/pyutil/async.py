# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

def wrap_method(method):
    exec("""def wrapper(*args, **kwargs):
        if 'run' in kwargs:
            return method(*args, **kwargs)
        def %s(callback, errback=None):
            kwargs['callback'] = callback
            if errback:
                kwargs['errback'] = errback
            return method(*args, **kwargs)
        return %s""" % (method.__name__, method.__name__), locals())
    wrapper.__name__ = method.__name__
    wrapper.__doc__ = method.__doc__
    wrapper.__raw__ = method
    return wrapper

class Dispatcher(object):
    """An async process dispatcher."""

    def __init__(self, gen):
        self.gen = gen
        self.callback(None)

    def callback(self, arg):
        try:
            self.gen.send(arg)(callback=self.callback, errback=self.errback)
        except StopIteration:
            pass

    def errback(self, arg):
        try:
            self.gen.throw(arg)(callback=self.callback, errback=self.errback)
        except StopIteration:
            pass

def async(func):
    def wrapper(*args, **kwargs):
        Dispatcher(func(*args, **kwargs))
    wrapper.__name__ = func.__name__
    return wrapper
