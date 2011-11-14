# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

include $(GOROOT)/src/Make.inc

TARG=amp/big
GOFILES=\
	$(GOROOT_ORIGINAL)/src/pkg/math/big/arith.go\
	$(GOROOT_ORIGINAL)/src/pkg/math/big/arith_decl.go\
	decimal.go\
	$(GOROOT_ORIGINAL)/src/pkg/math/big/int.go\
	$(GOROOT_ORIGINAL)/src/pkg/math/big/nat.go\
	$(GOROOT_ORIGINAL)/src/pkg/math/big/rat.go\

OFILES=\
	$(GOROOT_ORIGINAL)/src/pkg/math/big/arith_$(GOARCH).$O\

$(GOROOT_ORIGINAL)/src/pkg/math/big/arith_$(GOARCH).$O:
	cd $(GOROOT_ORIGINAL)/src/pkg/math/big && make

include $(GOROOT)/src/Make.pkg
