# Public Domain (-) 2010-2011 The Ampify Authors.
# See the UNLICENSE file for details.

COFFEE_FILES := $(wildcard *.coffee)
JS_FILES := $(subst .coffee,.js,$(COFFEE_FILES))

all:
	@coffee -c *.coffee

clean.files: $(addsuffix .clean, $(JS_FILES))

%.clean:
	+rm -f $*

clean: clean.files
