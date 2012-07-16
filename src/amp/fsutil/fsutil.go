// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package fsutil

import (
	"fmt"
	"os"
)

type NotFound struct {
	path string
}

func (err *NotFound) Error() string {
	return fmt.Sprintf("not found: %s", err.path)
}

type NotFile struct {
	path string
}

func (err *NotFile) Error() string {
	return fmt.Sprintf("directory found instead of file at: %s", err.path)
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, &NotFound{path}
	}
	return false, err
}

func FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return true, nil
		}
		return false, &NotFile{path}
	}
	if os.IsNotExist(err) {
		return false, &NotFound{path}
	}
	return false, err
}