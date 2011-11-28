// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package log

import (
	"amp/encoding"
	"fmt"
	"os"
)

var (
	ConsoleFilters = make(map[string]Filter)
	colors         = map[string]string{"info": "32", "error": "31"}
	colorify       = true
	checker        = make(chan int, 1)
	waiter         = make(chan int, 1)
	waitable       = false
)

type Filter func(items []interface{}) (bool, []interface{})

type ConsoleLogger struct {
	receiver chan *Record
}

func (logger *ConsoleLogger) log() {

	var record *Record
	var file *os.File
	var items []interface{}
	var prefix, status string
	var prefixErr, prefixInfo string
	var suffix []byte
	var write bool

	if colorify {
		prefixErr = fmt.Sprintf("\x1b[%sm", colors["error"])
		prefixInfo = fmt.Sprintf("\x1b[%sm", colors["info"])
		suffix = []byte("\x1b[0m\n")
	} else {
		suffix = []byte{'\n'}
	}

	for {
		select {
		case record = <-logger.receiver:
			items = record.Items
			write = true
			if filter, present := ConsoleFilters[record.Type]; present {
				write, items = filter(items)
				if !write || items == nil {
					continue
				}
			}
			if record.Error {
				file = os.Stderr
			} else {
				file = os.Stdout
			}
			if record.Error {
				prefix = prefixErr
				status = "ERROR: "
			} else {
				prefix = prefixInfo
				status = ""
			}
			mutex.RLock()
			fmt.Fprintf(file, "%s[%s-%s-%s %s:%s:%s] %s", prefix,
				encoding.PadInt64(utc.Year, 4), encoding.PadInt(utc.Month, 2),
				encoding.PadInt(utc.Day, 2), encoding.PadInt(utc.Hour, 2),
				encoding.PadInt(utc.Minute, 2), encoding.PadInt(utc.Second, 2),
				status)
			mutex.RUnlock()
			fmt.Fprint(file, items...)
			file.Write(suffix)
		case <-checker:
			if len(logger.receiver) > 0 {
				checker <- 1
				continue
			}
			waiter <- 1
		}
	}

}

func AddConsoleLogger() {
	waitable = true
	console := &ConsoleLogger{
		receiver: make(chan *Record, 100),
	}
	go console.log()
	AddReceiver(console.receiver, MixedLog)
}

func AddConsoleFilter(typeId string, filter Filter) {
	if filter == nil {
		return
	}
	ConsoleFilters[typeId] = filter
}

func DisableConsoleColors() {
	colorify = false
}

func SetConsoleColors(mapping map[string]string) {
	colors = mapping
}

func Wait() {
	if waitable {
		checker <- 1
		<-waiter
	}
}
