# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

CLOSURE := ../../environ/local/bin/closure-2010-03-24.jar
YUI := ../../environ/local/bin/yuicompressor-2.4.2.jar

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: all

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

all: live/zerolive.js www/index.html www/zerosite.css www/zerosite.js www/zerosite.min.css www/zerosite.min.js
	@touch .latest

clean:
	@rm -f live/zerolive.js
	@rm -f www/index.html
	@rm -f www/zerosite.css
	@rm -f www/zerosite.js
	@rm -f www/zerosite.min.css
	@rm -f www/zerosite.min.js

live/zerolive.js: live/zerolive.coffee
	@echo "Generating: zerolive.js"
	@coffee --no-wrap --compile live/zerolive.coffee

www/debug.html: site/index.genshi
	@echo "Generating: debug.html"

www/index.html: site/index.genshi
	@echo "Generating: index.html"
	@touch www/index.html

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
