// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package term

import (
	"golang.org/x/sys/unix"
)

func getTermios(fd int) (*unix.Termios, error) {
	return unix.IoctlGetTermios(fd, unix.TCGETS)
}

func setTermios(fd int, termios *unix.Termios) error {
	return unix.IoctlSetTermios(fd, unix.TCSETS, termios)
}
