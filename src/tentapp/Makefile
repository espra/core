# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

BIN=$(AMPIFY_ROOT)/environ/local/bin

all: static/site.min.css static/tentapp.min.js

clean:
	+rm -f static/site.source.css
	+rm -f static/site.css
	+rm -f static/site.min.css
	+rm -f static/site.ie.css
	+rm -f static/site.ie.min.css
	+rm -f static/tentapp.js
	+rm -f static/tentapp.min.js

static/site.source.css: static/site.sass
	@echo "# Generating CSS files from the SASS input."
	@sass static/site.sass static/site.source.css

static/site.css: static/site.source.css
	@echo "# Expanding embedded files from the CSS input."
	@cssembed static/site.source.css --root static > static/site.css
	@cssembed static/site.source.css --static-root /static/ --IE > static/site.ie.css

static/site.min.css: static/site.css
	@echo "# Minifying CSS files."
	@java -jar $(BIN)/yuicompressor-2.4.2.jar --charset utf-8 static/site.css -o static/site.min.css
	@java -jar $(BIN)/yuicompressor-2.4.2.jar --charset utf-8 static/site.ie.css -o static/site.ie.min.css

static/tentapp.js: static/tentapp.coffee
	@echo "# Generating JS files from the CoffeeScript input."
	@coffee -c -o static static/tentapp.coffee

static/tentapp.min.js: static/tentapp.js
	@echo "# Minifying JS files."
	@java -jar $(BIN)/closure-2010-03-24.jar --js static/tentapp.js --js_output_file static/tentapp.min.js
