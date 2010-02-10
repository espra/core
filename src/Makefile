# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

WAF=../environ/startup/waf --jobs=1

.PHONY: all benchmark clean debug deps distclean docs install test

all: install

benchmark:
	@$(WAF) benchmark

clean:
	@$(WAF) clean

debug:
	@$(WAF) build --debug

deps:
	@$(WAF) deps

distclean:
	@$(WAF) distclean

docs:
	@$(WAF) docs

install:
	@$(WAF) build

test:
	@$(WAF) test
