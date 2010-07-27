# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

AMP := environ/amp

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all build clean debug distclean docs test zero

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

all: zero
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

docs: build
	@./environ/yatiblog doc

test:
	@$(AMP) test

zero: build
