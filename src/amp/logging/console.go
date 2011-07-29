// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package logging

import (
	"amp/encoding"
	"fmt"
	"os"
)

var (
	ConsoleFilters = make([]Filter, 0)
	checker        = make(chan int, 1)
	waiter         = make(chan int, 1)
	waitable       = false
)

type ConsoleLogger struct {
	receiver chan *Record
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
	ConsoleFilters = append(ConsoleFilters, defaultConsoleFilter)
}

func defaultConsoleFilter(record *Record) (write bool, data []interface{}) {
	if len(record.Items) > 0 {
		meta := record.Items[0]
		switch meta.(type) {
		case string:
			if meta.(string) == "m" {
				return true, record.Items[1:]
			}
		}
	}
	return true, nil
}

func AddConsoleFilter(filter Filter) {
	if filter == nil {
		return
	}
	ConsoleFilters = append(ConsoleFilters, filter)
}

func Wait() {
	if waitable {
		checker <- 1
		<-waiter
	}
}
