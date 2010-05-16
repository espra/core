# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

WAF := ../environ/waf --jobs=1

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all benchmark build clean debug deps distclean install test

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

all: build
	@touch .latest

benchmark:
	@$(WAF) benchmark

build:
	@$(WAF) build

clean:
	@$(WAF) clean

debug:
	@$(WAF) build --debug

deps:
	@$(WAF) deps

dist:
	@$(WAF) dist

distclean:
	@$(WAF) distclean

install:
	@$(WAF) install

test:
	@$(WAF) test
