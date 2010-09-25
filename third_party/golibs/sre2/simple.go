package sre2

// Simple regexp matcher entry point. Just returns true/false for matching re,
// and completely ignores submatches.
func (r *sregexp) Match(src string) bool {
	curr := NewStateSet(len(r.prog), len(r.prog))
	next := NewStateSet(len(r.prog), len(r.prog))
	parser := NewSafeReader(src)

	// always start with state zero
	addstate(curr, r.prog[0], &parser)

	for parser.nextCh() != -1 {
		ch := parser.curr()
		if curr.Length() == 0 {
			return false // no more possible states, short-circuit failure
		}

		// move along rune paths
		for _, st := range curr.Get() {
			i := r.prog[st]
			if i.match(ch) {
				addstate(next, i.out, &parser)
			}
		}
		curr, next = next, curr
		next.Clear() // clear next so it can be re-used
	}

	// search for matching state
	for _, st := range curr.Get() {
		if r.prog[st].mode == kMatch {
			return true
		}
	}
	return false
}

// Helper method - just descends through split/alt states and places them all
// in the given StateSet.
func addstate(set *StateSet, st *instr, p *SafeReader) {
	if st == nil || set.Put(st.idx) {
		return // invalid
	}
	switch st.mode {
	case kSplit:
		addstate(set, st.out, p)
		addstate(set, st.out1, p)
	case kAltBegin, kAltEnd:
		// ignore, just walk over
		addstate(set, st.out, p)
	case kBoundaryCase:
		if st.matchBoundaryMode(p.curr(), p.peek()) {
			addstate(set, st.out, p)
		}
	}
}
