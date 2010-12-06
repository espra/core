# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

PACKAGES=\
	argo\
	big\
	command\
	keyspace\
	logging\
	lzf\
	optparse\
	pyksproxy\
	runtime\
	slice\
	encoding\
	tlsconf\
	url\
	yaml\

NOBENCH=\
	command\
	encoding\
	keyspace\
	logging\
	lzf\
	optparse\
	pyksproxy\
	slice\
	tlsconf\
	url\
	yaml\

NOTEST=\
	command\
	keyspace\
	logging\
	pyksproxy\
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
	@cd tentproxy && $(make) clean

distclean: nuke

nuke: nuke.dirs
	@echo rm -rf $(QUOTED_GOROOT)/pkg/amp
	@cd hubproxy && $(make) nuke
	@cd tentproxy && $(make) nuke

install: install.dirs
	@cd hubproxy && $(make) install
	@cd tentproxy && $(make) install

test: test.dirs

# ------------------------------------------------------------------------------
# Package Dependencies
# ------------------------------------------------------------------------------

argo.install: big.install
keyspace.install: runtime.install slice.install
logging.install: encoding.install
optparse.install: runtime.install slice.install yaml.install
pyksproxy.install: runtime.install slice.install
runtime.install: command.install
tlsconf.install: runtime.install