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

JSAMP_ROOT := jsamp
JSAMP_PATH = $(JSAMP_ROOT)/$(jsamp_file).js

NODELINT := bin/nodelint.js

closure := bin/closure-2009-12-17.jar
yui := bin/yuicompressor-2.4.2.jar

css_files := screen
css_files := $(foreach css_file,$(css_files),$(CSS_PATH))

jsamp_source_files := sanitise
jsamp_source_files := $(foreach jsamp_file,$(jsamp_source_files),$(JSAMP_PATH))

jsamp_js := $(JSAMP_ROOT)/jsamp.js
jsamp_min_js := $(JSAMP_ROOT)/jsamp.min.js

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

$(jsamp_js): $(jsamp_source_files)
	@echo
	@echo "# Building JSAmp from source files"
	@echo
	@cat $(jsamp_source_files) > $(jsamp_js)

$(jsamp_min_js): $(closure) $(jsamp_js)
	@echo
	@echo "# Linting JSAmp"
	@echo
	@$(NODELINT) $(jsamp_js) || echo -e "\n# Linting failed !!"
	@echo
	@echo "# Compressing JSAmp using Closure Compiler"
	@echo
	@java -jar $(closure) --js $(jsamp_js) --js_output_file $(jsamp_min_js)

js: $(jsamp_min_js)

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
	rm -f $(jsamp_js)
	rm -f $(jsamp_min_js)

all:
