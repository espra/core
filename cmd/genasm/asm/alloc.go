package asm

import (
	"fmt"
	"runtime"

	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

// Alloc maintains state across physical vector registers and the stack.
type Alloc struct {
	ctr    int
	ctx    *Context
	m      operand.Mem
	mslot  int
	n      int
	phys   []reg.VecPhysical
	regs   mem
	span   int
	spills int
	stack  mem
	values map[int]*Value
}

// Free prints any values that have leaked.
func (a *Alloc) Free() {
	for id, v := range a.values {
		fmt.Println("leaked value:", id, "==", v.id, "\n", v.stack)
	}
}

// FreeReg returns a free register, or -1 if none are available.
func (a *Alloc) FreeReg() int {
	n, ok := a.regs.alloc(a.n)
	if !ok {
		return -1
	}
	a.regs.free(n)
	return n
}

// Stats prints the current allocation stats and returns a function that should
// be deferred to print the final allocation stats.
func (a *Alloc) Stats(name string) func() {
	a.stats(name, "in")
	return func() { a.stats(name, "out") }
}

// Value creates a fresh value.
func (a *Alloc) Value() *Value {
	var buf [4096]byte
	a.ctr++
	v := &Value{
		a:     a,
		age:   a.ctr,
		id:    a.ctr,
		reg:   -1,
		stack: string(buf[:runtime.Stack(buf[:], false)]),
		state: stateEmpty{},
	}
	a.values[v.id] = v
	return v
}

// ValueFrom creates a value that is lazily loaded from the given source.
func (a *Alloc) ValueFrom(m operand.Mem) *Value {
	v := a.Value()
	v.state = stateLazy{mem: m}
	return v
}

// ValueWith creates a value that is lazily loaded with the given source.
func (a *Alloc) ValueWith(m operand.Mem) *Value {
	v := a.Value()
	v.state = stateLazy{broadcast: true, mem: m}
	return v
}

// Values creates a slice of fresh values.
func (a *Alloc) Values(n int) []*Value {
	out := make([]*Value, n)
	for i := range out {
		out[i] = a.Value()
	}
	return out
}

// ValuesWith creates a slice of values that are lazily loaded with the given
// source. The given sizeof in bits, determines the memory offset used for each
// of the slice elements.
func (a *Alloc) ValuesWith(n int, m operand.Mem, sizeof int) []*Value {
	size := sizeof / 8
	out := make([]*Value, n)
	for i := range out {
		out[i] = a.ValueWith(m.Offset(size * i))
	}
	return out
}

func (a *Alloc) allocReg(except *Value) int {
	reg, ok := a.regs.alloc(a.n)
	if ok {
		return reg
	}
	oldest := a.findOldestLive(except)
	state := oldest.state.(stateLive)
	oldest.displaceTo(a.allocSpot())
	a.regs[state.reg] = struct{}{}
	return state.reg
}

func (a *Alloc) allocSpot() valueState {
	reg, ok := a.regs.alloc(a.n)
	if ok {
		return a.newStateLive(reg)
	}
	slot := a.stack.mustAlloc()
	a.spills++
	if slot > a.mslot {
		a.mslot = slot
	}
	return stateSpilled{
		aligned: true,
		mem:     a.m,
		slot:    slot,
		span:    a.span,
	}
}

func (a *Alloc) findOldestLive(except *Value) *Value {
	var oldest *Value
	for _, v := range a.values {
		if oldest == except || !v.state.live() {
			continue
		}
		if oldest == nil || v.age < oldest.age {
			oldest = v
		}
	}
	return oldest
}

func (a *Alloc) newStateLive(reg int) stateLive {
	return stateLive{
		phys: a.phys,
		reg:  reg,
	}
}

func (a *Alloc) stats(name, when string) {
	fmt.Printf("// [%s] %s: %d/%d free (%d total + %d spills + %d slots)\n",
		name, when, a.n-len(a.regs), a.n, len(a.values), a.spills, a.mslot+1)
}

type mem map[int]struct{}

func (m mem) alloc(max int) (n int, ok bool) {
	for max == 0 || n < max {
		if _, ok := m[n]; !ok {
			m[n] = struct{}{}
			return n, true
		}
		n++
	}
	return 0, false
}

func (m mem) free(n int) {
	delete(m, n)
}

func (m mem) mustAlloc() (n int) {
	n, ok := m.alloc(0)
	if !ok {
		panic("unable to alloc")
	}
	return n
}

func newAlloc(ctx *Context, base reg.Register, r RegisterSet) *Alloc {
	return &Alloc{
		ctx:    ctx,
		m:      operand.Mem{Base: base},
		mslot:  -1,
		n:      r.n,
		phys:   r.registers,
		regs:   mem{},
		span:   r.size / 8,
		stack:  mem{},
		values: map[int]*Value{},
	}
}
