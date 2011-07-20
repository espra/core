// Public Domain (-) 2010-2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

// YAML Parser
// ===========
//
// An extremely primitive and incomplete YAML parser.
//
// The main ``Parse`` function returns a ``map[string]*Data`` object, whilst the
// ``ParseDict`` function only handles pure string key-value pairs and returns a
// ``map[string]string``.
//
package yaml

import (
	"bytes"
	"io/ioutil"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	Null = iota
	String
	Bool
	Float
	Int
	List
	Map
)

type Data struct {
	Type   int
	String string
	Bool   bool
	Float  float64
	Int    int
	List   []*Data
	Map    map[string]*Data
}

func matchNumber(value string) (match bool, floatesque bool) {
	for i, char := range value {
		if char == '.' {
			floatesque = true
			continue
		} else if char == '-' && i == 0 {
			continue
		} else if char < '0' || char > '9' {
			return false, false
		}
	}
	return true, floatesque
}

func setValue(elem *Data, value string) {
	valueLength := len(value)
	lowercase := strings.ToLower(value)
	if lowercase == "true" || lowercase == "on" || lowercase == "yes" {
		elem.Type = Bool
		elem.Bool = true
	} else if lowercase == "false" || lowercase == "off" || lowercase == "no" {
		elem.Type = Bool
		elem.Bool = false
	} else if value == "~" || lowercase == "null" {
		elem.Type = Null
	} else if valueLength > 1 && ((value[0] == '"' && value[valueLength-1] == '"') || (value[0] == '\'' && value[valueLength-1] == '\'')) {
		elem.Type = String
		elem.String = value[1 : valueLength-1]
	} else if match, floatesque := matchNumber(value); match {
		if floatesque {
			floatval, err := strconv.Atof64(value)
			if err == nil {
				elem.Type = Float
				elem.Float = floatval
			} else {
				elem.Type = String
				elem.String = value
			}
		} else {
			intval, err := strconv.Atoi(value)
			if err == nil {
				elem.Type = Int
				elem.Int = intval
			} else {
				elem.Type = String
				elem.String = value
			}
		}
	} else {
		elem.Type = String
		elem.String = value
	}
}

func getKeyValue(lineno int, line string) (key, value string, err os.Error) {
	split := strings.SplitN(line, ":", 2)
	if len(split) != 2 {
		err = fmt.Errorf(
			"YAML Error: Expected a property name on line %d: %q",
			lineno, line)
		return
	}
	key = split[0]
	value = strings.TrimSpace(split[1])
	return
}

func Parse(input string) (root map[string]*Data, err os.Error) {

	var (
		elem   *Data
		indent int
		key    string
		lineno int
		value  string
	)

	root = make(map[string]*Data)

	for _, line := range strings.Split(input, "\n") {

		for i, total := 0, len(line); i < total; i++ {
			if line[i] != ' ' {
				indent = i
				break
			}
		}

		line = line[indent:]
		lineno += 1

		// Ignore blank lines.
		if line == "" {
			continue
		}

		// Strip out all comments.
		if line[0] == '#' {
			continue
		}

		if key == "" {
			key, value, err = getKeyValue(lineno, line)
			if err != nil {
				return
			}
			elem = &Data{}
			root[key] = elem
			if value != "" {
				setValue(elem, value)
				key = ""
			}
			continue
		}

		if len(line) > 1 && line[:2] == "- " {
			value = strings.TrimSpace(line[2:])
			listElem := &Data{}
			if value != "" {
				setValue(listElem, value)
			}
			switch elem.Type {
			case Null:
				elem.Type = List
				elem.List = []*Data{listElem}
			case List:
				elem.List = append(elem.List, listElem)
			default:
				return nil, fmt.Errorf(
					"YAML Error: Conflicting %s value for list item on line %d: %q",
					getTypeName(elem.Type), lineno, line)
			}
		} else {
			key, value, err = getKeyValue(lineno, line)
			if err != nil {
				return
			}
			elem = &Data{}
			root[key] = elem
			if value != "" {
				setValue(elem, value)
				key = ""
			}

		}

	}

	return

}

func ParseFile(filename string) (map[string]*Data, os.Error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return Parse(string(input))
}

func ParseDict(input string) map[string]string {
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

func ParseDictFile(filename string) (map[string]string, os.Error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseDict(string(input)), nil
}

func getTypeName(typ int) (name string) {
	switch typ {
	case String:
		name = "String"
	case Bool:
		name = "Bool"
	case Float:
		name = "Float"
	case Int:
		name = "Int"
	case List:
		name = "List"
	case Map:
		name = "Map"
	case Null:
		name = "Null"
	}
	return
}

func Display(root map[string]*Data) string {
	buffer := &bytes.Buffer{}
	for key, data := range root {
		fmt.Fprintf(buffer, "%s: ", key)
		data.Display(buffer, "  ")
		buffer.WriteByte('\n')
	}
	return buffer.String()
}

func (data *Data) Display(buffer *bytes.Buffer, indent string) {
	switch data.Type {
	case String:
		fmt.Fprintf(buffer, "%q", data.String)
	case Bool:
		fmt.Fprintf(buffer, "%t", data.Bool)
	case Float:
		fmt.Fprintf(buffer, "%f", data.Float)
	case Int:
		fmt.Fprintf(buffer, "%d", data.Int)
	case List:
		for _, elem := range data.List {
			fmt.Fprintf(buffer, "\n%s- ", indent)
			elem.Display(buffer, indent+"  ")
		}
	case Map:
		for key, elem := range data.Map {
			fmt.Fprintf(buffer, "\n%s%s: ", indent, key)
			elem.Display(buffer, indent+"  ")
		}
	case Null:
		fmt.Fprint(buffer, "null")
	}
}
