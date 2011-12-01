# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

PACKAGES=\
	argo\
	big\
	command\
	encoding\
	livequery\
	log\
	lzf\
	model\
	optparse\
	refmap\
	runtime\
	structure\
	tlsconf\
	yaml\

NOBENCH=\
	command\
	encoding\
	livequery\
	log\
	lzf\
	optparse\
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
	log\
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

bench.dirs: pkgroot.bench $(addsuffix .bench, $(BENCH))
clean.dirs: pkgroot.clean $(addsuffix .clean, $(PACKAGES))
nuke.dirs: pkgroot.nuke $(addsuffix .nuke, $(PACKAGES))
install.dirs: pkgroot.install $(addsuffix .install, $(PACKAGES))
test.dirs: pkgroot.test $(addsuffix .test, $(TEST))

%.bench:
	+cd pkg/$* && $(make) bench

%.clean:
	+cd pkg/$* && $(make) clean

%.install:
	+cd pkg/$* && $(make) install

%.nuke:
	+cd pkg/$* && $(make) nuke

%.test:
	+cd pkg/$* && $(make) test

pkgroot.bench:
	# +cd pkg && $(make) bench

pkgroot.clean:
	# +cd pkg && $(make) clean

pkgroot.install:
	# +cd pkg && $(make) install

pkgroot.nuke:
	# +cd pkg && $(make) nuke

pkgroot.test:
	# +cd pkg && $(make) test

bench: bench.dirs

clean: clean.dirs
	@cd cmd/amp && $(make) clean

distclean: nuke

nuke: nuke.dirs
	@echo rm -rf $(QUOTED_GOROOT)/pkg/amp
	@cd cmd/amp && $(make) nuke

install: install.dirs
	#cd cmd/amp && $(make) install

test: test.dirs

# ------------------------------------------------------------------------------
# Package Dependencies
# ------------------------------------------------------------------------------

argo.install: big.install
livequery.install: refmap.install
log.install: encoding.install
optparse.install: structure.install yaml.install
runtime.install: command.install log.install optparse.install
server.install: log.install tlsconf.install
tlsconf.install: runtime.install
web.install: optparse.install structure.install
