// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// The refmap provides a utility map-like object that provides integer
// references to be used in place of longer string variables.
package refmap

import "sync"

const Zero uint64 = 0
const One uint64 = 1

type Map struct {
	Info   map[uint64]*Ref
	Lookup map[string]uint64
	Mutex  sync.RWMutex
	val    uint64
}

type Ref struct {
	s string /* String identifier */
	v uint64 /* Value of the current refcount */
}

func (refmap *Map) Create(s string) uint64 {
	refmap.Mutex.Lock()
	defer refmap.Mutex.Unlock()
	ref, found := refmap.Lookup[s]
	if found {
		return ref
	}
	refmap.val += 1
	ref = refmap.val
	refmap.Lookup[s] = ref
	refmap.Info[ref] = &Ref{s: s, v: One}
	return ref
}

func (refmap *Map) Delete(s string) {
	refmap.Mutex.Lock()
	defer refmap.Mutex.Unlock()
	ref, found := refmap.Lookup[s]
	if !found {
		return
	}
	delete(refmap.Lookup, s)
	delete(refmap.Info, ref)
}

func (refmap *Map) DeleteRef(v uint64) {
	refmap.Mutex.Lock()
	defer refmap.Mutex.Unlock()
	ref, found := refmap.Info[v]
	if !found {
		return
	}
	delete(refmap.Lookup, ref.s)
	delete(refmap.Info, v)
}

func (refmap *Map) Get(s string) uint64 {
	refmap.Mutex.RLock()
	defer refmap.Mutex.RUnlock()
	if ref, found := refmap.Lookup[s]; found {
		return ref
	}
	return Zero
}

func (refmap *Map) Incref(s string, incref int) uint64 {
	refmap.Mutex.Lock()
	defer refmap.Mutex.Unlock()
	ref, found := refmap.Lookup[s]
	if found {
		refmap.Info[ref].v += uint64(incref)
		return ref
	}
	refmap.val += 1
	ref = refmap.val
	refmap.Lookup[s] = ref
	refmap.Info[ref] = &Ref{s: s, v: uint64(incref)}
	return ref
}

func (refmap *Map) Decref(s string, decref int) {
	refmap.Mutex.Lock()
	defer refmap.Mutex.Unlock()
	ref, found := refmap.Lookup[s]
	if !found {
		return
	}
	i := refmap.Info[ref]
	v := i.v - uint64(decref)
	if v <= 0 {
		delete(refmap.Lookup, i.s)
		delete(refmap.Info, ref)
	}
}

func (refmap *Map) MultiGet(xs ...string) []uint64 {
	resp := make([]uint64, len(xs))
	refmap.Mutex.RLock()
	defer refmap.Mutex.RUnlock()
	for idx, s := range xs {
		if ref, found := refmap.Lookup[s]; found {
			resp[idx] = ref
		} else {
			resp[idx] = Zero
		}
	}
	return resp
}

func (refmap *Map) ReverseLookup(id uint64) (s string) {
	refmap.Mutex.RLock()
	defer refmap.Mutex.RUnlock()
	if ref, found := refmap.Info[id]; found {
		return ref.s
	}
	return
}

func NewWithVal(start uint64) *Map {
	info := make(map[uint64]*Ref)
	lookup := make(map[string]uint64)
	refmap := &Map{Info: info, Lookup: lookup, val: start}
	return refmap
}

func New() *Map {
	return NewWithVal(Zero)
}
