// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package mockfs mocks interactions with the filesystem.
package mockfs

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"dappui.com/pkg/sys"
)

var (
	errCloseFailure = errors.New("mockfs: failed to close file")
	errOpenFailure  = errors.New("mockfs: failed to open file")
	errReadFailure  = errors.New("mockfs: failed to read file")
	errStatFailure  = errors.New("mockfs: failed to stat file")
)

// File provides a mock implementation of the sys.File interface.
type File struct {
	closed    bool
	data      *bytes.Buffer
	failClose bool
	failRead  bool
	mu        sync.Mutex // protects closed
}

// Close implements the interface for sys.File.
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return os.ErrClosed
	}
	if f.failClose {
		return errCloseFailure
	}
	f.closed = true
	return nil
}

// Read implements the interface for sys.File.
func (f *File) Read(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, os.ErrClosed
	}
	if f.failRead {
		return 0, errReadFailure
	}
	return f.data.Read(p)
}

// FileInfo provides a mock implementation of the os.FileInfo interface.
type FileInfo struct {
	data      string
	dir       bool
	failClose bool
	failOpen  bool
	failRead  bool
	failStat  bool
	name      string
}

// FailClose will mark a file to fail when its Close method is called.
func (f *FileInfo) FailClose() *FileInfo {
	f.failClose = true
	return f
}

// FailOpen will return an error for Open calls at the current path.
func (f *FileInfo) FailOpen() *FileInfo {
	f.failOpen = true
	return f
}

// FailRead will mark a file to fail when its Read method is called.
func (f *FileInfo) FailRead() *FileInfo {
	f.failRead = true
	return f
}

// FailStat will return an error for Lstat/Stat calls at the current path.
func (f *FileInfo) FailStat() *FileInfo {
	f.failStat = true
	return f
}

// IsDir implements the interface for os.FileInfo.
func (f *FileInfo) IsDir() bool {
	return f.dir
}

// ModTime implements the interface for os.FileInfo.
func (f *FileInfo) ModTime() time.Time {
	return time.Time{}
}

// Mode implements the interface for os.FileInfo.
func (f *FileInfo) Mode() os.FileMode {
	if f.dir {
		return os.ModeDir
	}
	return 0
}

// Name implements the interface for os.FileInfo.
func (f *FileInfo) Name() string {
	return f.name
}

// Size implements the interface for os.FileInfo.
func (f *FileInfo) Size() int64 {
	return int64(len(f.data))
}

// Sys implements the interface for os.FileInfo.
func (f *FileInfo) Sys() interface{} {
	return f.name
}

// FileSystem provides a configurable mock implementation of the sys.FileSystem
// interface.
type FileSystem struct {
	files map[string]*FileInfo
	mu    sync.RWMutex // protects files
}

// Lstat implements the interface for sys.Filesystem.
func (f *FileSystem) Lstat(path string) (os.FileInfo, error) {
	return f.Stat(path)
}

// Mkdir creates a directory at the given path, along with any necessary parent
// directories. If a directory already exists, then the method will exit early.
//
// WriteFile implicitly creates directories, so this method is only really
// useful for creating empty directories.
func (f *FileSystem) Mkdir(path string) *FileInfo {
	f.mu.Lock()
	defer f.mu.Unlock()
	path = filepath.Clean("/" + path)
	if info, ok := f.files[path]; ok {
		return info
	}
	dir, name := filepath.Split(path)
	info := &FileInfo{
		dir:  true,
		name: name,
	}
	f.files[path] = info
	path = dir
	for path != "/" {
		path = filepath.Clean(path)
		if _, ok := f.files[path]; ok {
			break
		}
		dir, name := filepath.Split(path)
		f.files[path] = &FileInfo{
			dir:  true,
			name: name,
		}
		path = dir
	}
	return info
}

// Open implements the interface for sys.Filesystem.
func (f *FileSystem) Open(path string) (sys.File, error) {
	path = filepath.Clean("/" + path)
	f.mu.RLock()
	info, ok := f.files[path]
	f.mu.RUnlock()
	if !ok {
		return nil, os.ErrNotExist
	}
	if info.failOpen {
		return nil, errOpenFailure
	}
	if info.failStat {
		return nil, errStatFailure
	}
	return &File{
		data:      bytes.NewBufferString(info.data),
		failClose: info.failClose,
		failRead:  info.failRead,
	}, nil
}

// Stat implements the interface for sys.Filesystem.
func (f *FileSystem) Stat(path string) (os.FileInfo, error) {
	path = filepath.Clean("/" + path)
	f.mu.RLock()
	info, ok := f.files[path]
	f.mu.RUnlock()
	if !ok {
		return nil, os.ErrNotExist
	}
	if info.failStat {
		return nil, errStatFailure
	}
	return info, nil
}

// Walk implements the interface for sys.Filesystem.
func (f *FileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	paths := []string{}
	root = filepath.Clean("/" + root)
	dir := root + "/"
	if root == "/" {
		dir = "/"
	}
	for path := range f.files {
		if path == root || strings.HasPrefix(path, dir) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	var err error
	for _, path := range paths {
		info := f.files[path]
		if info.failStat {
			err = errStatFailure
		}
		// TODO(tav): Add support for filepath.SkipDir.
		if err := walkFn(path, info, err); err != nil {
			return err
		}
		err = nil
	}
	return nil
}

// WriteFile creates a file with the given data at the specified path. It will
// implicitly create any parent directories as necessary.
func (f *FileSystem) WriteFile(path string, data string) *FileInfo {
	f.mu.Lock()
	path = filepath.Clean("/" + path)
	dir, name := filepath.Split(path)
	info := &FileInfo{
		data: data,
		name: name,
	}
	f.files[path] = info
	f.mu.Unlock()
	if dir != "/" {
		f.Mkdir(dir)
	}
	return info
}

// New returns a mockable filesystem for testing purposes.
func New() *FileSystem {
	return &FileSystem{
		files: map[string]*FileInfo{
			"/": {
				dir: true,
			},
		},
	}
}
