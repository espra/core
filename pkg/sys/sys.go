// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package sys abstracts interactions with the base operating system.
package sys

import (
	"io"
	"os"
	"path/filepath"
)

// OSFileSystem implements the FileSystem interface for the os package.
var OSFileSystem FileSystem = osFS{}

// File defines an abstract interface for os.File-like implementations.
type File interface {
	io.Closer
	io.Reader
}

// FileSystem defines an abstract interface for interacting with concrete and
// virtual filesystems.
type FileSystem interface {
	Lstat(path string) (os.FileInfo, error)
	Open(path string) (File, error)
	Stat(path string) (os.FileInfo, error)
	Walk(root string, walkFn filepath.WalkFunc) error
}

type osFS struct{}

func (o osFS) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (o osFS) Open(path string) (File, error) {
	return os.Open(path)
}

func (o osFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (o osFS) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}
