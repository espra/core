package asm

import (
	"fmt"

	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

// Value represents a state value within Alloc.
type Value struct {
	a     *Alloc
	age   int
	id    int
	reg   int // currently allocated register (sometimes dup'd in state)
	stack string
	state valueState
}

// Become assigns the value to the given register, displacing as necessary.
func (v *Value) Become(reg int) {
	if v.reg == reg {
		return
	}
	if _, ok := v.a.regs[reg]; !ok {
		v.a.regs[reg] = struct{}{}
		v.displaceTo(v.a.newStateLive(reg))
		return
	}
	for _, cand := range v.a.values {
		if cand.reg != reg {
			continue
		}
		state := cand.state
		cand.displaceTo(cand.a.allocSpot())
		v.displaceTo(state)
		return
	}
}

// Consume frees the value, and returns the register for the value — assigning a
// register and loading data into it, if necessary.
func (v *Value) Consume() reg.VecPhysical {
	reg := v.Get()
	v.free()
	return reg
}

// ConsumeOp frees the value, and returns the location for its current state —
// assigning a register and loading data into it, if necessary.
func (v *Value) ConsumeOp() operand.Op {
	op := v.GetOp()
	v.free()
	return op
}

// Get returns the register for the value. If the value is not already live,
// then a register will be assigned and data will be loaded into it if
// necessary.
func (v *Value) Get() reg.VecPhysical {
	v.touch()
	switch state := v.state.(type) {
	case stateEmpty:
		v.alloc()
	case stateLazy:
		v.alloc()
		if !state.broadcast {
			v.a.ctx.VMOVDQU(state.mem, v.state.(stateLive).register())
		} else {
			v.a.ctx.VPBROADCASTD(state.mem, v.state.(stateLive).register())
		}
	case stateSpilled:
		reg := v.alloc()
		if state.aligned {
			v.a.ctx.VMOVDQA(state.getMem(), v.a.phys[reg])
		} else {
			v.a.ctx.VMOVDQU(state.getMem(), v.a.phys[reg])
		}
	}
	return v.state.(stateLive).register()
}

// GetOp returns the location of the state. If the value is not already live, or
// not on the stack, or is broadcasted (repeated) from a source location, then a
// register will be assigned and data will be loaded into it if necessary.
func (v *Value) GetOp() operand.Op {
	v.touch()
	switch state := v.state.(type) {
	case stateEmpty:
		v.alloc()
	case stateLazy:
		if !state.broadcast {
			return state.mem
		}
		reg := v.alloc()
		v.a.ctx.VPBROADCASTD(state.mem, v.a.phys[reg])
	case stateSpilled:
		return state.getMem()
	}
	return v.state.(stateLive).register()
}

// HasReg returns whether the value has been assigned to a register.
func (v *Value) HasReg() bool {
	return v.reg >= 0
}

// Reg returns the offset for the value's register. The value will be assigned
// to a register if that's not already the case.
func (v *Value) Reg() int {
	if v.reg < 0 {
		v.reg = v.a.allocReg(v)
	}
	return v.reg
}

func (v *Value) String() string {
	return fmt.Sprintf("Value(reg:%-2d state:%s)", v.reg, v.state)
}

func (v *Value) alloc() int {
	reg := v.reg
	if reg < 0 {
		reg = v.a.allocReg(v)
	}
	v.setState(v.a.newStateLive(reg))
	return reg
}

func (v *Value) displaceTo(dest valueState) {
	if state, ok := dest.(stateSpilled); ok && state.aligned {
		v.a.ctx.VMOVDQA(v.Get(), dest.op())
	} else {
		v.a.ctx.VMOVDQU(v.Get(), dest.op())
	}
	v.setState(dest)
}

func (v *Value) free() {
	v.setState(nil)
	delete(v.a.values, v.id)
}

func (v *Value) setState(state valueState) {
	switch state := v.state.(type) {
	case stateLive:
		v.a.regs.free(state.reg)
		v.reg = -1
	case stateSpilled:
		v.a.stack.free(state.slot)
	}
	v.state = state
	switch state := state.(type) {
	case stateLive:
		v.a.regs[state.reg] = struct{}{}
		v.reg = state.reg
	case stateSpilled:
		v.a.stack[state.slot] = struct{}{}
	}
}

func (v *Value) touch() {
	v.a.ctr++
	v.age = v.a.ctr
}

type stateEmpty struct{}

func (s stateEmpty) String() string {
	return "Empty"
}

func (s stateEmpty) live() bool {
	return false
}

func (s stateEmpty) op() operand.Op {
	panic("no location for this state")
}

type stateLazy struct {
	broadcast bool
	mem       operand.Mem
}

func (s stateLazy) String() string {
	return fmt.Sprintf("Lazy(%t, %s)", s.broadcast, s.mem.Asm())
}

func (s stateLazy) live() bool {
	return false
}

func (s stateLazy) op() operand.Op {
	panic("no location for this state")
}

type stateLive struct {
	phys []reg.VecPhysical
	reg  int
}

func (s stateLive) String() string {
	return fmt.Sprintf("Live(%d)", s.reg)
}

func (s stateLive) live() bool {
	return true
}

func (s stateLive) op() operand.Op {
	return s.register()
}

func (s stateLive) register() reg.VecPhysical {
	return s.phys[s.reg]
}

type stateSpilled struct {
	aligned bool
	mem     operand.Mem
	slot    int
	span    int
}

func (s stateSpilled) String() string {
	return fmt.Sprintf("Spilled(%d)", s.slot)
}

func (s stateSpilled) getMem() operand.Mem {
	return s.mem.Offset(s.span * s.slot)
}

func (s stateSpilled) live() bool {
	return false
}

func (s stateSpilled) op() operand.Op {
	return s.getMem()
}

type valueState interface {
	live() bool
	op() operand.Op
}
