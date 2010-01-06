# Released into the Public Domain by tav <tav@espians.com>

"""Service input validation support."""

from time import time
from sha import new as sha1

from tentapp.core.exception import Error

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

CO_VARARGS = 4
CO_VARKEYWORDS = 8

# ------------------------------------------------------------------------------
# kore validate funktion
# ------------------------------------------------------------------------------

def validate(**spec):
    """Validation decorator which enforces the specified ``spec``."""

    def __decorate(func):

        rkey = sha1(str(time())).hexdigest()[:7]
        rkw = 'kw_%s' % rkey

        code = func.func_code
        varnames = list(code.co_varnames)
        defaults = func.func_defaults or ()
        func_name = func.func_name

        varargs = varkwargs = None

        if code.co_flags & CO_VARKEYWORDS:
            varkwargs = varnames.pop(-1)

        if code.co_flags & CO_VARARGS:
            varargs = varnames.pop(-1)

        params = [varnames.pop(0)]; add = params.append
        params2 = params[:]; add2 = params2.append
        default_pointer = len(varnames) - len(defaults)

        for idx, varname in enumerate(varnames):
            if idx < default_pointer:
                add(varname)
                add2("%s[%r]" % (rkw, varname))
            else:
                add("%s=defaults_%s[%s]" % (varname, rkey, idx-default_pointer))
                add2("%s=%s.get(%r, defaults_%s[%s])" % (varname, rkw, varname, rkey, idx-default_pointer))

        if varargs:
            add("*%s" % varargs)
            add2("*%s[%r]" % (rkw, varargs))
        if varkwargs:
            add("**%s" % varkwargs)
            add2("**%s[%r]" % (rkw, varkwargs))
            
        params = ", ".join(params)
        params2 = ", ".join(params2)

        source = """
def %(func_name)s(%(params)s):
    kws_%(r)s = locals()
    kw_%(r)s = {}
    for key_%(r)s in kws_%(r)s:
        if key_%(r)s in spec_%(r)s: 
            try:
                kw_%(r)s[key_%(r)s] = spec_%(r)s[key_%(r)s](kws_%(r)s[key_%(r)s])
            except Exception:
                raise Error(
                    "Could not validate input argument '%%s=%%s'" %% (key_%(r)s, kws_%(r)s[key_%(r)s])
                    )
        else:
            kw_%(r)s[key_%(r)s] = kws_%(r)s[key_%(r)s]
    return func_%(r)s(%(params2)s)
""" % dict(params=params, params2=params2, r=rkey, func_name=func_name)

        env = {
            'func_%s' % rkey: func,
            'spec_%s' % rkey: spec,
            'defaults_%s' % rkey: defaults,
            'Error': Error
            }
        exec source in env

        return env[func_name]

    return __decorate
