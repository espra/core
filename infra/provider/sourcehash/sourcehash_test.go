// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package sourcehash

import (
	"fmt"
	"testing"

	"dappui.com/pkg/mock/sys"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResource(t *testing.T) {
	r := Resource()
	if err := r.InternalValidate(nil, false); err != nil {
		t.Fatalf("unable to validate the resource with a nil schema map: %s", err)
	}
	digest := "7a4d64e9271b460aa92f75a4726bba165784153d32fbc9268dd63d64d20d02d0"
	testRead(t, r, digest, "test")
	testRead(t, r, digest, "test", "test/a")
	testReadFailure(t, r, func(fs *sys.FileSystem) {
		fs.Mkdir("test").FailStat()
	})
	testReadFailure(t, r, func(fs *sys.FileSystem) {
		fs.WriteFile("test/x", "data").FailClose()
	})
	testReadFailure(t, r, func(fs *sys.FileSystem) {
		fs.WriteFile("test/x", "data").FailOpen()
	})
	testReadFailure(t, r, func(fs *sys.FileSystem) {
		fs.WriteFile("test/x", "data").FailRead()
	})
	testReadFailure(t, r, func(fs *sys.FileSystem) {
		fs.WriteFile("test/x", "data").FailStat()
	})
}

func testRead(t *testing.T, r *schema.Resource, expected string, paths ...string) {
	data := toTerraformData(r, paths)
	if err := read(data, nil); err != nil {
		t.Errorf("unable to evaluate digest for test path: %s", err)
		return
	}
	digest := data.Get("digest").(string)
	if digest == "" {
		t.Error("failed to calculate digest for test path")
		return
	}
	if digest != expected {
		t.Errorf("got invalid digest for test path: %q", digest)
		return
	}
}

func testReadFailure(t *testing.T, r *schema.Resource, init func(*sys.FileSystem)) {
	mock := sys.NewFileSystem()
	fs = mock
	mock.WriteFile("test/a", "some data")
	init(mock)
	data := toTerraformData(r, []string{"test"})
	if err := read(data, nil); err == nil {
		t.Errorf("successfully read resource that should have failed")
	}
}

func toTerraformData(r *schema.Resource, paths []string) *schema.ResourceData {
	attrs := map[string]string{}
	attrs["paths.#"] = fmt.Sprintf("%d", len(paths))
	for idx, path := range paths {
		attrs[fmt.Sprintf("paths.%d", idx)] = path
	}
	return r.Data(&terraform.InstanceState{
		Attributes: attrs,
	})
}
