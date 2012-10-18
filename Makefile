# Public Domain (-) 2010-2012 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

branch = $(shell git rev-parse --abbrev-ref HEAD)
git = $(shell which git)
yatiblog = $(shell which yatiblog)

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all build clean docs nuke test update

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

all: update build

build:
	@./environ/redpill build

clean:
	rm -rf build

docs:
ifeq ($(yatiblog),)
	@echo
	@echo "!! ERROR: You need to install yatiblog to generate the documentation."
	@echo
	@echo "   To install yatiblog, run the equivalent of the following for your"
	@echo "   Python version:"
	@echo
	@echo
	@echo "       sudo easy_install-2.7 yatiblog"
	@echo
	@echo
	@exit 1
else
	@yatiblog doc
endif

nuke: clean
	rm -rf environ/local
	rm -rf environ/receipts

update:
ifeq ($(branch),masters)
	@git pull origin master
	@git submodule update --init
else ifeq ($(git),)
	@echo "## Skipping 'make update' due to not finding a git binary."
else
	@echo
	@echo "!! ERROR: You need to be on the master branch before 'make update' will run."
	@echo "          You are currently on the '$(branch)' branch."
	@echo
	@exit 1
endif
