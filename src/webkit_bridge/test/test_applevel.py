# Public Domain (-) 2009-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

from pypy.conftest import gettestobjspace
from webkit_bridge.jsobj import JSObject, JavaScriptContext
from webkit_bridge.webkit_rffi import (JSGlobalContextCreate,
                                       JSGlobalContextRelease,
                                       empty_object)
from pypy.interpreter.gateway import interp2app


class AppTestBindings(object):

    def setup_class(cls):
        ctx = JavaScriptContext(cls.space)
        ctx._ctx = JSGlobalContextCreate()
        cls.w_js_obj = cls.space.wrap(JSObject(ctx, ctx.eval('[]')))
        this = ctx.eval('[]')
        ctx.eval('this.x = function(a, b) { return(a + b); }', this)
        cls.w_func = cls.space.wrap(JSObject(ctx, ctx.get(this, 'x')))
        cls.w_js_obj_2 = cls.space.wrap(JSObject(ctx, ctx.eval('[]')))

        def interpret(source):
            return ctx.js_to_python(ctx.eval(source))

        space = cls.space
        cls.w_interpret = space.wrap(interp2app(interpret, unwrap_spec=[str]))
        cls.w_globals = ctx.globals()
        cls.w_JSException = ctx.w_js_exception

    def test_getattr_none(self):
        assert self.js_obj.x == None

    def test_getsetattr_obj(self):
        self.js_obj['x'] = 3
        assert isinstance(self.js_obj.x, float)
        assert self.js_obj.x == 3
        self.js_obj.y = 3
        assert isinstance(self.js_obj['y'], float)
        assert self.js_obj['y'] == 3

    def test_str(self):
        assert str(self.js_obj) == 'JSObject()'

    def test_call(self):
        assert self.func(3, 4) == 7
        assert self.func('a', 'bc') == 'abc'

    def test_obj_wrap_unwrap(self):
        self.js_obj['x'] = self.js_obj_2
        assert str(self.js_obj.x) == 'JSObject()'

    def test_floats(self):
        self.js_obj.y = 3.5
        assert self.js_obj.y == 3.5

    def test_bools(self):
        self.js_obj.x = True
        assert self.js_obj.x

    def test_none(self):
        self.js_obj.x = None

    def test_property_list(self):
        x = self.interpret('''
        function c () {
            this.x = 1
            this.y = 2
            this.z = 3
        }
        new c()
        ''')
        assert x.__dict__ == {'x':1, 'y':2, 'z':3}

    def test_global(self):
        self.interpret('''
        xxx = 3
        ''')
        assert self.globals.xxx == 3

    def test_method(self):
        x = self.interpret('''
        function c () {
            this.zzz = 3;
            this.f = function (x) {
                return (this.zzz + x);
            };
        }
        new c()
        ''')
        assert x.f(3) == 6

    def test_raising_call(self):
        f = self.interpret('''
        function f(x) {
            throw TypeError;
        }
        f
        ''')
        raises(self.JSException, f, 3)

    # XXX more raising tests

    def test_wrapped_callback(self):
        f = self.interpret('''
        function f(x) {
            return x(3);
        }
        f
        ''')
        res = f(lambda x: x + 3)
        assert res == 3 + 3
