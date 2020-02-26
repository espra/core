// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package container defines a resource for creating containers.
//
// The specified container will be built using docker and pushed to the repo.
package container

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Resource returns the schema definition for the container resource.
func Resource() *schema.Resource {
	return &schema.Resource{
		Create: create,
		Delete: schema.Noop,
		Read:   schema.Noop,
		Schema: map[string]*schema.Schema{
			"image": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"repo": {
				Required: true,
				Type:     schema.TypeString,
			},
			"source": {
				Required: true,
				Type:     schema.TypeString,
			},
			"tag": {
				Required: true,
				Type:     schema.TypeString,
			},
		},
		Update: update,
	}
}

func build(repo string, d *schema.ResourceData, meta interface{}) error {
	source := d.Get("source").(string)
	tag := d.Get("tag").(string)
	image := repo + ":" + tag
	cmd := exec.Command("docker", "build", "-t", image, ".")
	cmd.Dir = source
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("container: failed to build the %q image:\n\n%s", image, string(out))
	}
	cmd = exec.Command("docker", "push", image)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("container: failed to push the %q image:\n\n%s", image, string(out))
	}
	d.Set("image", image)
	return nil
}

func create(d *schema.ResourceData, meta interface{}) error {
	repo := d.Get("repo").(string)
	if err := build(repo, d, nil); err != nil {
		return err
	}
	d.SetId(repo)
	return nil
}

func update(d *schema.ResourceData, meta interface{}) error {
	repo := d.Get("repo").(string)
	return build(repo, d, nil)
}
