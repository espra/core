package asm

import (
	"github.com/mmcloughlin/avo/attr"
	"github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
)

// Context maintains state for incrementally building an avo File.
type Context struct {
	*build.Context
	NOSPLIT attr.Attribute
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
		Context: build.NewContext(),
		NOSPLIT: attr.NOSPLIT,
	}
}
