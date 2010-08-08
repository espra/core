# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

from tornado.ioloop import IOLoop

Loop = IOLoop.instance()

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

class TornadoWebDispatcher(object):
    """An async process dispatcher for tornado web handler methods."""

    def __init__(self, gen, handler):
        self.gen = gen
        self.handler = handler
        self.callback(None)

    def callback(self, arg, errback=None):
        try:
            if errback:
                self.cb = self.gen.throw(arg)
            else:
                self.cb = self.gen.send(arg)
            Loop.add_callback(self.next)
        except StopIteration:
            self.cb = None
            if not self.handler._finished:
                self.handler.finish()
        except Exception, error:
            self.cb = None
            if self.handler._headers_written:
                logging.error('Exception after headers written', exc_info=True)
            else:
                self.handler._handle_request_exception(error)

    def errback(self, arg):
        self.callback(arg, errback=1)

    def next(self):
        self.cb(callback=self.callback, errback=self.errback)

def web_async(method):
    def wrapper(handler, *args, **kwargs):
        handler._auto_finish = 0
        TornadoWebDispatcher(method(handler, *args, **kwargs), handler)
    wrapper.__name__ = method.__name__
    return wrapper
