package sre2

// MatchIndex is the top-level complex NFA matcher used in sre2, where
// submatches are recorded. This method will return a list of ints indicating
// the match positions; indexes 0:1 represent the entire match, and n*2:n*2+1
// store the start and end index of the nth subexpression. If no match is
// found, returns nil.
func (r *sregexp) MatchIndex(src string) []int {
	states_alloc := len(r.prog)
	m := &m_submatch{NewSafeReader(src), make([]pair, 0, states_alloc)}
	m.addstate(r.prog[0], nil)
	curr := m.next
	m.next = make([]pair, 0, states_alloc)

	for m.parser.nextCh() != -1 {
		ch := m.parser.curr()

		// move along rune paths
		for _, p := range curr {
			st := r.prog[p.state]
			if st.match(ch) {
				m.addstate(st.out, p.places)
			}
		}

		curr, m.next = m.next, curr
		m.next = m.next[0:0]
	}

	// Search for a terminal state (in current states). If one is found, allocate
	// and return submatch information for those encountered.
	for _, p := range curr {
		if r.prog[p.state].mode == iMatch {
			alt := make([]int, r.caps*2)
			for i := 0; i < len(alt); i++ {
				// if a particular submatch is not encountered, return -1.
				alt[i] = -1
			}

			a := p.places
			for a != nil {
				if alt[a.pos] == -1 {
					alt[a.pos] = a.index
				}
				a = a.prev
			}
			return alt
		}
	}

	return nil
}

// cappos represents a single captured position, and it's relationship with a
// stack of other positions (some naive attempt to save memory by a huge linked
// list fanout).
type cappos struct {
	pos   int     // position
	index int     // rune index
	prev  *cappos // previous in stack
}

// pair represents a current state index, and the head of a list containing
// captured positions.
type pair struct {
	state  int     // current state for this path
	places *cappos // all set positions for this path
}

// m_submatch is the primary state holder for the submatcher. Contains parser
// state as well as all possible next states.
type m_submatch struct {
	parser SafeReader // string parser for target
	next   []pair     // stores all next states (in terms of pairs, current state + alts)
}

// addstate is a helper method to traverse all possible non-consuming states,
// identifying captures and finally storing inside m_submatch's next state when
// the search ends.
func (m *m_submatch) addstate(st *instr, a *cappos) {
	if st == nil {
		return // invalid
	}
	switch st.mode {
	case iSplit:
		m.addstate(st.out, a)
		m.addstate(st.out1, a)
	case iIndexCap:
		m.addstate(st.out, &cappos{st.cid, m.parser.npos(), a})
	case iBoundaryCase:
		if st.matchBoundaryMode(m.parser.curr(), m.parser.peek()) {
			m.addstate(st.out, a)
		}
	default:
		// NB. This maintains only the *first* possible capture information for
		// any current state within this RE. AFAIK this emulates Go's existing
		// regexp module. It's useless to keep more, anyway, considering we only
		// return the first found option when we reach the iMatch state.

		// TODO: speed this up with a bitset
		for _, p := range m.next {
			if p.state == st.idx {
				return
			}
		}

		// terminal, store (s.idx, altpos) in state
		// note that s.idx won't always be unique (but if both are equal, we could use this)
		pos := len(m.next)
		if pos == cap(m.next) {
			panic("should never hit this cap")
		}
		m.next = m.next[0 : pos+1]
		m.next[pos] = pair{st.idx, a}
	}
}

