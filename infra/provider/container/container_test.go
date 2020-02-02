// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package container

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResource(t *testing.T) {
	r := Resource()
	if err := r.InternalValidate(nil, true); err != nil {
		t.Fatalf("unable to validate the resource with a nil schema map: %s", err)
	}
	data := r.Data(&terraform.InstanceState{
		Attributes: map[string]string{
			"repo":   "docker.pkg.github.com/dappui/core/container-test",
			"source": "testdata/success",
			"tag":    "latest",
		},
	})
	if err := r.Create(data, nil); err != nil {
		t.Fatalf("unable to create resource: %s", err)
	}
	if err := r.Update(data, nil); err != nil {
		t.Fatalf("unable to update resource: %s", err)
	}
	data = r.Data(&terraform.InstanceState{
		Attributes: map[string]string{
			"repo":   "docker.pkg.github.com/dappui/core/container-test",
			"source": "testdata/failure",
			"tag":    "latest",
		},
	})
	if err := r.Create(data, nil); err == nil {
		t.Fatalf("unexpected success when creating an invalid resource")
	}
	data = r.Data(&terraform.InstanceState{
		Attributes: map[string]string{
			"repo":   "docker.pkg.github.com/site/invalid/invalid",
			"source": "testdata/success",
			"tag":    "latest",
		},
	})
	if err := r.Create(data, nil); err == nil {
		t.Fatalf("unexpected success when creating an invalid resource")
	}
}
