// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package trust

import (
	"amp/log"
)

type TrustMap struct {
	Mapping map[string][]string
	Root    string
}

type TrustValue struct {
	Value      float64
	Identifier string
}

type Computation []*TrustValue

func (trustmap *TrustMap) Compute() (result Computation) {
	log.Info("Computing ...")
	return
}
