// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package amp

import (
	"json"
	"testing"
)

type city string

type address struct {
	Street  string
	Town    string
	Country string
}

type metaStruct struct {
	Address  *address
	City     city
	Duration float64
	Partner  string
	Stylists []string
	Month    *string
	Sexy     string
}

func TestHeader(t *testing.T) {

	header := Header{
		"name":       "tav",
		"age":        float64(29),
		"housemates": []string{"Dave", "Mamading", "Sofia", "Sylvana"},
		"meta": map[string]interface{}{
			"City":     "London",
			"Duration": 6,
			"Month":    "March",
			"Address": map[string]interface{}{
				"Town":    "Brixton",
				"Country": "U.K.",
			},
			"Sexy":     "jeffarch",
			"Stylists": []string{"Phillipe", "Diane"},
			"Birth":    float64(1982),
		},
	}

	name, ok := header.GetString("name")
	if !ok {
		t.Errorf("Couldn't get the header value for 'name'.")
		return
	}

	if name != "tav" {
		t.Errorf("Go unexpected value for the 'name' header: %q", name)
		return
	}

	var housemates []string

	err := header.Get("housemates", &housemates)
	if err != nil {
		t.Errorf("Got an error getting the header value for 'housemates': %s", err)
		return
	}

	if len(housemates) != 4 {
		t.Errorf("Got an invalid value for the housemates header: %#v", housemates)
		return
	}

	var age float64

	header.Get("age", &age)

	if age != 29 {
		t.Errorf("Got an invalid value for the age header: %f", age)
		return
	}

	var meta metaStruct

	err = header.Get("meta", &meta)
	if err != nil {
		t.Errorf("Got an error getting the header value for 'meta': %s", err)
		return
	}

	if meta.City != "London" {
		t.Errorf("Got an invalid value for meta.City: %v", meta.City)
		return
	}

	if *meta.Month != "March" {
		t.Errorf("Got an invalid value for *meta.Month: %v", *meta.Month)
		return
	}

}

func TestHeaderFromJSON(t *testing.T) {

	header := &Header{}

	jsonerr := json.Unmarshal([]byte(`{
		"name":       "tav",
		"age":        29,
		"housemates": ["Dave", "Mamading", "Sofia", "Sylvana"],
		"meta": {
			"City":     "London",
			"Duration": 6,
			"Month":    "March",
			"Address": {
				"Town":    "Brixton",
				"Country": "U.K."
			},
			"Sexy":     "jeffarch",
			"Stylists": ["Phillipe", "Diane"],
			"Birth":    1982
		}
	}`), header)

	if jsonerr != nil {
		t.Errorf("Error parsing JSON data as a header: %s", jsonerr)
		return
	}

	name, ok := header.GetString("name")
	if !ok {
		t.Errorf("Couldn't get the header value for 'name'.")
		return
	}

	if name != "tav" {
		t.Errorf("Go unexpected value for the 'name' header: %q", name)
		return
	}

	var housemates []string

	err := header.Get("housemates", &housemates)
	if err != nil {
		t.Errorf("Got an error getting the header value for 'housemates': %s", err)
		return
	}

	if len(housemates) != 4 {
		t.Errorf("Got an invalid value for the housemates header: %#v", housemates)
		return
	}

	var age float64

	header.Get("age", &age)

	if age != 29 {
		t.Errorf("Got an invalid value for the age header: %d", age)
		return
	}

	var meta metaStruct

	err = header.Get("meta", &meta)
	if err != nil {
		t.Errorf("Got an error getting the header value for 'meta': %s", err)
		return
	}

	if meta.City != "London" {
		t.Errorf("Got an invalid value for meta.City: %v", meta.City)
		return
	}

	if *meta.Month != "March" {
		t.Errorf("Got an invalid value for *meta.Month: %v", *meta.Month)
	}

}
