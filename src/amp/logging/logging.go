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
	"strconv"
	"time"
)

// Logs can be set to rotate either hourly, daily or never.
const (
	RotateNever = iota
	RotateHourly
	RotateDaily
	RotateTest
)

const (
	InfoLog = 1 << iota
	ErrorLog
	MixedLog = InfoLog | ErrorLog
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
	ErrorReceivers = make([]chan *Record, 0)
	InfoReceivers  = make([]chan *Record, 0)
	ConsoleFilters = make([]Filter, 0)
)

var (
	checker  = make(chan int, 1)
	waiter   = make(chan int, 1)
	waitable = false
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
	switch logger.rotate {
	case RotateDaily:
		interval = 86400000000000
	case RotateHourly:
		interval = 3600000000000
	case RotateTest:
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
				message := strconv.Quote(fmt.Sprint(record.Items[i]))
				fmt.Fprintf(logger.file, "\xfe%s", message[0:len(message)-1])
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
		select {
		case record = <-logger.receiver:
			items = record.Items
			write = true
		case <-checker:
			if len(logger.receiver) > 0 {
				checker <- 1
				continue
			}
			waiter <- 1
			continue
		}
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
			status = "ERR"
		} else {
			status = "INF"
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

func Info(message string, v ...interface{}) {
	if len(v) > 0 {
		message = fmt.Sprintf(message, v...)
	}
	record := &Record{false, []interface{}{"m", message}}
	for _, receiver := range InfoReceivers {
		receiver <- record
	}
}

func InfoData(v ...interface{}) {
	record := &Record{false, v}
	for _, receiver := range InfoReceivers {
		receiver <- record
	}
}

func Error(message string, v ...interface{}) {
	if len(v) > 0 {
		message = fmt.Sprintf(message, v...)
	}
	record := &Record{true, []interface{}{"m", message}}
	for _, receiver := range ErrorReceivers {
		receiver <- record
	}
}

func ErrorData(v ...interface{}) {
	record := &Record{true, v}
	for _, receiver := range ErrorReceivers {
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

func AddFileLogger(name string, directory string, rotate int, logType int) (logger *FileLogger, err os.Error) {
	logger = &FileLogger{
		name:      name,
		directory: directory,
		rotate:    rotate,
		receiver:  make(chan *Record, 100),
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
	AddReceiver(logger.receiver, logType)
	return logger, nil
}

func AddConsoleLogger() {
	waitable = true
	console := &ConsoleLogger{
		receiver: make(chan *Record, 100),
	}
	go console.log()
	AddReceiver(console.receiver, MixedLog)
}

func AddReceiver(receiver chan *Record, logType int) {
	if logType&InfoLog != 0 {
		InfoReceivers = append(InfoReceivers, receiver)
	}
	if logType&ErrorLog != 0 {
		ErrorReceivers = append(ErrorReceivers, receiver)
	}
}

func AddConsoleFilter(filter Filter) {
	ConsoleFilters = append(ConsoleFilters, filter)
}

func Wait() {
	if waitable {
		checker <- 1
		<-waiter
	}
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
