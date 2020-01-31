// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package sys

import (
	"os"
	"testing"
)

func TestOSFileSystem(t *testing.T) {
	fs := OSFileSystem
	if _, err := fs.Lstat("sys.go"); err != nil {
		t.Fatalf("failed to lstat path: %s", err)
	}
	f, err := fs.Open("sys.go")
	if err != nil {
		t.Fatalf("failed to open path: %s", err)
	}
	f.Close()
	if _, err := fs.Stat("sys.go"); err != nil {
		t.Fatalf("failed to stat path: %s", err)
	}
	if err := fs.Walk(".", func(path string, info os.FileInfo, err error) error {
		return nil
	}); err != nil {
		t.Fatalf("failed to walk path: %s", err)
	}
}
