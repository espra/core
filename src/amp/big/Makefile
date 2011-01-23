# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

include $(GOROOT)/src/Make.inc

TARG=amp/big
GOFILES=\
	$(GOROOT)/src/pkg/big/arith.go\
	$(GOROOT)/src/pkg/big/arith_decl.go\
	decimal.go\
	$(GOROOT)/src/pkg/big/int.go\
	$(GOROOT)/src/pkg/big/nat.go\
	$(GOROOT)/src/pkg/big/rat.go\

OFILES=\
	$(GOROOT)/src/pkg/big/arith_$(GOARCH).$O\

include $(GOROOT)/src/Make.pkg
