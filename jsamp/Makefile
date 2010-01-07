# Released into the Public Domain by tav <tav@espians.com>

#
# A simple Makefile for building and testing jsamp.js
#

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

makefile_path := $(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))
main_root := $(dir $(abspath $(dir $(makefile_path))))

latest := .latest
download := curl -O

jars = ../bin/yuicompressor-2.4.2.jar ../bin/closure-2009-12-17.jar

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: update

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

latest: $(latest)

$(latest): update
	@touch $(latest)

$(jars):
	@echo
	@echo "# Downloading $(@F)"
	@echo
	@$(download) http://cloud.github.com/downloads/tav/ampify/$(@F)
	@mv "$(@F)" $(main_root)/bin/

download: $(jars)

jsamp: download
	@echo boo