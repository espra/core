# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

WAF := environ/waf --jobs=1

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all benchmark build build-zero clean debug dist distclean docs install test

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

all: install-zero
	@touch .latest

benchmark:
	@$(WAF) benchmark

build:
	@$(WAF) build

build-zero:
	@$(WAF) build --zero

clean:
	@$(WAF) clean

debug:
	@$(WAF) build --debug

dist:
	@$(WAF) dist

distclean:
	@$(WAF) distclean

docs:
	@$(WAF) docs

install:
	@$(WAF) install

install-zero:
	@$(WAF) install --zero

test:
	@$(WAF) test
