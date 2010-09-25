package sre2

type altpos struct {
	alt    int     // alt index
	is_end bool    // end (true) or begin (false)
	pos    int     // character pos
	prev   *altpos // previous in stack
}

type pair struct {
	state int
	alt   *altpos
}

type m_submatch struct {
	parser SafeReader
	next   []pair
}

func (m *m_submatch) addstate(st *instr, a *altpos) {
	if st == nil {
		return // invalid
	}
	switch st.mode {
	case kSplit:
		m.addstate(st.out, a)
		m.addstate(st.out1, a)
	case kAltBegin:
		m.addstate(st.out, &altpos{st.alt, false, m.parser.npos(), a})
	case kAltEnd:
		m.addstate(st.out, &altpos{st.alt, true, m.parser.npos(), a})
	case kBoundaryCase:
		if st.matchBoundaryMode(m.parser.curr(), m.parser.peek()) {
			m.addstate(st.out, a)
		}
	default:
		// terminal, store (s.idx, altpos) in state
		// note that s.idx won't always be unique (but if both are equal, we could use this)
		pos := len(m.next)
		if pos == cap(m.next) {
			// out of storage, grow to hold onto more states
			hold := m.next
			m.next = make([]pair, pos, pos*2)
			copy(m.next, hold)
		}
		m.next = m.next[0 : pos+1]
		m.next[pos] = pair{st.idx, a}
	}
}

// Submatch regexp matcher entry point. Returns nil if no match found, or an
// array of submatch locations on success (with the entire match in index 0:1).
func (r *sregexp) MatchIndex(src string) []int {
	states_alloc := 64
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
				m.addstate(st.out, p.alt)
			}
		}

		curr, m.next = m.next, curr
		m.next = m.next[0:0]
	}

	// Search for a terminal state (in current states). If one is found, allocate
	// and return submatch information for those encountered.
	for _, p := range curr {
		if r.prog[p.state].mode == kMatch {
			alt := make([]int, r.alts*2)
			for i := 0; i < len(alt); i++ {
				// if a particular submatch is not encountered, return -1.
				alt[i] = -1
			}

			a := p.alt
			for a != nil {
				pos := (a.alt * 2)
				if a.is_end {
					pos += 1
				}
				if alt[pos] == -1 {
					alt[pos] = a.pos
				}
				a = a.prev
			}
			return alt
		}
	}

	return nil
}
