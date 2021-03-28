// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// +build darwin dragonfly freebsd netbsd openbsd

package term

import (
	"golang.org/x/sys/unix"
)

func getTermios(fd int) (*unix.Termios, error) {
	return unix.IoctlGetTermios(fd, unix.TIOCGETA)
}

func setTermios(fd int, termios *unix.Termios) error {
	return unix.IoctlSetTermios(fd, unix.TIOCSETA, termios)
}
