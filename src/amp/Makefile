# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

PACKAGES=\
	argo\
	big\
	command\
	encoding\
	livequery\
	logging\
	lzf\
	optparse\
	pyksproxy\
	refmap\
	runtime\
	tlsconf\
	url\
	yaml\

NOBENCH=\
	command\
	encoding\
	livequery\
	logging\
	lzf\
	optparse\
	pyksproxy\
	refmap\
	runtime\
	tlsconf\
	url\
	yaml\

NOTEST=\
	command\
	encoding\
	logging\
	pyksproxy\
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
	@cd amp && $(make) clean

distclean: nuke

nuke: nuke.dirs
	@echo rm -rf $(QUOTED_GOROOT)/pkg/amp
	@cd amp && $(make) nuke

install: install.dirs
	@cd amp && $(make) install

test: test.dirs

# ------------------------------------------------------------------------------
# Package Dependencies
# ------------------------------------------------------------------------------

argo.install: big.install
livequery.install: refmap.install
logging.install: encoding.install
optparse.install: runtime.install yaml.install
pyksproxy.install: runtime.install
runtime.install: command.install
tlsconf.install: runtime.install
