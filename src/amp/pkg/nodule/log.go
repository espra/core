// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package nodule

import (
	"amp/logging"
	"fmt"
)

func Log(msg string, args ...interface{}) {
	if len(args) > 0 {
		logging.InfoData("node", fmt.Sprintf(msg, args...))
	} else {
		logging.InfoData("node", msg)
	}
}

func FilterConsoleLog(record *logging.Record) (write bool, data []interface{}) {
	if len(record.Items) > 0 {
		meta := record.Items[0]
		switch meta.(type) {
		case string:
			if meta.(string) == "node" {
				return true, record.Items[1:]
			}
		}
	}
	return true, nil
}
