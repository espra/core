// Public Domain (-) 2021-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package wsl

import (
	"bytes"
	"os"
)

var readFile = os.ReadFile

func detect() bool {
	// Supposedly, /proc/sys/kernel/osrelease will be of the form:
	//
	//   <major>.<minor>.<patch>-microsoft-WSL<version>-<flavour>
	//
	// For example: 4.19.112-microsoft-WSL2-standard
	//
	// Source:
	// https://github.com/microsoft/WSL/issues/423#issuecomment-611086412
	os, err := readFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	if bytes.Contains(os, []byte("WSL")) {
		return true
	}
	// TODO(tav): Remove this check once the kernel/osrelease change has been
	// backported, as Microsoft may create a popular non-WSL Linux distro at
	// some point.
	return bytes.Contains(bytes.ToLower(os), []byte("microsoft"))

}
