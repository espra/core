This package provides a bridge between WebKit and a PyPy-based interpreter.

To use, please create a symlink to the WebKit build as `webkit`. It should
contain a debug build and libJavaScriptCore.so. To create libJavaScriptCore.so,
go to WebKitBuild/Debug/.libs and perform:

    $ ar x libJavaScriptCore.a
    $ g++ -shared -o libJavaScriptCore.so *.o -lpthread -lglib-2.0 `icu-config --ldflags`

And, finally, make sure the .so is on your LD_LIBRARY_PATH.
