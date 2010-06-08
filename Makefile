# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

WAF := environ/waf --jobs=1

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all benchmark build clean debug dist distclean docs install test zero

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

all: zero
	@touch .latest

benchmark:
	@$(WAF) benchmark

build:
	@$(WAF) build

clean:
	@$(WAF) uninstall --zero --force
	@$(WAF) clean --zero --force

debug:
	@$(WAF) build --debug

dist:
	@$(WAF) dist

distclean: clean
	@$(WAF) distclean

docs:
	@$(WAF) docs

install:
	@$(WAF) install

test:
	@$(WAF) test

zero:
	@$(WAF) install --zero
