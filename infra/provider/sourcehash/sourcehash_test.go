// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package sourcehash

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestRead(t *testing.T) {
	resource := &schema.Resource{
		Schema: Resource().Schema,
	}
	data := resource.Data(&terraform.InstanceState{
		Attributes: map[string]string{
			"paths.#": "1",
			"paths.0": "test",
		},
	})
	if err := read(data, nil); err != nil {
		t.Fatalf("unable to evaluate digest for test path: %s", err)
	}
	digest := data.Get("digest").(string)
	if digest == "" {
		t.Fatal("failed to calculate digest for test path")
	}
	if digest != "7a4d64e9271b460aa92f75a4726bba165784153d32fbc9268dd63d64d20d02d0" {
		t.Fatalf("got invalid digest for test path: %q", digest)
	}
}

func TestResource(t *testing.T) {
	if err := Resource().InternalValidate(nil, false); err != nil {
		t.Fatalf("unable to validate the resource with a nil schema map: %s", err)
	}
}
