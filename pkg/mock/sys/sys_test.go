// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package sys

import (
	"io"
	"os"
	"testing"
)

func TestFS(t *testing.T) {
	fs := NewFileSystem()
	path := "/path/to/test"
	fs.WriteFile(path, "hello world")
	info, err := fs.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %s", err)
	}
	if info.IsDir() {
		t.Fatal("unexpected reporting of file as directory in file info")
	}
	if !info.ModTime().IsZero() {
		t.Fatal("unexpected modtime value in file info")
	}
	if !info.Mode().IsRegular() {
		t.Fatal("unexpected reporting of file mode in file info")
	}
	if info.Name() != "test" {
		t.Fatalf("file info contained the wrong name: got %q, want %q", info.Name(), "test")
	}
	if info.Size() != 11 {
		t.Fatalf("file info contained the wrong size: got %d, want 11", info.Size())
	}
	if info.Sys() == nil {
		t.Fatal("unexpected sys value in file info")
	}
	f, err := fs.Open(path)
	if err != nil {
		t.Fatalf("failed to open file: %s", err)
	}
	buf := make([]byte, 11)
	if _, err := io.ReadFull(f, buf); err != nil {
		t.Fatalf("failed to read file: %s", err)
	}
	if string(buf) != "hello world" {
		t.Fatalf("got wrong contents when reading file: got %q; want %q", string(buf), "hello world")
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close file: %s", err)
	}
	if err := f.Close(); err == nil {
		t.Fatal("successfully closed already closed file")
	}
	if _, err := io.ReadFull(f, buf); err == nil {
		t.Fatal("successfully read already closed file")
	}
	path = "/some/other/path"
	fs.WriteFile(path, "hello world").FailRead().FailClose()
	f, err = fs.Open(path)
	if err != nil {
		t.Fatalf("failed to open file: %s", err)
	}
	if _, err := io.ReadFull(f, buf); err == nil {
		t.Fatal("successfully read file marked to FailRead")
	}
	if err := f.Close(); err == nil {
		t.Fatal("successfully closed file marked to FailClose")
	}
	path = "/fail/to/open"
	fs.WriteFile(path, "data").FailOpen()
	if f, err = fs.Open(path); err == nil {
		t.Fatal("successfully opened file marked to FailOpen")
	}
	info, err = fs.Stat("/some/other")
	if err != nil {
		t.Fatalf("failed to stat directory: %s", err)
	}
	if !info.IsDir() || !info.Mode().IsDir() {
		t.Fatal("info does not report path as directory")
	}
	info = fs.Mkdir("/some/other")
	if !info.IsDir() || !info.Mode().IsDir() {
		t.Fatal("info does not report path as directory")
	}
	info = fs.Mkdir("/some/other/subpath")
	if !info.IsDir() || !info.Mode().IsDir() {
		t.Fatal("info does not report path as directory")
	}
	path = "/does/not/exist"
	if _, err := fs.Lstat(path); err == nil {
		t.Fatal("successfully did lstat on non-existent file")
	}
	if _, err := fs.Open(path); err == nil {
		t.Fatal("successfully opened non-existent file")
	}
	if err := fs.Walk("/", func(path string, info os.FileInfo, err error) error {
		return err
	}); err != nil {
		t.Fatalf("failed to walk path: %s", err)
	}
	path = "/broken/file"
	fs.WriteFile(path, "some data").FailStat()
	if _, err := fs.Open(path); err == nil {
		t.Fatal("successfully opened path marked to FailStat")
	}
	if _, err := fs.Stat(path); err == nil {
		t.Fatal("successfully did stat on path marked to FailStat")
	}
	if err := fs.Walk("/", func(path string, info os.FileInfo, err error) error {
		return err
	}); err == nil {
		t.Fatal("successfully walked path marked to FailStat")
	}
}
