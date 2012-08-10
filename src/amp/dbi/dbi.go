// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package dbi

type AllocRequest struct {
	Amount int
	Kind   string
	Parent *Key
}

type Allocation struct {
	Low  int64
	High int64
}

type Entity []Property

type EntityList struct {
	List []Entity
}

type Filter struct {
	Constraint string
	Value      interface{}
}

type Key struct {
	Kind     string
	StringID string
	IntID    int64
	Parent   *Key
}

type KeyList struct {
	List []*Key
}

type KeyValue struct {
	Key   *Key
	Value Entity
}

type Property struct {
	Name     string
	Value    interface{}
	NoIndex  bool
	Multiple bool
}

type Query struct {
	Filters  []Filter
	Ancestor *Key
	Order    string
	KeysOnly bool
	Limit    int
	Start    string
	End      string
}
