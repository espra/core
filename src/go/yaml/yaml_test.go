// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package yaml

import (
	"testing"
)

func TestYAMLFile(t *testing.T) {

	data, err := ParseFile("test.yaml")

	if err != nil {
		t.Errorf("Got an unexpected error reading the test.yaml file: %s", err)
	}

	if len(data) != 4 {
		t.Errorf("Got an invalid number of results back from the parsed test.yaml file: %d", len(data))
		return
	}

	if data["application"] != "espra" {
		t.Error("Got an invalid value for the 'application' key in the parsed test.yaml file.")
		return
	}
}
