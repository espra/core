// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package logging

import (
	"fmt"
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

var (
	Now            = time.Seconds()
	UTC            = time.UTC()
	ErrorReceivers = make([]chan *Record, 0)
	InfoReceivers  = make([]chan *Record, 0)
)

type Filter func(record *Record) (write bool, data []interface{})

type Record struct {
	Error bool
	Items []interface{}
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

func AddReceiver(receiver chan *Record, logType int) {
	if logType&InfoLog != 0 {
		InfoReceivers = append(InfoReceivers, receiver)
	}
	if logType&ErrorLog != 0 {
		ErrorReceivers = append(ErrorReceivers, receiver)
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
