package asm

import (
	"github.com/mmcloughlin/avo/attr"
	"github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

// Context maintains state for incrementally building an avo File.
type Context struct {
	*build.Context
	DUPOK         attr.Attribute
	NEEDCTXT      attr.Attribute
	NOFRAME       attr.Attribute
	NOPROF        attr.Attribute
	NOPTR         attr.Attribute
	NOSPLIT       attr.Attribute
	REFLECTMETHOD attr.Attribute
	RODATA        attr.Attribute
	TLSBSS        attr.Attribute
	TOPFRAME      attr.Attribute
	WRAPPER       attr.Attribute
}

// DATA adds a data value to the active data section.
func (c *Context) DATA(offset int, v operand.Constant) {
	c.AddDatum(offset, v)
}

// GLOBL declares a new static global data section with the given attributes.
func (c *Context) GLOBL(name string, a attr.Attribute) operand.Mem {
	g := c.StaticGlobal(name)
	c.DataAttributes(a)
	return g
}

// NewAlloc instantiates a new Alloc instance for the given vector register set.
func (c *Context) NewAlloc(base reg.Register, r RegisterSet) *Alloc {
	return newAlloc(c, base, r)
}

// TEXT starts building a new function called name, with attributes a, and sets
// its signature (see SignatureExpr).
func (c *Context) TEXT(name string, a attr.Attribute, signature string) {
	c.Function(name)
	c.Attributes(a)
	c.SignatureExpr(signature)
}

// NewContext initializes an empty build Context.
func NewContext() *Context {
	return &Context{
		Context:       build.NewContext(),
		DUPOK:         attr.DUPOK,
		NEEDCTXT:      attr.NEEDCTXT,
		NOFRAME:       attr.NOFRAME,
		NOPROF:        attr.NOPROF,
		NOPTR:         attr.NOPTR,
		NOSPLIT:       attr.NOSPLIT,
		REFLECTMETHOD: attr.REFLECTMETHOD,
		RODATA:        attr.RODATA,
		TLSBSS:        attr.TLSBSS,
		TOPFRAME:      attr.TOPFRAME,
		WRAPPER:       attr.WRAPPER,
	}
}
