// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"
)

const (
	endOfRecord  = '\n'
	terminalByte = '\xff'
)

var endOfLogRecord = []byte{'\xff', '\n'}

type FileLogger struct {
	name      string
	directory string
	rotate    int
	file      *os.File
	filename  string
	receiver  chan *Record
}

func (logger *FileLogger) log() {

	rotateSignal := make(chan string)
	if logger.rotate > 0 {
		go signalRotation(logger, rotateSignal)
	}

	var record *Record
	var filename string

	for {
		select {
		case filename = <-rotateSignal:
			if filename != logger.filename {
				file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
				if err == nil {
					logger.file.Close()
					logger.file = file
					logger.filename = filename
				} else {
					fmt.Fprintf(os.Stderr, "ERROR: Couldn't rotate log: %s\n", err)
				}
			}
		case record = <-logger.receiver:
			argLength := len(record.Items)
			if record.Error {
				logger.file.Write([]byte{'E'})
			} else {
				logger.file.Write([]byte{'I'})
			}
			mutex.RLock()
			fmt.Fprintf(logger.file, "%v", now)
			mutex.RUnlock()
			for i := 0; i < argLength; i++ {
				message := strconv.Quote(fmt.Sprint(record.Items[i]))
				fmt.Fprintf(logger.file, "\xfe%s", message[0:len(message)-1])
			}
			logger.file.Write(endOfLogRecord)
		}
	}

}

func (logger *FileLogger) GetFilename(timestamp time.Time) string {
	var suffix string
	switch logger.rotate {
	case RotateNever:
		suffix = ""
	case RotateDaily:
		suffix = timestamp.Format(".2006-01-02")
	case RotateHourly:
		suffix = timestamp.Format(".2006-01-02.03")
	case RotateTest:
		suffix = timestamp.Format(".2006-01-02.03-04-05")
	}
	filename := logger.name + suffix + ".log"
	return path.Join(logger.directory, filename)
}

func FixUpLog(filename string) (pointer int) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	var seenTerminal bool
	for idx, char := range content {
		if char == terminalByte {
			seenTerminal = true
		} else if seenTerminal {
			if char == endOfRecord {
				pointer = idx + 1
			}
			seenTerminal = false
		}
	}
	os.Truncate(filename, int64(pointer))
	return pointer
}

func signalRotation(logger *FileLogger, signalChannel chan string) {
	var interval time.Duration
	var filename string
	switch logger.rotate {
	case RotateDaily:
		interval = 86400000000000
	case RotateHourly:
		interval = 3600000000000
	case RotateTest:
		interval = 3000000000
	}
	for {
		mutex.RLock()
		filename = logger.GetFilename(now)
		mutex.RUnlock()
		if filename != logger.filename {
			signalChannel <- filename
		}
		<-time.After(interval)
	}
}

func AddFileLogger(name string, directory string, rotate int, logType int) (logger *FileLogger, err error) {
	logger = &FileLogger{
		name:      name,
		directory: directory,
		rotate:    rotate,
		receiver:  make(chan *Record, 100),
	}
	mutex.RLock()
	filename := logger.GetFilename(now)
	mutex.RUnlock()
	pointer := FixUpLog(filename)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return logger, err
	}
	if pointer > 0 {
		file.Seek(int64(pointer), 0)
	}
	logger.file = file
	logger.filename = filename
	go logger.log()
	AddReceiver(logger.receiver, logType)
	return logger, nil
}
