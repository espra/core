// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// Package term provides support for interacting with terminals.
package term

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"web4.cc/pkg/process"
)

// Control characters.
const (
	KeyNull InputKey = iota
	KeyCtrlA
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlG
	KeyCtrlH
	KeyCtrlI
	KeyCtrlJ
	KeyCtrlK
	KeyCtrlL
	KeyCtrlM
	KeyCtrlN
	KeyCtrlO
	KeyCtrlP
	KeyCtrlQ
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlV
	KeyCtrlW
	KeyCtrlX
	KeyCtrlY
	KeyCtrlZ
	KeyCtrlLeftBracket
	KeyCtrlBackslash
	KeyCtrlRightBracket
	KeyCtrlCaret
	KeyCtrlUnderscore
)

// Common aliases for control characters.
const (
	KeyBackspace InputKey = 127
	KeyEnter     InputKey = '\n'
	KeyEscape    InputKey = 27
	KeyInterrupt InputKey = KeyCtrlC
	KeyTab       InputKey = '\t'
)

// Arrow keys.
const (
	KeyDown InputKey = iota + 128
	KeyEnd
	KeyHome
	KeyLeft
	KeyPageDown
	KeyPageUp
	KeyRight
	KeyUp
)

const isWindows = runtime.GOOS == "windows"

// Error values.
var (
	errInvalidResponse = errors.New("term: invalid response from terminal")
)

var (
	cursorHidden bool
	mu           sync.Mutex // protects cursorHidden
)

// Device represents a device file that has been "converted" using `MakeRaw`.
//
// Devices must be reset by calling the `Reset` method before the program exits.
// Otherwise, external systems like the user's shell might be left in a broken
// state.
type Device struct {
	d *device
}

// Read does a raw read on the device.
func (d *Device) Read(p []byte) (int, error) {
	return d.d.Read(p)
}

// Reset resets the device file back to its initial state before it was
// converted.
func (d *Device) Reset() error {
	return setDevice(d.d)
}

// Dimensions represents the window dimensions for the terminal.
type Dimensions struct {
	Cols int
	Rows int
}

// Input represents input received from the terminal. The `Byte` field
// represents non-control characters. When the `Byte` field is `0`, then the
// `Key` field represents a control character or arrow key.
//
// Note that both `\n` and `\r` are mapped to `KeyEnter` for simplicity, and
// that the `\n` and `\t` characters are returned in the `Key` field, while the
// space character `' '` is returned in the `Byte` field.
type Input struct {
	Byte byte
	Key  InputKey
}

// InputKey represents a special input key received from the terminal. This
// encompasses both control characters and arrow keys.
type InputKey int

// Pos represents the cursor position on the terminal.
type Pos struct {
	Col int
	Row int
}

// RawConfig specifies the configuration for `MakeRaw`. It can be specified
// using one of the `RawOption` functions.
type RawConfig struct {
	block bool
	canon bool
	crnl  bool
	sig   bool
}

// RawOption functions configure `RawConfig` for `MakeRaw` calls.
type RawOption func(*RawConfig)

// Screen provides an interface for building interactive terminal applications.
//
// On UNIX systems, `Screen` reads and writes from `/dev/tty`. On Windows, it
// reads from `CONIN$` and writes to `CONOUT$`.
type Screen struct {
	dev     *Device
	in      *os.File
	intr    bool
	pending []byte
	out     *os.File
}

// Bell tells the terminal to emit a beep/bell.
func (s *Screen) Bell() {
	s.out.WriteString("\x07")
}

// ClearLine clears the current line.
func (s *Screen) ClearLine() {
	s.out.WriteString("\x1b[2K")
}

// ClearLineToEnd clears everything from the cursor to the end of the current
// line.
func (s *Screen) ClearLineToEnd() {
	s.out.WriteString("\x1b[0K")
}

// ClearLineToStart clears everything from the cursor to the start of the
// current line.
func (s *Screen) ClearLineToStart() {
	s.out.WriteString("\x1b[1K")
}

// ClearScreen clears the screen and moves the cursor to the top left.
func (s *Screen) ClearScreen() {
	s.out.WriteString("\x1b[2J\x1b[H")
}

// ClearToEnd clears everything from the cursor to the end of the screen.
func (s *Screen) ClearToEnd() {
	s.out.WriteString("\x1b[0J")
}

// ClearToStart clears everything from the cursor to the start of the screen.
func (s *Screen) ClearToStart() {
	s.out.WriteString("\x1b[1J")
}

// CursorDown moves the cursor down by the given amount.
func (s *Screen) CursorDown(n int) {
	fmt.Fprintf(s.out, "\x1b[%dB", n)
}

// CursorLeft moves the cursor left by the given amount.
func (s *Screen) CursorLeft(n int) {
	fmt.Fprintf(s.out, "\x1b[%dD", n)
}

// CursorPos returns the current position of the cursor.
func (s *Screen) CursorPos() (*Pos, error) {
	if err := s.makeRaw(); err != nil {
		return nil, err
	}
	defer s.reset()
	// Query the terminal.
	if _, err := s.out.WriteString("\x1b[6n"); err != nil {
		return nil, err
	}
	// Read the response.
	buf := [20]byte{}
	i := 0
	for i < len(buf) {
		n, err := s.Read(buf[i : i+1])
		if n == 1 && buf[i] == 'R' {
			break
		}
		if err != nil {
			return nil, err
		}
		i++
	}
	// Exit on invalid data from the device.
	if !(i >= 5 && buf[0] == '\x1b' && buf[1] == '[') {
		return nil, errInvalidResponse
	}
	// Parse the response to get the position.
	split := bytes.Split(buf[2:i], []byte{';'})
	if len(split) != 2 {
		return nil, errInvalidResponse
	}
	row, err := strconv.ParseUint(string(split[0]), 10, 16)
	if err != nil {
		return nil, errInvalidResponse
	}
	col, err := strconv.ParseUint(string(split[1]), 10, 16)
	if err != nil {
		return nil, errInvalidResponse
	}
	return &Pos{
		Col: int(col),
		Row: int(row),
	}, nil
}

// CursorRight moves the cursor right by the given amount.
func (s *Screen) CursorRight(n int) {
	fmt.Fprintf(s.out, "\x1b[%dC", n)
}

// CursorTo moves the cursor to the given position.
func (s *Screen) CursorTo(row, column int) {
	fmt.Fprintf(s.out, "\x1b[%d;%dH", row, column)
}

// CursorUp moves the cursor up by the given amount.
func (s *Screen) CursorUp(n int) {
	fmt.Fprintf(s.out, "\x1b[%dA", n)
}

// HideCursor hides the cursor on the terminal. It also registers a
// `process.Exit` handler to restore the cursor when the process exits.
func (s *Screen) HideCursor() {
	mu.Lock()
	if !cursorHidden {
		process.SetExitHandler(s.ShowCursor)
		cursorHidden = true
	}
	mu.Unlock()
	s.out.WriteString("\x1b[?25l")
}

// Interruptible sets whether the reading of input from the terminal can be
// interrupted by certain control characters. If interruptible:
//
// * `^C` exits the process with an exit code of 130.
//
// * `^\` aborts the process with a panic and prints stacktraces.
//
// All `Screen` instances are interruptible by default.
func (s *Screen) Interruptible(state bool) {
	s.intr = state
}

// Print formats the operands like `fmt.Print` and writes to the terminal
// output.
func (s *Screen) Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(s, a...)
}

// Printf formats the operands like `fmt.Printf` and writes to the terminal
// output.
func (s *Screen) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(s, format, a...)
}

// Println formats the operands like `fmt.Println` and writes to the terminal
// output.
func (s *Screen) Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(s, a...)
}

// Read reads the terminal input.
func (s *Screen) Read(p []byte) (n int, err error) {
	if s.dev == nil {
		if err := s.makeRaw(); err != nil {
			return 0, err
		}
		defer s.reset()
	}
	if !s.intr {
		return s.in.Read(p)
	}
	n, err = s.in.Read(p)
	if n > 0 {
		for _, char := range p[:n] {
			switch InputKey(char) {
			case KeyCtrlC:
				s.reset()
				process.Exit(130) // 128 + SIGINT (signal 2)
			case KeyCtrlBackslash:
				s.reset()
				process.Crash()
			}
		}
	}
	return n, err
}

// ReadInput reads `Input` from the terminal.
func (s *Screen) ReadInput() (*Input, error) {
	if err := s.makeRaw(); err != nil {
		return nil, err
	}
	defer s.reset()
	var buf []byte
	rem := false
	if len(s.pending) > 0 {
		buf = s.pending
		rem = true
	} else {
		buf = make([]byte, 1)
		_, err := s.Read(buf)
		if err != nil {
			return nil, err
		}
	}
	defer func() {
		if rem {
			s.pending = s.pending[1:]
		}
	}()
	switch char := buf[0]; char {
	case '\r':
		return &Input{
			Key: KeyEnter,
		}, nil
	case 27:
		key := InputKey(0)
		seq := make([]byte, 1)
		_, err := s.in.Read(seq)
		if err != nil {
			goto escape
		}
		if seq[0] != '[' {
			s.pending = append(s.pending, seq[0])
			goto escape
		}
		_, err = s.in.Read(seq)
		if err != nil {
			s.pending = append(s.pending, '[')
			goto escape
		}
		switch seq[0] {
		case 'A':
			key = KeyUp
		case 'B':
			key = KeyDown
		case 'C':
			key = KeyRight
		case 'D':
			key = KeyLeft
		case 'F':
			key = KeyEnd
		case 'H':
			key = KeyHome
		case '5', '6':
			seq2 := make([]byte, 1)
			_, err = s.in.Read(seq2)
			if err != nil {
				s.pending = append(s.pending, '[', seq[0])
				goto escape
			}
			if seq2[0] != '~' {
				s.pending = append(s.pending, '[', seq[0], seq2[0])
				goto escape
			}
			switch seq[0] {
			case '5':
				key = KeyPageUp
			case '6':
				key = KeyPageDown
			}
		default:
			s.pending = append(s.pending, '[', seq[0])
			goto escape
		}
		return &Input{
			Key: key,
		}, nil
	escape:
		return &Input{
			Key: KeyEscape,
		}, nil
	default:
		if char < 32 || char == 127 {
			return &Input{
				Key: InputKey(char),
			}, nil
		}
		return &Input{
			Byte: char,
		}, nil
	}
}

// ReadSecret prompts the user for a secret without echoing. The prompt is
// written to `os.Stderr` as that is more likely to be seen, e.g. if a user has
// redirected stdout.
//
// Unlike the other read-related methods, this method defers to the platform for
// processing input and handling interrupt characters like `^C`. Special
// consideration is only given to the backspace character which overwrites the
// previous byte when one is present.
func (s *Screen) ReadSecret(prompt string) ([]byte, error) {
	if prompt != "" {
		_, err := os.Stderr.WriteString(prompt)
		if err != nil {
			return nil, err
		}
	}
	if err := s.makeRaw(Canonical, GenSignals); err != nil {
		return nil, err
	}
	defer s.reset()
	return s.readline()
}

// Readline keeps reading from the terminal input until a `\n` (or `\r` on
// Windows) is encountered.
func (s *Screen) Readline() ([]byte, error) {
	return nil, nil
}

// ReadlineWithPrompt emits the given `prompt` to the terminal output before
// invoking `Readline`.
func (s *Screen) ReadlineWithPrompt(prompt string) ([]byte, error) {
	_, err := s.out.WriteString(prompt)
	if err != nil {
		return nil, err
	}
	return s.Readline()
}

// ShowCursor makes the cursor visible.
func (s *Screen) ShowCursor() {
	s.out.WriteString("\x1b[?25h")
}

// TrueColor returns whether the terminal supports 24-bit colors.
func (s *Screen) TrueColor() bool {
	return s.trueColor(os.Getenv("COLORTERM"))
}

// Write writes to the terminal output.
func (s *Screen) Write(p []byte) (n int, err error) {
	return s.out.Write(p)
}

// WriteString is like `Write`, but writes a string instead of a byte slice.
func (s *Screen) WriteString(p string) (n int, err error) {
	return s.Write([]byte(p))
}

func (s *Screen) makeRaw(opts ...RawOption) error {
	if s.dev != nil {
		return nil
	}
	opts = append(opts, ConvertLineEndings)
	d, err := MakeRaw(s.in, opts...)
	if err != nil {
		return err
	}
	s.dev = d
	return nil
}

func (s *Screen) readline() ([]byte, error) {
	var out []byte
	buf := make([]byte, 1)
	for {
		n, err := s.Read(buf)
		if n == 1 {
			switch buf[0] {
			case '\b', 127:
				if len(out) > 0 {
					out = out[:len(out)-1]
				}
			case '\n':
				if !isWindows {
					return out, nil
				}
			case '\r':
				if isWindows {
					return out, nil
				}
			default:
				out = append(out, buf[0])
			}
		}
		if err != nil {
			if err == io.EOF {
				return out, nil
			}
			return nil, err
		}
	}
}

func (s *Screen) reset() error {
	if s.dev != nil {
		err := s.dev.Reset()
		s.dev = nil
		return err
	}
	return nil
}

func (s *Screen) trueColor(env string) bool {
	// NOTE(tav): We assume the terminal is not lying if COLORTERM has a valid
	// value. However, this may not be set system-wide or forwarded via sudo,
	// ssh, etc.
	//
	// So we fallback by setting a 24-bit value followed by a query to the
	// terminal to see if it actually set the color. Unfortunately, some common
	// terminals don't support DECRQSS SGR requests.
	if env == "truecolor" || env == "24bit" {
		return true
	}
	// Get the current cursor position.
	prev, err := s.CursorPos()
	if err != nil {
		return false
	}
	if err := s.makeRaw(NonBlocking); err != nil {
		return false
	}
	defer func() {
		s.reset()
		// If the cursor moved after the DECRQSS request, e.g. due to the
		// terminal not parsing the request properly, move the cursor back and
		// overwrite the unintended output.
		now, err := s.CursorPos()
		if err != nil {
			return
		}
		if now.Col != prev.Col || now.Row != prev.Row {
			s.CursorTo(prev.Row, prev.Col)
			s.ClearToEnd()
		}
	}()
	// Set an unlikely foreground color, and then send the terminal a DECRQSS
	// SGR request to see if it has set it.
	_, err = s.out.WriteString("\x1b[38:2::1:2:3m\x1bP$qm\x1b\\")
	if err != nil {
		return false
	}
	// Give the terminal some time to respond.
	time.Sleep(100 * time.Millisecond)
	// Try reading a response.
	buf := make([]byte, 32)
	n, err := s.Read(buf)
	// Make sure to clear the set style after reading the response.
	defer s.out.WriteString("\x1b[0m")
	// Exit early on invalid data.
	if err != nil || n < 13 {
		return false
	}
	resp := string(buf[:n])
	if !strings.HasPrefix(resp, "\x1bP1$r") || !strings.HasSuffix(resp, "m\x1b\\") {
		return false
	}
	return strings.Contains(resp, ":1:2:3m")
}

// Canonical provides a `RawOption` to enable line-based canonical/cooked input
// processing.
func Canonical(r *RawConfig) {
	r.canon = true
}

// ConvertLineEndings provides a `RawOption` to automatically convert CR to NL
// on input, and NL to CRNL when writing output.
func ConvertLineEndings(r *RawConfig) {
	r.crnl = true
}

// GenSignals provides a `RawOption` to turn control characters like `^C` and
// `^Z` into signals instead of passing them through directly as characters.
func GenSignals(r *RawConfig) {
	r.sig = true
}

// IsTTY checks whether the given device file is connected to a terminal.
func IsTTY(f *os.File) bool {
	return isTTY(int(f.Fd()))
}

// MakeRaw converts the given device file into "raw" mode, i.e. disables
// echoing, disables special processing of certain characters, etc.
func MakeRaw(f *os.File, opts ...RawOption) (*Device, error) {
	cfg := &RawConfig{
		block: true,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	d, err := getDevice(int(f.Fd()))
	if err != nil {
		return nil, err
	}
	if err := makeRaw(d, *cfg); err != nil {
		return nil, err
	}
	return &Device{
		d: d,
	}, nil
}

// New instantiates a new `Screen` for interactive terminal applications.
func New() (*Screen, error) {
	var (
		err error
		in  *os.File
		out *os.File
	)
	if isWindows {
		in, err = os.OpenFile("CONIN$", os.O_RDWR, 0)
		if err != nil {
			return nil, fmt.Errorf("term: unable to open terminal input: %w", err)
		}
		out, err = os.OpenFile("CONOUT$", os.O_RDWR, 0)
		if err != nil {
			return nil, fmt.Errorf("term: unable to open terminal output: %w", err)
		}
	} else {
		in, err = os.OpenFile("/dev/tty", os.O_RDWR|os.O_SYNC|syscall.O_NOCTTY, 0)
		if err != nil {
			return nil, fmt.Errorf("term: unable to open terminal input/output: %w", err)
		}
		out = in
	}
	return &Screen{
		in:   in,
		intr: true,
		out:  out,
	}, nil
}

// NonBlocking provides a `RawOption` that configures the device to simulate
// non-blocking behavior by allowing `Read` calls to return immediately when
// there is no data to read.
//
// This should be used sparingly as it could degrade performance. However, it
// should still be better than changing the blocking mode with `O_NONBLOCK`,
// e.g. changing the mode on stdin could easily break the shell under normal
// circumstances when the program exits.
func NonBlocking(r *RawConfig) {
	r.block = false
}

// WatchResize sends updated dimensions whenever the terminal window is resized.
func WatchResize(ctx context.Context, f *os.File) (<-chan Dimensions, error) {
	_, err := WindowSize(f)
	if err != nil {
		return nil, err
	}
	ch := make(chan Dimensions)
	go watchResize(ctx, int(f.Fd()), ch)
	return ch, nil
}

// WindowSize returns the dimensions of the terminal.
func WindowSize(f *os.File) (Dimensions, error) {
	return windowSize(int(f.Fd()))
}
