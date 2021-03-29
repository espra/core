// Public Domain (-) 2021-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// Package wsl provides support for Windows Subsystem for Linux (WSL).
package wsl

// Detect returns whether the program looks like it's running under WSL.
func Detect() bool {
	return detect()
}
