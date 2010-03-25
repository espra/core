# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

CLOSURE := ../../environ/local/bin/closure-2010-03-24.jar
WAF := ../../environ/startup/waf --jobs=1
YUI := ../../environ/local/bin/yuicompressor-2.4.2.jar
ZEROSITE := ./site/zerosite

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all benchmark clean debug deps distclean install test

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

all: live/zerolive.js www/index.html www/zerosite.css www/zerosite.js \
     www/zerosite.min.css www/zerosite.min.js
	@touch .latest

clean:
	rm -f live/zerolive.js
	rm -f www/debug.html
	rm -f www/index.html
	rm -f www/zerosite.css
	rm -f www/zerosite.js
	rm -f www/zerosite.min.css
	rm -f www/zerosite.min.js
	@$(WAF) clean

benchmark:
	@$(WAF) benchmark

debug:
	@$(WAF) build --debug

deps:
	@$(WAF) deps

distclean:
	@$(WAF) distclean

install:
	@$(WAF) build

live/zerolive.js: live/zerolive.coffee
	@echo "Generating: zerolive.js"
	@coffee --no-wrap --compile live/zerolive.coffee

test:
	@$(WAF) test

www/index.html: site/index.genshi
	@echo "Generating: index.html"
	@$(ZEROSITE) -i site/index.genshi -o www/index.html -d www/debug.html

www/zerosite.css: site/zerosite.sass
	@echo "Generating: zerosite.css"
	@sass site/zerosite.sass www/zerosite.css

www/zerosite.js: site/zerosite.coffee
	@echo "Generating: zerosite.js"
	@coffee --no-wrap --compile site/zerosite.coffee --output www

www/zerosite.min.css: www/zerosite.css
	@echo "Generating: zerosite.min.css"
	@java -jar $(YUI) --charset utf-8 www/zerosite.css -o www/zerosite.min.css

www/zerosite.min.js: www/zerosite.js
	@echo "Generating: zerosite.min.js"
	@java -jar $(CLOSURE) --js www/zerosite.js --js_output_file www/zerosite.min.js
