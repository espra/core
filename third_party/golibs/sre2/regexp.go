package sre2

import (
	"container/vector"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Regexp definition. Simple, just a list of states and a number of alts.
type sregexp struct {
	prog []*instr // list of states
	alts int      // number of marked alts [()'s] in this regexp
}

// Writes the given regexp to Stderr, for debugging.
func (r *sregexp) DebugOut() {
	for i := 0; i < len(r.prog); i++ {
		fmt.Fprintln(os.Stderr, i, r.prog[i].String())
	}
}

// Provide the number of marked alts [()'s] in this regexp.
func (r *sregexp) NumAlts() int {
	// we always have an outer () to match the whole re, subtract it
	return r.alts - 1
}

// Instruction type definitions, for "instr.mode".
type instrMode byte

const (
	kSplit        instrMode = iota // proceed down out & out1
	kAltBegin                      // begin of alt section, i.e. '('
	kAltEnd                        // end of alt section, i.e. ')'
	kBoundaryCase                  // match left/right runes here
	kRuneClass                     // if match rune, proceed down out
	kMatch                         // success state!
)

// Constants for kBoundaryCase, stored in 'instr.lr'.
type boundaryMode byte

const (
	bNone            boundaryMode = iota
	bBeginText                    // beginning of text
	bBeginLine                    // beginning of text or line
	bEndText                      // end of text
	bEndLine                      // end of text or line
	bWordBoundary                 // ascii word boundary
	bNotWordBoundary              // inverse of above, not ascii word boundary
)

// Represents a single instruction in any regexp.
type instr struct {
	idx  int       // index of this instr
	mode instrMode // mode (as above)
	out  *instr    // next instr to process

	// alternate path, for kSplit
	out1 *instr

	// boundary mode, for kBoundaryCase
	lr boundaryMode

	// rune class to match against, for kRuneClass
	rune RuneFilter

	// identifier of submatch for kAltBegin and kAltEnd
	alt    int    // numbered index
	alt_id string // string identifier (blank=none)
}

// This provides a string-representation of any given instruction.
func (i *instr) String() string {
	str := fmt.Sprintf("{%d", i.idx)
	out := ""
	if i.out != nil {
		out += fmt.Sprintf(" out=%d", i.out.idx)
	}
	switch i.mode {
	case kSplit:
		str += " kSplit"
		if i.out1 != nil {
			out += fmt.Sprintf(" out1=%d", i.out1.idx)
		}
	case kAltBegin, kAltEnd:
		if i.mode == kAltBegin {
			str += " kAltBegin"
		} else {
			str += " kAltEnd"
		}
		str += fmt.Sprintf(" alt=%d", i.alt)
		if len(i.alt_id) != 0 {
			str += fmt.Sprintf(" alt_id=%s", i.alt_id)
		}
	case kBoundaryCase:
		var mode string
		switch i.lr {
		case bBeginText:
			mode = "bBeginText"
		case bBeginLine:
			mode = "bBeginLine"
		case bEndText:
			mode = "bEndText"
		case bEndLine:
			mode = "bEndLine"
		case bWordBoundary:
			mode = "bWordBoundary"
		case bNotWordBoundary:
			mode = "bNotWordBoundary"
		}
		str += fmt.Sprintf(" kBoundaryCase [%s]", mode)
	case kRuneClass:
		str += fmt.Sprint(" kRuneClass ", i.rune)
	case kMatch:
		str += " kMatch"
	}
	return str + out + "}"
}

// Matcher method for consuming runes, thus only matches kRuneClass.
func (s *instr) match(rune int) bool {
	return s.mode == kRuneClass && s.rune(rune)
}

// Matcher method for kBoundaryCase. If either left or right is not within the
// target string, then -1 should be provided.
func (s *instr) matchBoundaryMode(left int, right int) bool {
	if s.mode != kBoundaryCase {
		return false
	}
	switch s.lr {
	case bBeginText:
		return left == -1
	case bBeginLine:
		return left == -1 || left == '\n'
	case bEndText:
		return right == -1
	case bEndLine:
		return right == -1 || right == '\n'
	case bWordBoundary, bNotWordBoundary:
		// TODO: This is ASCII-only at this point.
		wb := (unicode.Is(_word, left) && unicode.Is(_whitespace, right)) || (unicode.Is(_whitespace, left) && unicode.Is(_word, right))
		if s.lr == bWordBoundary {
			return wb
		} else {
			return !wb
		}
	}
	panic("unexpected lr mode")
}

// Escape constants and their mapping to actual Unicode runes.
var (
	ESCAPES = map[int]int{
		'a': 7, 't': 9, 'n': 10, 'v': 11, 'f': 12, 'r': 13,
	}
)

// Transient parser state, a combination of regexp and string iterator.
type parser struct {
	re    *sregexp
	src   SafeReader
	flags int64 // on/off state for flags 64-127 (subtract 64, uses bits)
}

// Generate a new instruction struct for use in regexp. By default, the instr
// will be of type 'kSplit'.
func (p *parser) instr() *instr {
	pos := len(p.re.prog)
	if pos == cap(p.re.prog) {
		if pos == 0 {
			panic("should not have cap of zero")
		}
		local := p.re.prog
		p.re.prog = make([]*instr, pos, pos*2)
		copy(p.re.prog, local)
	}
	p.re.prog = p.re.prog[0 : pos+1]
	i := &instr{pos, kSplit, nil, nil, bNone, nil, -1, ""}
	p.re.prog[pos] = i
	return i
}

// Determine whether the given flag is set. Requires flag in range 64-127,
// subtracts 64 and checks for bit set in flags int64.
func (p *parser) flag(flag int) bool {
	if flag < 64 || flag > 127 {
		panic(fmt.Sprintf("can't check flag, out of range: %c", flag))
	}
	return (p.flags & (1 << byte(flag-64))) != 0
}

// Helper method to connect instr 'from' to instr 'out'.
// TODO: Use safer connection helpers.
func (p *parser) out(from *instr, to *instr) {
	if from.out == nil {
		from.out = to
	} else if from.mode == kSplit && from.out1 == nil {
		from.out1 = to
	} else {
		panic("can't out")
	}
}

// Consume some alternate regexps. That is, (regexp[|regexp][|regexp]...).
// This method will return when it encounters an outer ')', and the cursor
// will rest on that character.
func (p *parser) alt(alt_id string, capture bool) (start *instr, end *instr) {
	// Hold onto the current set of flags; reset after.
	old_flags := p.flags
	defer func() {
		p.flags = old_flags
	}()

	end = p.instr() // shared end state for alt
	alt_begin := p.instr()

	// Optionally mark this as a capturing group.
	if capture {
		alt_begin.mode = kAltBegin
		alt_begin.alt = p.re.alts
		end.mode = kAltEnd
		end.alt = p.re.alts

		alt_begin.alt_id = alt_id
		end.alt_id = alt_id

		// Increment alt counter.
		p.re.alts += 1
	}

	b_start, b_end := p.regexp()
	start = b_start
	p.out(b_end, end)

	for p.src.curr() == '|' {
		start = p.instr()
		p.out(start, b_start)

		p.src.nextCh()
		b_start, b_end = p.regexp()
		p.out(start, b_start)
		p.out(b_end, end)
		b_start = start
	}

	// Note: We don't move over this final bracket.
	if p.src.curr() != ')' {
		panic("alt must end with ')'")
	}

	// Wire up the start of this alt to the first regexp part.
	p.out(alt_begin, start)

	return alt_begin, end
}

// Consume a single rune; assumes this is being invoked as the last possible
// option and will panic if an invalid escape sequence is found. Will return the
// found rune (as an integer) and with cursor past the entire representation.
func (p *parser) single_rune() int {
	if rune := p.src.curr(); rune != '\\' {
		// This is just a regular character; return it immediately.
		p.src.nextCh()
		return rune
	}

	if p.src.peek() == 'x' {
		// Match hex character code.
		var hex string
		p.src.nextCh()
		if p.src.nextCh() == '{' {
			hex = p.src.literal("{", "}")
		} else {
			hex = fmt.Sprintf("%c%c", p.src.curr(), p.src.nextCh())
			p.src.nextCh() // Step over the end of the hex code.
		}

		// Parse and return the corresponding rune.
		rune, err := strconv.Btoui64(hex, 16)
		if err != nil {
			panic(fmt.Sprintf("couldn't parse hex: %s", hex))
		}
		return int(rune)
	} else if rune := ESCAPES[p.src.peek()]; rune != 0 {
		// Literally match '\n', '\r', etc.
		p.src.nextCh()
		p.src.nextCh()
		return rune
	} else if unicode.Is(_punct, p.src.peek()) {
		// Allow punctuation to be blindly escaped.
		rune := p.src.nextCh()
		p.src.nextCh()
		return rune
	} else if unicode.IsDigit(p.src.peek()) {
		// Match octal character code (begins with digit, up to three digits).
		oct := ""
		p.src.nextCh()
		for i := 0; i < 3; i++ {
			oct += fmt.Sprintf("%c", p.src.curr())
			if !unicode.IsDigit(p.src.nextCh()) {
				break
			}
		}

		// Parse and return the corresponding rune.
		rune, err := strconv.Btoui64(oct, 8)
		if err != nil {
			panic(fmt.Sprintf("couldn't parse oct: %s", oct))
		}
		return int(rune)
	}

	// This is an escape sequence which does not identify a single rune.
	panic(fmt.Sprintf("not a valid escape sequence: \\%c", p.src.peek()))
}

// Consume a single character class and provide an implementation of the
// RuneFilter interface. Consumes the entire definition.
func (p *parser) class(within_class bool) (filter RuneFilter) {
	negate := false
	switch p.src.curr() {
	case '.':
		if p.flag('s') {
			filter = func(rune int) bool {
				return true
			}
		} else {
			filter = func(rune int) bool {
				return rune != '\n'
			}
		}
		p.src.nextCh()
	case '[':
		if p.src.peek() == ':' {
			// Match an ASCII class name.
			ascii_class := p.src.literal("[:", ":]")
			if ascii_class[0] == '^' {
				negate = true
				ascii_class = ascii_class[1:]
			}
			if filter = MatchAsciiClass(ascii_class); filter == nil {
				panic(fmt.Sprintf("could not identify ascii class: %s", ascii_class))
			}
		} else {
			if within_class {
				panic("can't match a [...] class within another class")
			}
			if p.src.nextCh() == '^' {
				negate = true
				p.src.nextCh()
			}

			// Consume and merge all valid classes within this [...] block.
			var filters vector.Vector
			for p.src.curr() != ']' {
				filters.Push(p.class(true))
			}
			p.src.nextCh() // Move over final ']'.
			filter = MergeFilter(filters)
		}
	case '\\':
		// Match some escaped character or escaped combination.
		switch p.src.peek() {
		case 'p', 'P':
			// Match a Unicode class name.
			negate = (p.src.nextCh() == 'P')
			unicode_class := fmt.Sprintf("%c", p.src.nextCh())
			if unicode_class[0] == '{' {
				unicode_class = p.src.literal("{", "}")
			} else {
				p.src.nextCh() // Step over the end of the hex code.
			}

			// Find and return the class.
			if filter = MatchUnicodeClass(unicode_class); filter == nil {
				panic(fmt.Sprintf("could not identify unicode class: %s", unicode_class))
			}
		case 'd', 'D':
			// Match digits.
			negate = (p.src.nextCh() == 'D')
			p.src.nextCh()
			filter = MatchUnicodeClass("Nd")
		case 's', 'S':
			// Match whitespace.
			negate = (p.src.nextCh() == 'S')
			p.src.nextCh()
			filter = MatchAsciiClass("whitespace")
		case 'w', 'W':
			// Match word characters.
			negate = (p.src.nextCh() == 'W')
			p.src.nextCh()
			filter = MatchAsciiClass("word")
		}
	}

	if filter == nil {
		// Match a single rune literal, or a range (when inside a character class).
		rune := p.single_rune()
		if p.src.curr() == '-' {
			if !within_class {
				panic(fmt.Sprintf("can't match a range outside class: %c-%c", rune, p.src.nextCh()))
			}
			p.src.nextCh() // move over '-'
			rune_high := p.single_rune()
			if rune_high < rune {
				panic(fmt.Sprintf("unexpected range: %c >= %c", rune, rune_high))
			}
			filter = MatchRuneRange(rune, rune_high)
		} else {
			filter = MatchRune(rune)
		}
	}

	if negate {
		return filter.Not()
	}
	return filter
}

// Build a left-right matcher of the given mode.
func (p *parser) makeBoundaryInstr(mode boundaryMode) *instr {
	instr := p.instr()
	instr.mode = kBoundaryCase
	instr.lr = mode
	return instr
}

// Consume a single term at the current cursor position. This may include a
// bracketed expression. When this function returns, the cursor will have moved
// past the final rune in this term.
func (p *parser) term() (start *instr, end *instr) {
	switch p.src.curr() {
	case -1:
		panic("EOF in term")
	case '*', '+', '{', '?':
		panic(fmt.Sprintf("unexpected expansion char: %c at %d", p.src.curr(), p.src.opos))
	case ')', '}', ']':
		panic("unexpected close element")
	case '(':
		// Match a bracketed expression (or modify current flags, with '?').
		capture := true
		alt_id := ""
		old_flags := p.flags
		if p.src.nextCh() == '?' {
			// Do something interesting before descending into this alt.
			p.src.nextCh()
			if p.src.curr() == 'P' {
				p.src.nextCh() // move to '<'
				alt_id = p.src.literal("<", ">")
			} else {
				// anything but 'P' means flags (and, non-captured).
				capture = false
				set := true
			outer:
				for {
					switch p.src.curr() {
					case ':':
						p.src.nextCh() // move past ':'
						break outer    // no more flags, process re
					case ')':
						// Return immediately: there's no instructions here, just flag sets!
						p.src.nextCh()
						start = p.instr()
						return start, start
					case '-':
						// now we're clearing flags
						set = false
					default:
						if p.src.curr() < 64 && p.src.curr() > 127 {
							panic(fmt.Sprintf("flag not in range: %c", p.src.curr()))
						}
						flag := byte(p.src.curr() - 64)
						if set {
							p.flags |= (1 << flag)
						} else {
							p.flags &= ^(1 << flag)
						}
					}
					p.src.nextCh()
				}
			}
		}

		// Now actually consume the bracketed expression.
		start, end = p.alt(alt_id, capture)
		if p.src.curr() != ')' {
			panic("alt should finish on end bracket")
		}
		p.src.nextCh()
		p.flags = old_flags
		return start, end
	case '$':
		// Match the end of text, or (with 'm') the end of a line.
		p.src.nextCh() // consume '$'
		mode := bEndText
		if p.flag('m') {
			mode = bEndLine
		}
		start = p.makeBoundaryInstr(mode)
		return start, start
	case '^':
		// Match the beginning of text, or (with 'm') the start of a line.
		p.src.nextCh() // consume '^'
		mode := bBeginText
		if p.flag('m') {
			mode = bBeginLine
		}
		start = p.makeBoundaryInstr(mode)
		return start, start
	case '\\':
		// Peek forward to match backslash-escaped terms which are not character classes.
		// If any of these branches trigger, they will return past the consumed 'term'.
		switch p.src.peek() {
		case 'Q':
			// Match a complete string literal, contained between '\Q' and the nearest '\E'.
			// Note that we don't use p.single_rune() to retrieve runes here, because this
			// works on literals exclusively and does not interpret e.g. \x00 or \] (punct).
			literal := p.src.literal("\\Q", "\\E")
			start = p.instr()
			end = start
			for _, ch := range literal {
				instr := p.instr()
				instr.mode = kRuneClass
				instr.rune = MatchRune(ch)
				p.out(end, instr)
				end = instr
			}
			return start, end
		case 'A':
			// Match only the beginning of text.
			p.src.consume("\\A")
			start = p.makeBoundaryInstr(bBeginText)
			return start, start
		case 'z':
			// Match only the end of text.
			p.src.consume("\\z")
			start = p.makeBoundaryInstr(bEndText)
			return start, start
		case 'b':
			// Match an ASCII word boundary.
			p.src.consume("\\b")
			start = p.makeBoundaryInstr(bWordBoundary)
			return start, start
		case 'B':
			// Match a non-ASCII word boundary.
			p.src.consume("\\B")
			start = p.makeBoundaryInstr(bNotWordBoundary)
			return start, start
		}
	}

	// Try to consume a rune class.
	start = p.instr()
	start.mode = kRuneClass
	start.rune = p.class(false)

	if p.flag('i') {
		// Mark this class as case-insensitive.
		start.rune = start.rune.IgnoreCase()
	}

	return start, start
}

// Safely retrieve a given term from the given position and alt count. If the
// passed first is true, then set it to false and perform a no-op. Otherwise,
// retrieve the new term.
func (p *parser) safe_term(src SafeReader, alt int, first *bool, start **instr, end **instr) {
	if *first {
		*first = false
		return
	}
	p.src = src
	p.re.alts = alt
	*start, *end = p.term()
}

// Consume a closure, defined as (term[repitition]). When this function returns,
// the cursor will be resting past the final rune in this closure.
func (p *parser) closure() (start *instr, end *instr) {

	// Store state of pos/alts in case we have to reparse term.
	revert_alts := p.re.alts
	revert := p.src

	// Grab first term.
	start = p.instr()
	end = start
	t_start, t_end := p.term()
	first := true // While true, we have a pending term.

	// Req and opt represent the number of required cases, and the number of
	// optional cases, respectively. Opt may be -1 to indicate no optional limit.
	var req int
	var opt int

	// By default, greedily choose an optional step over continuing. If 'U' is
	// flagged, swap this behaviour.
	greedy := true
	if p.flag('U') {
		greedy = false
	}
	switch p.src.curr() {
	case '?':
		p.src.nextCh()
		req, opt = 0, 1
	case '*':
		p.src.nextCh()
		req, opt = 0, -1
	case '+':
		p.src.nextCh()
		req, opt = 1, -1
	case '{':
		raw := p.src.literal("{", "}")
		parts := strings.Split(raw, ",", 2)
		// TODO: handle malformed int
		req, _ = strconv.Atoi(parts[0])
		if len(parts) == 2 {
			if len(parts[1]) > 0 {
				// TODO: handle malformed int
				opt, _ = strconv.Atoi(parts[1])
			} else {
				opt = -1
			}
		}
	default:
		return t_start, t_end // nothing to see here
	}

	if p.src.curr() == '?' {
		greedy = !greedy
		p.src.nextCh()
	}
	end_src := p.src

	if req < 0 || opt < -1 || req == 0 && opt == 0 {
		panic("invalid req/opt combination")
	}

	// Generate all required steps.
	for i := 0; i < req; i++ {
		p.safe_term(revert, revert_alts, &first, &t_start, &t_end)

		p.out(end, t_start)
		end = t_end
	}

	// Generate all optional steps.
	if opt == -1 {
		helper := p.instr()
		p.out(end, helper)
		if greedy {
			helper.out = t_start // greedily choose optional step
		} else {
			helper.out1 = t_start // optional step is 2nd preference
		}
		if end != t_end {
			// This is a little kludgy, but basically only wires up the term to the
			// helper iff it hasn't already been done.
			p.out(t_end, helper)
		}
		end = helper
	} else {
		real_end := p.instr()

		for i := 0; i < opt; i++ {
			p.safe_term(revert, revert_alts, &first, &t_start, &t_end)

			helper := p.instr()
			p.out(end, helper)
			if greedy {
				helper.out = t_start // greedily choose optional step
			} else {
				helper.out1 = t_start // optional step is 2nd preference
			}
			p.out(helper, real_end)

			end = p.instr()
			p.out(t_end, end)
		}

		p.out(end, real_end)
		end = real_end
	}

	p.src = end_src
	return start, end
}

// Match a regexp (defined as ([closure]*)) from the parser until either: EOF,
// the literal '|' or the literal ')'. At return, the cursor will still rest
// on this final terminal character.
func (p *parser) regexp() (start *instr, end *instr) {
	start = p.instr()
	curr := start

	for {
		if p.src.curr() == -1 || p.src.curr() == '|' || p.src.curr() == ')' {
			break
		}
		s, e := p.closure()
		p.out(curr, s)
		curr = e
	}

	end = p.instr()
	p.out(curr, end)
	return start, end
}

// Cleanup the given program. Assumes the given input is a flat slice containing
// no nil instructions. Will not clean up the first instruction, as it is always
// the canonical entry point for the regexp.
// Returns a similarly flat slice containing no nil instructions, however the
// slice may potentially be smaller.
func cleanup(prog []*instr) []*instr {
	// Detect kSplit recursion. We can remove this and convert it to a single path.
	// This might happen in cases where we loop over some instructions which are
	// not matchers, e.g. \Q\E*.
	states := NewStateSet(len(prog), len(prog))
	for i := 1; i < len(prog); i++ {
		states.Clear()
		pi := prog[i]
		var fn func(ci *instr) bool
		fn = func(ci *instr) bool {
			if ci != nil && ci.mode == kSplit {
				if states.Put(ci.idx) {
					// We've found a recursion.
					return true
				}
				if fn(ci.out) {
					ci.out = nil
				}
				if fn(ci.out1) {
					ci.out1 = nil
				}
				return false
			}
			return false
		}
		fn(pi)
	}

	// Iterate through the program, and remove single-instr kSplits.
	// NB: Don't parse the first instr, it will always be single.
	for i := 1; i < len(prog); i++ {
		pi := prog[i]
		if pi.mode == kSplit && (pi.out1 == nil || pi.out == pi.out1) {
			for j := 0; j < len(prog); j++ {
				if prog[j] == nil {
					continue
				}
				pj := prog[j]
				if pj.out == pi {
					pj.out = pi.out
				}
				if pj.out1 == pi {
					pj.out1 = pi.out
				}
			}
			prog[i] = nil
		}
	}

	// We may now have nil gaps: shift everything up.
	last := 0
	for i := 0; i < len(prog); i++ {
		if prog[i] != nil {
			last = i
		} else {
			// find next non-nil, move here
			var found int
			for found = i; found < len(prog); found++ {
				if prog[found] != nil {
					break
				}
			}
			if found == len(prog) {
				break // no more entries
			}

			// move found to i
			prog[i] = prog[found]
			prog[i].idx = i
			prog[found] = nil
			last = i
		}
	}

	return prog[0 : last+1]
}

// Public interface to a compiled regexp.
type Re interface {
	NumAlts() int
	Match(s string) bool
	MatchIndex(s string) []int
}

// Generates a simple, straight-forward NFA. Matches an entire regexp from the
// given input string. If the regexp could not be parsed, returns a non-nil
// error string: the regexp will be nil in this case.
func Parse(src string) (re Re, err *string) {
	defer func() {
		if r := recover(); r != nil {
			re = nil // clear re so it can't be used by caller
			switch x := r.(type) {
			case string:
				response := fmt.Sprintf("could not parse `%s`, error: %s", src, x)
				err = &response
			default:
				panic(fmt.Sprint("unknown parse error: ", r))
			}
		}
	}()

	p := parser{&sregexp{make([]*instr, 0, 1), 0}, NewSafeReader(".*?(" + src + ").*?"), 0}
	begin := p.instr()
	match := p.instr()
	match.mode = kMatch

	p.src.nextCh()
	start, end := p.regexp()

	if p.src.curr() != -1 {
		panic("could not consume all of regexp!")
	}

	p.out(begin, start)
	p.out(end, match)

	p.re.prog = cleanup(p.re.prog)
	return p.re, nil
}

// Generates a NFA from the given source. If the regexp could not be parsed,
// panics with a string error.
func MustParse(src string) Re {
	re, err := Parse(src)
	if err != nil {
		panic(*err)
	}
	return re
}
