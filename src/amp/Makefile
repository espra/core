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
	server\
	structure\
	tlsconf\
	url\
	yaml\
	web\

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
	server\
	structure\
	tlsconf\
	url\
	yaml\
	web\

NOTEST=\
	command\
	encoding\
	logging\
	pyksproxy\
	server\
	tlsconf\
	web\

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
optparse.install: structure.install yaml.install
pyksproxy.install: runtime.install
runtime.install: command.install logging.install optparse.install
server.install: logging.install tlsconf.install
tlsconf.install: runtime.install
web.install: optparse.install structure.install
