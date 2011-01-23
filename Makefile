# Public Domain (-) 2010-2011 The Ampify Authors.
# See the UNLICENSE file for details.

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

AMP := environ/amp

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all build clean debug distclean docs nuke test update

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

all: update build
	@touch .latest

build:
	@$(AMP) build

clean:
	@./environ/assetgen --clean
	rm -rf src/build
	rm -rf third_party/pylibs/build

debug:
	@$(AMP) build --debug

distclean: clean
	rm -rf environ/local
	rm -rf environ/receipts
	rm -f src/jsutil/ucd.js
	rm -f src/ampify/*.so

docs:
	@test -d third_party/pylibs || \
		(echo "ERROR: You need to checkout the third_party/pylibs submodule !!" \
			&& exit 1)
	@./environ/yatiblog doc

nuke: distclean
	rm -rf .sass-cache

test:
	@$(AMP) test

update:
	@./environ/git-update
