// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package log

import (
	"fmt"
	stdlog "log"
	"sync"
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
	mutex          sync.RWMutex
	now            = time.Seconds()
	utc            = time.UTC()
	ErrorReceivers = make([]chan *Record, 0)
	InfoReceivers  = make([]chan *Record, 0)
)

type Record struct {
	Error bool
	Items []interface{}
	Type  string
}

func Info(v ...interface{}) {
	record := &Record{false, []interface{}{fmt.Sprint(v...)}, "m"}
	for _, receiver := range InfoReceivers {
		receiver <- record
	}
}

func Infof(format string, v ...interface{}) {
	record := &Record{false, []interface{}{fmt.Sprintf(format, v...)}, "m"}
	for _, receiver := range ErrorReceivers {
		receiver <- record
	}
}

func InfoData(typeId string, v ...interface{}) {
	record := &Record{false, v, typeId}
	for _, receiver := range InfoReceivers {
		receiver <- record
	}
}

func Error(v ...interface{}) {
	record := &Record{true, []interface{}{fmt.Sprint(v...)}, "m"}
	for _, receiver := range ErrorReceivers {
		receiver <- record
	}
}

func Errorf(format string, v ...interface{}) {
	record := &Record{true, []interface{}{fmt.Sprintf(format, v...)}, "m"}
	for _, receiver := range ErrorReceivers {
		receiver <- record
	}
}

func ErrorData(typeId string, v ...interface{}) {
	record := &Record{true, v, typeId}
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

type dummyWriter struct{}

func (w *dummyWriter) Write(p []byte) (int, error) {
	InfoData("m", string(p))
	return len(p), nil
}

func init() {

	// Hijack the standard library's log functionality.
	stdlog.SetFlags(0)
	stdlog.SetOutput(&dummyWriter{})

	// Setup a goroutine to update the time every second.
	go func() {
		for {
			<-time.After(1000000000)
			mutex.Lock()
			now = time.Seconds()
			utc = time.UTC()
			mutex.Unlock()
		}
	}()

}
