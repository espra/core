# Released into the Public Domain by tav <tav@espians.com>

#
# A simple Makefile for building and testing amplify
#

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

MAKEFILE_PATH := $(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))
MAIN_ROOT := $(abspath $(dir $(MAKEFILE_PATH)))

APPENGINE := bin/appengine
CSS_PATH = appengine/static/css/$(css_file).min.css
DOWNLOAD := curl -O

JSUTIL_ROOT := jsutil
JSUTIL_PATH = $(JSUTIL_ROOT)/$(jsutil_file).js

NODELINT := bin/nodelint.js

closure := bin/closure-2009-12-17.jar
yui := bin/yuicompressor-2.4.2.jar

css_files := screen
css_files := $(foreach css_file,$(css_files),$(CSS_PATH))

jsutil_source_files := sanitise
jsutil_source_files := $(foreach jsutil_file,$(jsutil_source_files),$(JSUTIL_PATH))

jsutil_js := $(JSUTIL_ROOT)/jsutil.js
jsutil_min_js := $(JSUTIL_ROOT)/jsutil.min.js

latest := .latest
jars := $(closure) $(yui)

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all clean css js static

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

latest: $(latest)

$(latest): all
	@touch $(latest)

$(jars):
	@echo
	@echo "# Downloading $(@F)"
	@echo
	@$(DOWNLOAD) http://cloud.github.com/downloads/tav/ampify/$(@F)
	@mv "$(@F)" $(MAIN_ROOT)/bin/

$(jsutil_js): $(jsutil_source_files)
	@echo
	@echo "# Building jsutil from source files"
	@echo
	@cat $(jsutil_source_files) > $(jsutil_js)

$(jsutil_min_js): $(closure) $(jsutil_js)
	@echo
	@echo "# Linting jsutil"
	@echo
	@$(NODELINT) $(jsutil_js) || echo -e "\n# Linting failed !!"
	@echo
	@echo "# Compressing jsutil using Closure Compiler"
	@echo
	@java -jar $(closure) --js $(jsutil_js) --js_output_file $(jsutil_min_js)

js: $(jsutil_min_js)

$(css_files): %.min.css: %.css
	@echo
	@for n in "$<"; \
	   do \
	     echo "# Compressing CSS: $$n"; \
	     java -jar $(yui) --charset utf-8 $$n -o $@; \
	   done;
	@echo

css: $(yui) $(css_files)

static: css js

deploy: static
	@$(APPENGINE) deploy

run:
	@$(APPENGINE) run

clean:
	rm -f $(css_files)
	rm -f $(jsutil_js)
	rm -f $(jsutil_min_js)

all:
