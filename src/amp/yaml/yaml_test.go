// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

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
