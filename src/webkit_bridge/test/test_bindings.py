# No Copyright (-) 2009-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

from webkit_bridge.webkit_rffi import *
from pypy.rpython.lltypesystem import lltype, rffi


class TestBasic(object):

    def setup_method(cls, meth):
        cls.context = JSGlobalContextCreate()

    #def teardown_method(cls, meth):
    #    JSGlobalContextRelease(cls.context)

    def test_string(self):
        s = JSStringCreateWithUTF8CString('xyz')
        assert JSStringGetLength(s) == 3
        vs = JSValueMakeString(self.context, s)
        assert JSValueGetType(self.context, vs) == kJSTypeString

    def test_number(self):
        v = JSValueMakeNumber(self.context, 3.0)
        assert JSValueToNumber(self.context, v) == 3.0

    def test_interpret(self):
        script = '3'
        res = JSEvaluateScript(self.context, script, NULL)
        assert JSValueGetType(self.context, res) == kJSTypeNumber
        assert JSValueToNumber(self.context, res) == 3.0

    def test_interpret_object(self):
        script = '[1, 2, 3]'
        res = JSEvaluateScript(self.context, script, NULL)
        assert JSValueGetType(self.context, res) == kJSTypeObject
        s0 = JSStringCreateWithUTF8CString('0')
        fe = JSObjectGetProperty(self.context, res, s0)
        assert JSValueGetType(self.context, fe) == kJSTypeNumber
        assert JSValueToNumber(self.context, fe) == 1.0

    def test_get_set(self):
        script = '[]'
        obj = JSEvaluateScript(self.context, script, NULL)
        assert JSValueGetType(self.context, obj) == kJSTypeObject
        prop = JSStringCreateWithUTF8CString('prop')
        el = JSObjectGetProperty(self.context, obj, prop)
        assert JSValueGetType(self.context, el) == kJSTypeUndefined
        v = JSValueMakeNumber(self.context, 3)
        JSObjectSetProperty(self.context, obj, prop, v, 0)
        el = JSObjectGetProperty(self.context, obj, prop)
        assert JSValueGetType(self.context, el) == kJSTypeNumber

    def test_str(self):
        script = '[1,2,     3]'
        obj = JSEvaluateScript(self.context, script, NULL)
        s = JSValueToString(self.context, obj)
        assert JSStringGetUTF8CString(s) == '1,2,3'

    def test_function(self):
        script = 'this.x = function (a, b) { return (a + b); }'
        this = JSEvaluateScript(self.context, '[]', NULL)
        obj = JSEvaluateScript(self.context, script, this)
        name = JSStringCreateWithUTF8CString('x')
        f = JSObjectGetProperty(self.context, this, name)
        assert JSValueGetType(self.context, f) == kJSTypeObject
        args = [JSValueMakeNumber(self.context, 40),
                JSValueMakeNumber(self.context, 2)]
        res = JSObjectCallAsFunction(self.context, f, this, args)
        assert JSValueGetType(self.context, res) == kJSTypeNumber
        assert JSValueToNumber(self.context, res) == 42

    def test_global(self):
        glob = JSContextGetGlobalObject(self.context)
        prop = JSStringCreateWithUTF8CString('prop')
        JSObjectSetProperty(self.context, glob, prop,
                            JSValueMakeNumber(self.context, 3.0), 0)
        res = JSEvaluateScript(self.context, 'prop', NULL)
        assert JSValueGetType(self.context, res) == kJSTypeNumber
        assert JSValueToNumber(self.context, res) == 3.0

    def test_property_list(self):
        script = '''
        function myobject() {
            this.containedValue = 0;
            this.othercontainedValue = 0;
            this.anothercontainedValue = 0;
        }
        new myobject()
        '''
        x = JSEvaluateScript(self.context, script, NULL) 
        assert JSValueGetType(self.context, x) == kJSTypeObject
        nameref = JSObjectCopyPropertyNames(self.context, x)
        count = JSPropertyNameArrayGetCount(nameref)
        assert count == 3
        one = JSStringGetUTF8CString(JSPropertyNameArrayGetNameAtIndex(nameref,
                                                                       0))
        assert one == 'containedValue'
        three = JSStringGetUTF8CString(JSPropertyNameArrayGetNameAtIndex(nameref
                                                                         , 2))
        assert three == 'anothercontainedValue'
        assert JSPropertyList(self.context, x) == [
            'containedValue',
            'othercontainedValue',
            'anothercontainedValue'
            ]

    def test_callback_wrap(self):
        script = '''
        function f(x) {
            return (x(3));
        }
        f
        '''
        this = JSEvaluateScript(self.context, '[]', NULL)
        f = JSEvaluateScript(self.context, script, NULL)
        def _call(ctx, function, this, args):
            return args[0]
        y = create_js_callback(_call)(self.context, '_call')
        res = JSObjectCallAsFunction(self.context, f, this, [y])
        assert JSValueGetType(self.context, res) == kJSTypeNumber
        assert JSValueToNumber(self.context, res) == 3.0
