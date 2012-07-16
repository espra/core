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
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
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
	Root map[string]*Elem
}

type Elem struct {
	Type   int
	String string
	Bool   bool
	Float  float64
	Int    int64
	List   []*Elem
	Map    map[string]*Elem
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

func setValue(elem *Elem, value string) {
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
			floatval, err := strconv.ParseFloat(value, 64)
			if err == nil {
				elem.Type = Float
				elem.Float = floatval
			} else {
				elem.Type = String
				elem.String = value
			}
		} else {
			intval, err := strconv.ParseInt(value, 10, 64)
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

func getKeyValue(lineno int, line string, needkey bool) (key, value string, err error) {
	split := strings.SplitN(line, ":", 2)
	if needkey {
		if len(split) != 2 {
			err = fmt.Errorf(
				"YAML Error: Expected a property name on line %d: %q",
				lineno, line)
			return
		}
		key = split[0]
		value = trim(split[1])
	} else if len(split) == 2 {
		key = split[0]
		value = trim(split[1])
	} else {
		value = trim(split[0])
	}
	return
}

func trim(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var (
		quoted    bool
		delimiter byte
	)
	for idx, length := 0, len(value); idx < length; idx++ {
		char := value[idx]
		if idx == 0 {
			switch char {
			case '"':
				quoted = true
				delimiter = '"'
			case '\'':
				quoted = true
				delimiter = '\''
			case '#':
				return ""
			}
		} else {
			if quoted && char == delimiter {
				return value[0 : idx+1]
			} else if char == '#' {
				return strings.TrimSpace(value[0:idx])
			}
		}
	}
	return value
}

func Parse(input string) (data *Data, err error) {

	var (
		elem   *Elem
		indent int
		key    string
		lineno int
		value  string
	)

	data = &Data{make(map[string]*Elem)}
	root := data.Root

	for _, line := range strings.Split(input, "\n") {

		indent = 0
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
			key, value, err = getKeyValue(lineno, line, true)
			if err != nil {
				return
			}
			elem = &Elem{}
			root[key] = elem
			if value != "" {
				setValue(elem, value)
				key = ""
			}
			continue
		}

		if len(line) > 1 && line[:2] == "- " {
			value = trim(line[2:])
			listElem := &Elem{}
			switch elem.Type {
			case Null:
				elem.Type = List
				elem.List = []*Elem{listElem}
			case List:
				elem.List = append(elem.List, listElem)
			default:
				return nil, fmt.Errorf(
					"YAML Error: Conflicting %s value for list item on line %d: %q",
					getTypeName(elem.Type), lineno, line)
			}
			if value != "" {
				subkey, subvalue, _ := getKeyValue(lineno, value, false)
				if subkey != "" {
					// TODO(tav): Handle maps nested within lists.
					_ = subvalue
				} else {
					setValue(listElem, value)
				}
			}
		} else {
			key, value, _ = getKeyValue(lineno, line, false)
			if key != "" {
				elem = &Elem{}
				root[key] = elem
			}
			if value != "" {
				setValue(elem, value)
				key = ""
			}
		}

	}

	return

}

func ParseFile(filename string) (*Data, error) {
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

func ParseDictFile(filename string) (map[string]string, error) {
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

func (data *Data) String() string {
	buffer := &bytes.Buffer{}
	for key, elem := range data.Root {
		fmt.Fprintf(buffer, "%s: ", key)
		elem.Display(buffer, "  ")
		buffer.WriteByte('\n')
	}
	return buffer.String()
}

func (data *Data) Get(key string, subkeys ...string) (elem *Elem, ok bool) {
	elem, ok = data.Root[key]
	if !ok {
		return
	}
	for _, key := range subkeys {
		if elem.Type != Map {
			return elem, false
		}
		elem, ok = elem.Map[key]
	}
	return
}

func (data *Data) GetString(key string, subkeys ...string) (value string, ok bool) {
	elem, ok := data.Get(key, subkeys...)
	if !ok {
		return
	}
	if elem.Type == String {
		return elem.String, true
	}
	return
}

func (data *Data) GetBool(key string, subkeys ...string) (value bool, ok bool) {
	elem, ok := data.Get(key, subkeys...)
	if !ok {
		return
	}
	if elem.Type == Bool {
		return elem.Bool, true
	}
	return
}

func (data *Data) GetFloat(key string, subkeys ...string) (value float64, ok bool) {
	elem, ok := data.Get(key, subkeys...)
	if !ok {
		return
	}
	if elem.Type == Float {
		return elem.Float, true
	}
	return
}

func (data *Data) GetInt(key string, subkeys ...string) (value int64, ok bool) {
	elem, ok := data.Get(key, subkeys...)
	if !ok {
		return
	}
	if elem.Type == Int {
		return elem.Int, true
	}
	return
}

func (data *Data) GetList(key string, subkeys ...string) (value []*Elem, ok bool) {
	elem, ok := data.Get(key, subkeys...)
	if !ok {
		return
	}
	if elem.Type == List {
		return elem.List, true
	}
	return
}

func (data *Data) GetMap(key string, subkeys ...string) (value map[string]*Elem, ok bool) {
	elem, ok := data.Get(key, subkeys...)
	if !ok {
		return
	}
	if elem.Type == Map {
		return elem.Map, true
	}
	return
}

func (data *Data) GetStringList(key string, subkeys ...string) (value []string, ok bool) {
	elem, ok := data.Get(key, subkeys...)
	if !ok {
		return
	}
	if elem.Type != List {
		return
	}
	value = make([]string, len(elem.List))
	i := 0
	for _, listElem := range elem.List {
		if listElem.Type == String {
			value[i] = listElem.String
			i += 1
		}
	}
	return value, true
}

var NeedPointerStructError = errors.New("yaml error: can only decode into pointer structs")

func (data *Data) LoadStruct(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Type().Kind() != reflect.Ptr {
		return NeedPointerStructError
	}
	rv = rv.Elem()
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return NeedPointerStructError
	}
	buf := &bytes.Buffer{}
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Anonymous {
			continue
		}
		var name string
		if tag := field.Tag.Get("yaml"); tag != "" {
			if tag == "-" {
				continue
			}
			name = tag
		}
		if name == "" {
			fname := field.Name
			rune, _ := utf8.DecodeRuneInString(fname)
			if !unicode.IsUpper(rune) {
				continue
			}
			buf.Reset()
			prevCap := true
			for idx, char := range fname {
				if idx == 0 {
					buf.WriteRune(unicode.ToLower(char))
					continue
				}
				if unicode.IsUpper(char) {
					if prevCap {
						buf.WriteRune(unicode.ToLower(char))
					} else {
						buf.WriteRune('-')
						buf.WriteRune(unicode.ToLower(char))
						prevCap = true
					}
				} else {
					buf.WriteRune(char)
					prevCap = false
				}
			}
			name = buf.String()
		}
		switch ft := field.Type; ft.Kind() {
		case reflect.String:
			v, _ := data.GetString(name)
			rv.Field(i).SetString(v)
		case reflect.Int64:
			v, _ := data.GetInt(name)
			rv.Field(i).SetInt(v)
		case reflect.Bool:
			v, _ := data.GetBool(name)
			rv.Field(i).SetBool(v)
		case reflect.Slice:
			if ft.Elem().Kind() == reflect.String {
				v, _ := data.GetStringList(name)
				rv.Field(i).Set(reflect.ValueOf(v))
			}
		}
	}
	return nil
}

func (data *Data) Size() int {
	return len(data.Root)
}

func (elem *Elem) Display(buffer *bytes.Buffer, indent string) {
	switch elem.Type {
	case String:
		fmt.Fprintf(buffer, "%q", elem.String)
	case Bool:
		fmt.Fprintf(buffer, "%t", elem.Bool)
	case Float:
		fmt.Fprintf(buffer, "%f", elem.Float)
	case Int:
		fmt.Fprintf(buffer, "%d", elem.Int)
	case List:
		for _, listElem := range elem.List {
			fmt.Fprintf(buffer, "\n%s- ", indent)
			listElem.Display(buffer, indent+"  ")
		}
	case Map:
		for key, mapElem := range elem.Map {
			fmt.Fprintf(buffer, "\n%s%s: ", indent, key)
			mapElem.Display(buffer, indent+"  ")
		}
	case Null:
		fmt.Fprint(buffer, "null")
	}
}
