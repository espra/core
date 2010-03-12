# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

APPS=\
	ampnode\

PACKAGES=\
		command\
		runtime\

NULLSTRING :=
SPACE := $(NULLSTRING) # a space at the end
QUOTED_GOROOT=$(subst $(SPACE),\ ,$(GOROOT))

clean_dirs := $(addsuffix .clean, $(DIRS))
install_dirs := $(addsuffix .install, $(DIRS))
test_dirs := $(addsuffix .test, $(DIRS))

distclean_dirs := $(addsuffix .test, $(PACKAGES))

all: install

%.clean:
	+cd $* && make clean

%.install:
	+cd $* && make install

%.test:
	+cd $* && make test

clean: $(clean_dirs)

distclean: clean
	@echo rm -rf $(QUOTED_GOROOT)/pkg/amp/*

install: $(install_dirs)

test: $(test_dirs)

# ------------------------------------------------------------------------------
# package dependencies
# ------------------------------------------------------------------------------

