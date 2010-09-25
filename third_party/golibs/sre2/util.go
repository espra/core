package sre2

// Provides a storage mechanism for an ordered set of integer states.
type StateSet interface {
	Put(v int) bool // put the given int into set, false if successful
	Get() []int     // get the entire set of states
	Length() int    // shorthand for len(Get())
	Clear()         // clear the state set
}

// Create a new ordered bitset. States is the maximum state # that may be saved.
// Size is the maximum number of states that may be saved.
func NewStateSet(states int, size int) *StateSet {
	bwords := (states + 31) >> 5 // TODO: we just use lower 32 bits, even if int is int64

	// NOTE: I'd like to just return &obitset(...) here, but Go does not seem to be
	// happy to cast this as a StateSet. It's also not happy for me to return a cast
	// to "*StateSet(...)", so this expansion is required.
	ret := StateSet(&obitset{bwords, make([]int, bwords), make([]int, size), 0, size})
	return &ret
}

type obitset struct {
	bwords int
	bits   []int
	result []int
	pos    int
	size   int
}

func (o *obitset) Put(v int) bool {
	index := v >> 5
	value := 1 << (byte(v) & 31)

	// sanity-check
	if index > o.bwords {
		panic("can't insert, would overrun buffer")
	}
	if o.pos == o.size {
		panic("can't insert, no more storage")
	}

	// look for value set on bits[index].
	if (o.bits[index] & value) == 0 {
		o.bits[index] |= value
		o.result[o.pos] = v
		o.pos += 1
		return false
	}
	return true
}

func (o *obitset) Get() []int {
	return o.result[0:o.pos]
}

func (o *obitset) Length() int {
	return o.pos
}

func (o *obitset) Clear() {
	for i := 0; i < o.bwords; i++ {
		o.bits[i] = 0
	}
	o.pos = 0
}
