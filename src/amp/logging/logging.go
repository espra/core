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
)

const (
	endOfLine    = '\n'
	terminalByte = '\xff'
	template0    = "%v\xff\n"
	template1    = "%v\xfe%v\xff\n"
	template2    = "%v\xfe%v\xfe%v\xff\n"
	template3    = "%v\xfe%v\xfe%v\xfe%v\xff\n"
	template4    = "%v\xfe%v\xfe%v\xfe%v\xfe%v\xff\n"
	template5    = "%v\xfe%v\xfe%v\xfe%v\xfe%v\xfe%v\xff\n"
	template6    = "%v\xfe%v\xfe%v\xfe%v\xfe%v\xfe%v\xfe%v\xff\n"
)

type Logger struct {
	name      string
	directory string
	rotate    int
	file      *os.File
}

func (logger *Logger) Log(v ...interface{}) {
	_ = fmt.Printf
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
			encoding.PadInt(timestamp.Month, 2) + "-" + encoding.PadInt(timestamp.Day, 2) + ".h" +
			encoding.PadInt(timestamp.Hour, 2)
	}
	filename := logger.name + suffix + ".log"
	return path.Join(logger.directory, filename)
}

func fixUpLog(filename string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	var pointer int
	var seenTerminal bool
	for idx, char := range content {
		if char == terminalByte {
			seenTerminal = true
		} else if seenTerminal {
			if char == endOfLine {
				pointer = idx
			}
			seenTerminal = false
		}
	}
	os.Truncate(filename, int64(pointer))
}

func New(name string, directory string, rotate int) (logger *Logger, err os.Error) {
	logger = &Logger{
		name:      name,
		directory: directory,
		rotate:    rotate,
	}
	filename := logger.GetFilename(time.UTC())
	fixUpLog(filename)
	file, err := os.Open(filename, os.O_CREAT|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	logger.file = file
	return logger, nil
}
