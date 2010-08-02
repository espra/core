# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

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
	rm -f src/instance/www/*.css

debug:
	@$(AMP) build --debug

distclean: clean
	rm -rf environ/local
	rm -rf environ/receipts
	rm -rf src/build
	rm -f src/ampify/*.so

docs:
	@test -d third_party/pylibs || \
		(echo "ERROR: You need to checkout the third_party/pylibs submodule !!" \
			&& exit 1)
	@./environ/yatiblog doc

nuke: distclean
	rm -rf .build
	rm -f .build-lock
	rm -rf .sass-cache

test:
	@$(AMP) test

update:
	@./environ/git-update
