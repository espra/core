// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package trust

import (
	"amp/log"
	"testing"
)

func TestCompute(t *testing.T) {
	trustmap := &TrustMap{
		Mapping: map[string][]string{
			"tav":     []string{"micrypt", "thruflo"},
			"micrypt": []string{"h4rrydog"},
		},
		Root: "tav",
	}
	trustmap.Compute()
	log.Wait()
}

func init() {
	log.AddConsoleLogger()
}
