// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// The refmap provides a utility map-like object that provides integer
// references to be used in place of longer string variables.
package refmap

const (
	INCREF = iota
	DECREF
)

const zero uint64 = 0

type Map struct {
	Inbox  chan *Req
	Info   map[uint64]*Ref
	Lookup map[string]uint64
}

type Req struct {
	r chan uint64 /* Response channel */
	s string      /* String identifier */
	t int         /* Request type */
	v int         /* Value for incref/decref */
}

type Ref struct {
	s string /* String identifier */
	v uint64 /* Value of the current refcount */
}

func (refmap *Map) Incref(s string, incref int) uint64 {
	resp := make(chan uint64)
	req := &Req{r: resp, s: s, t: INCREF, v: incref}
	refmap.Inbox <- req
	ref := <-resp
	close(resp)
	return ref
}

func (refmap *Map) Decref(s string, decref int) {
	req := &Req{s: s, t: DECREF, v: decref}
	refmap.Inbox <- req
}

func NewWithVal(start uint64) *Map {
	inbox := make(chan *Req, 100)
	info := make(map[uint64]*Ref)
	lookup := make(map[string]uint64)
	refmap := &Map{Inbox: inbox, Info: info, Lookup: lookup}
	go func(ref uint64) {
		for req := range inbox {
			s := req.s
			_ref, found := lookup[s]
			switch req.t {
			case INCREF:
				if found {
					req.r <- _ref
					info[_ref].v += uint64(req.v)
				} else {
					ref += 1
					lookup[s] = ref
					info[ref] = &Ref{s: s, v: uint64(req.v)}
					req.r <- ref
				}
			case DECREF:
				if found {
					i := info[_ref]
					v := i.v - uint64(req.v)
					if v <= 0 {
						lookup[i.s] = zero, false
						info[_ref] = i, false
					}
				}
			}
		}
	}(start)
	return refmap
}

func New() *Map {
	return NewWithVal(zero)
}
