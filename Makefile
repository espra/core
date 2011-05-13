# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all build clean docs nuke test update

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

all: update build
	@touch .latest

build:
	@./environ/redpill build
	@cd src/amp && make install

clean:
	@cd src/amp && make nuke
	rm -rf src/ampify/build
	rm -rf third_party/pylibs/build

docs:
	@test -d third_party/pylibs || \
		(echo 'ERROR: You need to checkout the third_party/pylibs submodule !!' \
			&& exit 1)
	@./environ/yatiblog doc

nuke: clean
	rm -rf .sass-cache
	rm -rf environ/local
	rm -rf environ/receipts
	rm -f src/coffee/ucd.js

test:
	@cd src/amp && make test

update:
	@./environ/git-update
