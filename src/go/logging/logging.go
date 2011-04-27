// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package logging

import (
	"amp/encoding"
	"fmt"
	"io"
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
	endOfRecord  = '\n'
	terminalByte = '\xff'
)

var (
	endOfLogRecord = []byte{'\xff', '\n'}
)

var (
	Now            = time.Seconds()
	UTC            = time.UTC()
	Receivers      = make([]chan *Record, 0)
	ConsoleFilters = make([]Filter, 0)
)

type Filter func(record *Record) (write bool, data []interface{})

type ConsoleLogger struct {
	receiver chan *Record
}

type FileLogger struct {
	name      string
	directory string
	rotate    int
	file      *os.File
	filename  string
	receiver  chan *Record
}

type NetworkLogger struct {
	fallback *FileLogger
	stream   *io.Writer
	receiver chan *Record
}

type Record struct {
	Error bool
	Items []interface{}
}

func signalRotation(logger *FileLogger, signalChannel chan string) {
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
		filename = logger.GetFilename(UTC)
		if filename != logger.filename {
			signalChannel <- filename
		}
		<-time.After(interval)
	}
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
			fmt.Fprintf(logger.file, "%v", Now)
			for i := 0; i < argLength; i++ {
				fmt.Fprintf(logger.file, "\xfe%v", record.Items[i])
			}
			logger.file.Write(endOfLogRecord)
		}
	}

}

func (logger *ConsoleLogger) log() {

	var record *Record
	var file *os.File
	var items []interface{}
	var status string
	var write bool

	for {
		record = <-logger.receiver
		items = record.Items
		write = true
		for _, filter := range ConsoleFilters {
			write, data := filter(record)
			if !write {
				break
			}
			if data != nil {
				items = data
				break
			}
		}
		if !write {
			continue
		}
		argLength := len(items)
		if record.Error {
			file = os.Stderr
		} else {
			file = os.Stdout
		}
		if record.Error {
			status = "ERROR"
		} else {
			status = "INFO"
		}
		fmt.Fprintf(file, "%s [%s-%s-%s %s:%s:%s]", status,
			encoding.PadInt64(UTC.Year, 4), encoding.PadInt(UTC.Month, 2),
			encoding.PadInt(UTC.Day, 2), encoding.PadInt(UTC.Hour, 2),
			encoding.PadInt(UTC.Minute, 2), encoding.PadInt(UTC.Second, 2))
		for i := 0; i < argLength; i++ {
			fmt.Fprintf(file, " %v", items[i])
		}
		file.Write([]byte("\n"))
	}

}

func Log(message string, v ...interface{}) {
	if len(v) > 0 {
		message = fmt.Sprintf(message, v...)
	}
	items := make([]interface{}, 1)
	items[0] = message
	record := &Record{false, v}
	for _, receiver := range Receivers {
		receiver <- record
	}
}

func Info(v ...interface{}) {
	record := &Record{false, v}
	for _, receiver := range Receivers {
		receiver <- record
	}
}

func Error(v ...interface{}) {
	record := &Record{true, v}
	for _, receiver := range Receivers {
		receiver <- record
	}
}

func (logger *FileLogger) GetFilename(timestamp *time.Time) string {
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

func AddFileLogger(name string, directory string, rotate int) (logger *FileLogger, err os.Error) {
	receiver := make(chan *Record, 100)
	logger = &FileLogger{
		name:      name,
		directory: directory,
		rotate:    rotate,
		receiver:  receiver,
	}
	filename := logger.GetFilename(UTC)
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
	AddReceiver(logger.receiver)
	return logger, nil
}

func AddConsoleLogger() {
	stdReceiver := make(chan *Record, 100)
	console := &ConsoleLogger{
		receiver: stdReceiver,
	}
	go console.log()
	AddReceiver(console.receiver)
}

func AddReceiver(receiver chan *Record) {
	Receivers = append(Receivers, receiver)
}

func AddConsoleFilter(filter Filter) {
	ConsoleFilters = append(ConsoleFilters, filter)
}

func init() {

	go func() {
		for {
			<-time.After(1000000000)
			Now = time.Seconds()
			UTC = time.UTC()
		}
	}()

}
