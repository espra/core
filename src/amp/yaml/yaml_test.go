// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package yaml

import (
	"testing"
)

func TestParseDictFile(t *testing.T) {

	data, err := ParseDictFile("test.yaml")

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

func TestParseFile(t *testing.T) {

	data, err := ParseFile("test2.yaml")

	if err != nil {
		t.Errorf("Got an unexpected error reading the test2.yaml file: %s", err)
	}

	if len(data) != 8 {
		t.Errorf("Got an invalid number of results back from the parsed test2.yaml file: %d", len(data))
		return
	}

	t.Logf(Display(data))

	if data["admins"].Type != List && len(data["admins"].List) != 3 {
		t.Error("Got an invalid value for the 'admins' key in the parsed test2.yaml file.")
		return
	}

	t.Logf("admins[0] == %s\n", data["admins"].List[0].String)

}
