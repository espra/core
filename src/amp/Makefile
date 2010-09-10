# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

PACKAGES=\
	command\
	lzf\
	optparse\
	runtime\
	slice\
	encoding\
	tlsconf\
	yaml\

NOBENCH=\
	command\
	lzf\
	optparse\
	slice\
	tlsconf\
	encoding\
	yaml\

NOTEST=\
	command\
	slice\
	encoding\
	tlsconf\

ifeq ($(shell uname),FreeBSD)
make=gmake
else
make=make
endif

NULLSTRING :=
SPACE := $(NULLSTRING) # a space at the end
QUOTED_GOROOT=$(subst $(SPACE),\ ,$(GOROOT))

BENCH=\
	$(filter-out $(NOBENCH),$(PACKAGES))

TEST=\
	$(filter-out $(NOTEST),$(PACKAGES))

all: install

bench.dirs: $(addsuffix .bench, $(BENCH))
clean.dirs: $(addsuffix .clean, $(PACKAGES))
nuke.dirs: $(addsuffix .nuke, $(PACKAGES))
install.dirs: $(addsuffix .install, $(PACKAGES))
test.dirs: $(addsuffix .test, $(TEST))

%.bench:
	+cd $* && $(make) bench

%.clean:
	+cd $* && $(make) clean

%.install:
	+cd $* && $(make) install

%.nuke:
	+cd $* && $(make) nuke

%.test:
	+cd $* && $(make) test

bench: bench.dirs

clean: clean.dirs
	@cd hubproxy && $(make) clean
	@cd ampzero && $(make) clean

distclean: nuke

nuke: nuke.dirs
	@echo rm -rf $(QUOTED_GOROOT)/pkg/amp
	@cd hubproxy && $(make) nuke
	@cd ampzero && $(make) nuke

install: install.dirs
	@cd hubproxy && $(make) install
	@cd ampzero && $(make) install

test: test.dirs

# ------------------------------------------------------------------------------
# Package Dependencies
# ------------------------------------------------------------------------------

keyspace.install: runtime.install slice.install
logging.install: encoding.install
optparse.install: runtime.install slice.install yaml.install
runtime.install: command.install
tlsconf.install: runtime.install