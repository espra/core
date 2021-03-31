// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// +build darwin dragonfly freebsd linux netbsd openbsd

package term

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

type device struct {
	fd      int
	termios *unix.Termios
}

func (d *device) Read(p []byte) (int, error) {
	return unix.Read(d.fd, p)
}

func disableEcho(d *device) error {
	t := *d.termios
	t.Iflag |= unix.ICRNL // Enable CR -> NL translation
	t.Lflag &^= unix.ECHO // Disable echoing
	t.Lflag |= 0 |
		unix.ICANON | // Enable canonical/cooked input processing
		unix.ISIG // Enable signal generation for characters like DSUSP, INTR, QUIT, and SUSP
	return setTermios(d.fd, &t)
}

func getDevice(fd int) (*device, error) {
	t, err := getTermios(fd)
	return &device{fd, t}, err
}

func isTTY(fd int) bool {
	_, err := getTermios(fd)
	return err == nil
}

// This function behaves like `cfmakeraw` on various platforms:
//
// * FreeBSD:
// https://github.com/freebsd/freebsd-src/blob/master/lib/libc/gen/termios.c
//
// * Linux/glibc:
// https://sourceware.org/git/?p=glibc.git;a=blob;f=termios/cfmakeraw.c
//
// * OpenBSD:
// https://github.com/openbsd/src/blob/master/lib/libc/termios/cfmakeraw.c
//
// As well as like raw mode on other systems, e.g.
//
// * OpenSSH: https://github.com/openssh/openssh-portable/blob/master/sshtty.c
func makeRaw(d *device, mode *RawMode) error {
	// NOTE(tav): Given the historic nature of all of this, you are likely to
	// find better documentation from early UNIX systems than from more "modern"
	// systems.
	t := *d.termios
	// Configure input processing.
	t.Iflag &^= 0 |
		unix.BRKINT | // Ignore break conditions (in conjunction with IGNBRK)
		unix.IGNCR | // Process CR
		unix.IGNPAR | // Pass through bytes with framing/parity errors
		unix.INLCR | // Disable NL -> CR translation
		unix.INPCK | // Disable input parity checking
		unix.ISTRIP | // Disable stripping of the high bit in 8-bit characters
		unix.IXOFF | // Disable use of START and STOP characters for control flow on input
		unix.IXON | // Disable use of START and STOP characters for control flow on output
		unix.PARMRK // Do not mark framing/parity errors
	t.Iflag |= unix.IGNBRK // Ignore break conditions
	if mode != nil && mode.DisableCRNL {
		t.Iflag &^= unix.ICRNL // Disable CR -> NL translation
		t.Oflag &^= unix.OPOST // Disable output post-processing
	} else {
		t.Iflag |= unix.ICRNL  // Enable CR -> NL translation
		t.Oflag = unix.OPOST | // Enable output post-processing
			unix.ONLCR // Enable NL -> CRNL translation
	}
	// Configure local terminal functions.
	t.Lflag &^= 0 |
		unix.ECHO | // Disable echoing
		unix.ECHOE | // Disable echoing erasure of input by the ERASE character
		unix.ECHOK | // Disable echoing of NL after the KILL character
		unix.ECHONL | // Disable echoing of NL
		unix.ICANON | // Disable canonical/cooked input processing
		unix.IEXTEN | // Disable extended input processing like DISCARD and LNEXT
		unix.ISIG // Disable signal generation for characters like DSUSP, INTR, QUIT, and SUSP
	// Configure control modes.
	t.Cflag &^= 0 |
		unix.CSIZE | // Clear the current character size mask
		unix.PARENB // Disable parity checking
	t.Cflag |= 0 |
		unix.CREAD | // Enable receiving of characters
		unix.CS8 // Specify 8-bit character sizes
	// Set the minimum number of bytes for read calls.
	if mode != nil && mode.NonBlocking {
		t.Cc[unix.VMIN] = 0
	} else {
		t.Cc[unix.VMIN] = 1
	}
	// Disable timeouts on data transmissions.
	t.Cc[unix.VTIME] = 0
	return setTermios(d.fd, &t)
}

func setDevice(d *device) error {
	return setTermios(d.fd, d.termios)
}

func watchResize(ctx context.Context, fd int, ch chan Dimensions) {
	c := make(chan os.Signal, 100)
	signal.Notify(c, syscall.SIGWINCH)
	for {
		select {
		case <-ctx.Done():
			signal.Stop(c)
			close(ch)
			return
		case <-c:
			dim, err := windowSize(fd)
			if err != nil {
				continue
			}
			ch <- dim
		}
	}
}

func windowSize(fd int) (Dimensions, error) {
	w, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return Dimensions{}, err
	}
	return Dimensions{
		Cols: int(w.Col),
		Rows: int(w.Row),
	}, nil
}
