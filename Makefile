# Public Domain (-) 2010-2012 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

git = $(shell which git)
ifeq ($(git),)
  branch = unknown
else
  branch = $(shell git rev-parse --abbrev-ref HEAD)
endif

redpill = $(shell which redpill)
yatiblog = $(shell which yatiblog)

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: build clean docs nuke test update

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

build:
ifeq ($(redpill),)
	@echo
	@echo "!! ERROR: You need to install redpill to build Ampify."
	@echo
	@echo "   To install redpill, run:"
	@echo
	@echo "       sudo easy_install redpill"
	@echo
	@echo
	@exit 1
else
	@./environ/ampenv redpill build
endif

clean:
	rm -rf build

docs:
ifeq ($(yatiblog),)
	@echo
	@echo "!! ERROR: You need to install yatiblog to generate the documentation."
	@echo
	@echo "   To install yatiblog, run:"
	@echo
	@echo "       sudo easy_install yatiblog"
	@echo
	@echo
	@exit 1
else
	@yatiblog doc
endif

nuke: clean
	rm -rf environ/local
	rm -rf environ/receipts
	rm -rf third_party/rust/build
	rm -rf third_party/rusty/build

test: build
	@./environ/ampenv rusty test pkg

update:
ifeq ($(branch),master)
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
