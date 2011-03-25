# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

PACKAGES=\
	argo\
	big\
	command\
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
	@cd tentproxy && $(make) clean

distclean: nuke

nuke: nuke.dirs
	@echo rm -rf $(QUOTED_GOROOT)/pkg/amp
	@cd tentproxy && $(make) nuke

install: install.dirs
	@cd tentproxy && $(make) install

test: test.dirs

# ------------------------------------------------------------------------------
# Package Dependencies
# ------------------------------------------------------------------------------

argo.install: big.install
logging.install: encoding.install
optparse.install: runtime.install slice.install yaml.install
pyksproxy.install: runtime.install slice.install
runtime.install: command.install
tlsconf.install: runtime.install
