# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

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
