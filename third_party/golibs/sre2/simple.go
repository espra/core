package sre2

// This file defines the simple regexp matcher used by sre2. This backs onto
// a bitset to manage states, and does not record or care about submatches.

// Match is the simple regexp matcher entry point. Just returns true/false for
// matching re, completely ignoring submatches.
func (r *sregexp) Match(src string) bool {
	curr := makeStateList(len(r.prog))
	next := makeStateList(len(r.prog))
	parser := NewSafeReader(src)

	// always start with state zero
	curr.addstate(&parser, r.prog[0])

	for parser.nextCh() != -1 {
		ch := parser.curr()
		if len(curr.states) == 0 {
			return false // no more possible states, short-circuit failure
		}

		// move along rune paths
		for _, st := range curr.states {
			i := r.prog[st]
			if i.match(ch) {
				next.addstate(&parser, i.out)
			}
		}
		curr, next = next, curr
		next.clear() // clear next so it can be re-used
	}

	// search for success state
	for _, st := range curr.states {
		if r.prog[st].mode == iMatch {
			return true
		}
	}
	return false
}

// stateList is used by Match() to efficiently maintain an ordered list of
// current/next regexp integer states.
type stateList struct {
	bits   []int64
	states []int
}

// Build a new ordered bitset for use by at most m_size states, each with a
// maximum value of m_state.
func makeStateList(states int) *stateList {
	bwords := (states + 63) >> 6
	return &stateList{make([]int64, bwords), make([]int, 0, states)}
}

// addstate descends through split/alt states and places them all in the
// given stateList.
func (o *stateList) addstate(p *SafeReader, st *instr) {
	if st == nil || o.put(st.idx) {
		return // instr does not exist, or state already in set: fall out
	}
	switch st.mode {
	case iSplit:
		o.addstate(p, st.out)
		o.addstate(p, st.out1)
	case iIndexCap:
		// ignore, just walk over
		o.addstate(p, st.out)
	case iBoundaryCase:
		if st.matchBoundaryMode(p.curr(), p.peek()) {
			o.addstate(p, st.out)
		}
	}
}

// put places the given state into the stateList. Returns true if the state was
// previously set, and false if it was not.
func (o *stateList) put(v int) bool {
	index := v >> 6
	value := int64(1 << (byte(v) & 63))

	// sanity-check
	if index > len(o.bits) {
		panic("can't insert, would overrun buffer")
	} else if len(o.states) == cap(o.states) {
		panic("can't put state, no more storage")
	}

	// look for value set on bits[index].
	if (o.bits[index] & value) == 0 {
		o.bits[index] |= value
		pos := len(o.states)
		o.states = o.states[0 : pos+1]
		o.states[pos] = v
		return false
	}
	return true
}

// clear resets the stateList to be re-used.
func (o *stateList) clear() {
	for i := 0; i < len(o.bits); i++ {
		o.bits[i] = 0
	}
	o.states = o.states[0:0]
}
