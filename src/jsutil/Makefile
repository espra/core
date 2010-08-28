# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

COFFEE_FILES := $(wildcard *.coffee)
JS_FILES := $(subst .coffee,.js,$(COFFEE_FILES))

all:
	@coffee -c *.coffee

clean.files: $(addsuffix .clean, $(JS_FILES))

%.clean:
	+rm -f $*

clean: clean.files
