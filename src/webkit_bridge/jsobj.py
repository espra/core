# No Copyright (-) 2009-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

from pypy.interpreter.baseobjspace import Wrappable, W_Root, ObjSpace
from pypy.interpreter.gateway import interp2app
from pypy.interpreter.typedef import TypeDef
from pypy.interpreter.error import OperationError
from webkit_bridge.webkit_rffi import *
from pypy.rpython.lltypesystem import lltype, rffi
from pypy.interpreter.argument import Arguments
from pypy.rpython.memory.support import AddressDict


class JavaScriptContext(object):

    def __init__(self, space):
        self._ctx = lltype.nullptr(JSValueRef.TO)
        self.space = space
        self.w_js_exception = space.appexec([], '''():
        class JSException(Exception):
            pass
        return JSException
        ''')
        # note. It's safe to cast pointers in this dict to ints
        # as they're non-movable (raw) ones
        self.applevel_callbacks = {}
        def callback(ctx, js_function, js_this, js_args):
            arguments = Arguments(space,
                [self.js_to_python(arg) for arg in js_args])
            w_callable = self.applevel_callbacks.get(
                rffi.cast(lltype.Signed, js_function), None)
            if w_callable is None:
                raise Exception("Got wrong callback, should not happen")
            w_res = space.call_args(w_callable, arguments)
            return self.python_to_js(w_res)
        self.js_callback_factory = create_js_callback(callback)

    def python_to_js(self, w_obj):
        space = self.space
        if space.is_w(w_obj, space.w_None):
            return JSValueMakeUndefined(self._ctx)
        elif space.is_true(space.isinstance(w_obj, space.w_bool)):
            return JSValueMakeBoolean(self._ctx, space.is_true(w_obj))
        elif space.is_true(space.isinstance(w_obj, space.w_int)):
            return JSValueMakeNumber(self._ctx, space.int_w(w_obj))
        elif space.is_true(space.isinstance(w_obj, space.w_float)):
            return JSValueMakeNumber(self._ctx, space.float_w(w_obj))
        elif space.is_true(space.isinstance(w_obj, space.w_str)):
            return JSValueMakeString(self._ctx, self.newstr(space.str_w(w_obj)))
        elif isinstance(w_obj, JSObject):
            return w_obj.js_val
        elif space.is_true(space.callable(w_obj)):
            name = space.str_w(space.getattr(w_obj, space.wrap('__name__')))
            js_func = self.js_callback_factory(self._ctx, name)
            self.applevel_callbacks[rffi.cast(lltype.Signed, js_func)] = w_obj
            return js_func
        else:
            raise NotImplementedError()

    def js_to_python(self, js_obj, this=NULL):
        space = self.space
        tp = JSValueGetType(self._ctx, js_obj)
        if tp == kJSTypeUndefined:
            return space.w_None
        elif tp == kJSTypeNull:
            return space.w_None
        elif tp == kJSTypeBoolean:
            return space.wrap(JSValueToBoolean(self._ctx, js_obj))
        elif tp == kJSTypeNumber:
            return space.wrap(JSValueToNumber(self._ctx, js_obj))
        elif tp == kJSTypeString:
            return space.wrap(self.str_js(js_obj))
        elif tp == kJSTypeObject:
            return space.wrap(JSObject(self, js_obj, this))
        else:
            raise NotImplementedError(tp)

    def newstr(self, s):
        return JSStringCreateWithUTF8CString(s)

    def str_js(self, js_s):
        return JSStringGetUTF8CString(JSValueToString(self._ctx, js_s))

    def get(self, js_obj, name):
        return JSObjectGetProperty(self._ctx, js_obj, self.newstr(name))

    def set(self, js_obj, name, js_val):
        js_name = JSStringCreateWithUTF8CString(name)
        JSObjectSetProperty(self._ctx, js_obj, js_name, js_val, 0)

    def eval(self, s, this=NULL):
        return JSEvaluateScript(self._ctx, s, this)

    def call(self, js_val, args, this=NULL):
        try:
            return JSObjectCallAsFunction(self._ctx, js_val, this, args)
        except JSException, e:
            raise OperationError(self.w_js_exception, self.space.wrap(e.repr()))

    def propertylist(self, js_val):
        return JSPropertyList(self._ctx, js_val)

    def globals(self):
        return JSObject(self, JSContextGetGlobalObject(self._ctx))


class JSObject(Wrappable):

    def __init__(self, ctx, js_val, this=NULL):
        self.ctx = ctx
        self.js_val = js_val
        self.this = this

    def descr_get(self, space, w_name):
        name = space.str_w(w_name)
        if name == '__dict__':
            proplist = self.ctx.propertylist(self.js_val)
            w_d = space.newdict()
            for name in proplist:
                w_item = self.ctx.js_to_python(self.ctx.get(self.js_val, name))
                space.setitem(w_d, space.wrap(name), w_item)
            return w_d
        js_val = self.ctx.get(self.js_val, name)
        return self.ctx.js_to_python(js_val, self.js_val)

    def descr_set(self, space, w_name, w_value):
        name = space.str_w(w_name)
        if name == '__dict__':
            raise OperationError(space.w_ValueError,
                                 space.wrap("Cannot change __dict__"))
        js_val = self.ctx.python_to_js(w_value)
        self.ctx.set(self.js_val, name, js_val)
        return space.w_None

    def call(self, space, args_w):
        js_res = self.ctx.call(self.js_val, [self.ctx.python_to_js(arg)
                                             for arg in args_w], self.this)
        return self.ctx.js_to_python(js_res)

    def str(self, space):
        return space.wrap('JSObject(' + self.ctx.str_js(self.js_val) + ')')

JSObject.typedef = TypeDef("JSObject",
        __getattribute__ = interp2app(JSObject.descr_get,
                                      unwrap_spec=['self', ObjSpace, W_Root]),
        __getitem__ = interp2app(JSObject.descr_get,
                                 unwrap_spec=['self', ObjSpace, W_Root]),
        __setitem__ = interp2app(JSObject.descr_set,
                              unwrap_spec=['self', ObjSpace, W_Root, W_Root]),
        __setattr__ = interp2app(JSObject.descr_set,
                              unwrap_spec=['self', ObjSpace, W_Root, W_Root]),
        __str__ = interp2app(JSObject.str,
                             unwrap_spec=['self', ObjSpace]),
        __call__ = interp2app(JSObject.call,
                              unwrap_spec=['self', ObjSpace, 'args_w'])
)
