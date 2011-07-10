// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// YAML Parser
// ===========
//
// An extremely primitive and incomplete YAML parser that only handles string
// key-value pairs for now.
package yaml

import (
	"io/ioutil"
	"os"
	"strings"
)

func Parse(input string) map[string]string {
	data := make(map[string]string)
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		split := strings.SplitN(line, ":", 2)
		if len(split) != 2 {
			continue
		}
		key := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])
		if len(key) == 0 || len(value) == 0 {
			continue
		}
		data[key] = value
	}
	return data
}

func ParseFile(filename string) (data map[string]string, err os.Error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return Parse(string(input)), nil
}
