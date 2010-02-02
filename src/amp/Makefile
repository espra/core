# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

include $(GOROOT)/src/Make.$(GOARCH)

TARG=ampnode
GOFILES=\
        main.go\

all:
	@cd runtime
	@make install
	@cd ..

include $(GOROOT)/src/Make.cmd