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

	t.Logf("%v", data)

	// if len(root) != 8 {
	// 	t.Errorf("Got an invalid number of results back from the parsed test2.yaml file: %d", len(root))
	// 	return
	// }

	admins, ok := data.Get("admins")
	if !ok {
		t.Error("Wasn't able to find a value for the 'admins' key in the parsed test2.yaml file.")
	}

	if admins.Type != List && len(admins.List) != 3 {
		t.Error("Got an invalid value for the 'admins' key in the parsed test2.yaml file.")
		return
	}

	t.Logf("admins[0] == %s\n", admins.List[0].String)

	if adminList, ok := data.GetStringList("admins"); ok {
		t.Logf("admins == %s\n", adminList)
	}

	if value, _ := data.GetString("application"); value != "espra" {
		t.Errorf("Got invalid value %q for the 'application' key in the parsed test2.yaml file.", value)
		return
	}

}
