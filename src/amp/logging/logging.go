// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package logging

import (
	"amp/encoding"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

// Logs can be set to rotate either hourly, daily or never.
const (
	RotateNever  = 0
	RotateHourly = 1
	RotateDaily  = 2
	RotateTest   = 3
)

const (
	endOfLine    = '\n'
	terminalByte = '\xff'
)

var (
	endOfLogLine = []byte{'\xff', '\n'}
	typeLog      = []byte{'L'}
	typeInfo     = []byte{'I'}
	typeDebug    = []byte{'D'}
	typeError    = []byte{'E'}
	stdReceiver  = make(chan *Line)
)

type Logger struct {
	name      string
	directory string
	rotate    int
	file      *os.File
	stdout    bool
	receiver  chan *Line
	filename  string
}

type Line struct {
	format []byte
	error  bool
	items  []interface{}
}

func signalRotation(logger *Logger, signalChannel chan string) {
	var interval int64
	var filename string
	if logger.rotate == RotateDaily {
		interval = 86400000000000
	} else if logger.rotate == RotateHourly {
		interval = 3600000000000
	} else if logger.rotate == RotateTest {
		interval = 3000000000
	}
	for {
		filename = logger.GetFilename(time.UTC())
		if filename != logger.filename {
			signalChannel <- filename
		}
		time.Sleep(interval)
	}
}

func (logger *Logger) log() {

	timestamp := time.Seconds()
	go func() {
		for {
			time.Sleep(1000000)
			timestamp = time.Seconds()
		}
	}()

	rotateSignal := make(chan string)
	if logger.rotate > 0 {
		go signalRotation(logger, rotateSignal)
	}

	var line *Line
	var filename string

	for {
		select {
		case filename = <-rotateSignal:
			if filename != logger.filename {
				file, err := os.Open(filename, os.O_CREAT|os.O_WRONLY, 0666)
				if err == nil {
					logger.file.Close()
					logger.file = file
					logger.filename = filename
				} else {
					fmt.Printf("ERROR: Couldn't rotate log: %s\n", err)
				}
			}
		case line = <-logger.receiver:
			argLength := len(line.items)
			logger.file.Write(line.format)
			fmt.Fprintf(logger.file, "%v", timestamp)
			for i := 0; i < argLength; i++ {
				fmt.Fprintf(logger.file, "\xfe%v", line.items[i])
			}
			logger.file.Write(endOfLogLine)
		}
	}

}

func stdlog() {

	var line *Line
	timestamp := time.UTC()

	go func() {
		for {
			time.Sleep(1000000)
			timestamp = time.UTC()
		}
	}()

	for {
		line = <-stdReceiver
		argLength := len(line.items)
		_ = argLength
		// 		logger.file.Write(line.format)
		// 		logger.file.Write([]byte("2010-02"))
		// 		for i := 0; i < argLength; i++ {
		// 			fmt.Fprintf(logger.file, "\xfe%v", line.items[i])
		// 		}
		// 		logger.file.Write(endOfLogLine)
	}

}

func (logger *Logger) Log(v ...interface{}) {
	logger.receiver <- &Line{typeLog, false, v}
}

func (logger *Logger) Info(v ...interface{}) {
	logger.receiver <- &Line{typeInfo, false, v}
}

func (logger *Logger) Debug(v ...interface{}) {
	logger.receiver <- &Line{typeDebug, false, v}
}

func (logger *Logger) Error(v ...interface{}) {
	logger.receiver <- &Line{typeError, true, v}
}

func (logger *Logger) GetFilename(timestamp *time.Time) string {
	var suffix string
	switch logger.rotate {
	case RotateNever:
		suffix = ""
	case RotateDaily:
		suffix = "." + encoding.PadInt64(timestamp.Year, 4) + "-" +
			encoding.PadInt(timestamp.Month, 2) + "-" + encoding.PadInt(timestamp.Day, 2)
	case RotateHourly:
		suffix = "." + encoding.PadInt64(timestamp.Year, 4) + "-" +
			encoding.PadInt(timestamp.Month, 2) + "-" + encoding.PadInt(timestamp.Day, 2) + "." +
			encoding.PadInt(timestamp.Hour, 2)
	case RotateTest:
		suffix = "." + encoding.PadInt64(timestamp.Year, 4) + "-" +
			encoding.PadInt(timestamp.Month, 2) + "-" + encoding.PadInt(timestamp.Day, 2) + "." +
			encoding.PadInt(timestamp.Hour, 2) + "-" + encoding.PadInt(timestamp.Minute, 2) + "-" +
			encoding.PadInt(timestamp.Second, 2)
	}
	filename := logger.name + suffix + ".log"
	return path.Join(logger.directory, filename)
}

func fixUpLog(filename string) (pointer int) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	var seenTerminal bool
	for idx, char := range content {
		if char == terminalByte {
			seenTerminal = true
		} else if seenTerminal {
			if char == endOfLine {
				pointer = idx + 1
			}
			seenTerminal = false
		}
	}
	fmt.Printf("TERMINAL: %v\n", pointer)
	os.Truncate(filename, int64(pointer))
	return pointer
}

func New(name string, directory string, rotate int, stdout bool) (logger *Logger, err os.Error) {
	receiver := make(chan *Line)
	logger = &Logger{
		name:      name,
		directory: directory,
		rotate:    rotate,
		stdout:    stdout,
		receiver:  receiver,
	}
	filename := logger.GetFilename(time.UTC())
	pointer := fixUpLog(filename)
	file, err := os.Open(filename, os.O_CREAT|os.O_WRONLY, 0666)
	if err != nil {
		return logger, err
	}
	if pointer > 0 {
		file.Seek(int64(pointer), 0)
	}
	logger.file = file
	logger.filename = filename
	go logger.log()
	return logger, nil
}

func init() {
	go stdlog()
}
