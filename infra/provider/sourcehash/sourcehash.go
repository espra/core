// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package sourcehash defines a datasource for hashing source paths.
//
// The idea for this datasource is inspired by Dragan Milic's approach in:
// https://github.com/draganm/terraform-provider-linuxbox
package sourcehash

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var newline = []byte{'\n'}

// Resource returns the schema definition for the sourcehash datasource.
func Resource() *schema.Resource {
	return &schema.Resource{
		Read: read,
		Schema: map[string]*schema.Schema{
			"digest": &schema.Schema{
				Computed: true,
				Type:     schema.TypeString,
			},
			"paths": &schema.Schema{
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				Type:     schema.TypeSet,
			},
		},
	}
}

func read(d *schema.ResourceData, meta interface{}) error {
	id := sha512.New512_256()
	seen := map[string]struct{}{}
	for _, elem := range d.Get("paths").(*schema.Set).List() {
		path := elem.(string)
		id.Write([]byte(path))
		id.Write(newline)
		info, err := os.Lstat(path)
		if err != nil {
			return fmt.Errorf("sourcehash: failed to stat path %q: %s", path, err)
		}
		if !info.IsDir() {
			seen[path] = struct{}{}
			continue
		}
		if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				seen[path] = struct{}{}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("sourcehash: failed to navigate path %q: %s", path, err)
		}
	}
	files := []string{}
	for file := range seen {
		files = append(files, file)
	}
	sort.Strings(files)
	buf := make([]byte, 65536)
	hasher := sha512.New512_256()
	for _, file := range files {
		hasher.Write([]byte(file))
		hasher.Write(newline)
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("sourcehash: failed to open file %q: %s", file, err)
		}
		_, err = io.CopyBuffer(hasher, f, buf)
		if err != nil {
			f.Close()
			return fmt.Errorf("sourcehash: failed to hash file %q: %s", file, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("sourcehash: failed to close file %q: %s", file, err)
		}
		hasher.Write(newline)
	}
	d.Set("digest", hex.EncodeToString(hasher.Sum(nil)))
	d.SetId(hex.EncodeToString(id.Sum(nil)))
	return nil
}
