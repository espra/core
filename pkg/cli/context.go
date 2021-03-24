// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package cli

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"time"

	"web4.cc/pkg/ident"
)

var (
	typeDuration        = reflect.TypeOf(time.Duration(0))
	typeSubcommands     = reflect.TypeOf(Subcommands{})
	typeTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeTime            = reflect.TypeOf(time.Time{})
)

func (c *Context) init() error {
	ptr := false
	rv := reflect.ValueOf(c.cmd)
	oriType := rv.Type()
	if rv.Kind() == reflect.Ptr {
		ptr = true
		rv = rv.Elem()
	}
	// Extract the subcommands mapping if a field with the right name and type
	// exists on a struct.
	rt := rv.Type()
	if rv.Kind() == reflect.Struct {
		field, ok := rt.FieldByName("Subcommands")
		if ok && field.Type == typeSubcommands {
			c.sub = rv.FieldByName("Subcommands").Interface().(Subcommands)
		}
	} else {
		ptr = false
	}
	// Skip processing of flags if the command isn't a struct pointer.
	if !ptr {
		return nil
	}
	// Process command line flags from the struct definition.
	seen := map[string]string{}
	flen := rt.NumField()
outer:
	for i := 0; i < flen; i++ {
		field := rt.Field(i)
		tag := field.Tag
		// Skip invalid fields.
		if field.PkgPath != "" || field.Anonymous || tag == "" {
			continue
		}
		// Process the field name.
		name, err := ident.FromPascal(field.Name)
		if err != nil {
			return fmt.Errorf(
				"cli: could not convert field name %s on %s: %s",
				field.Name, oriType, err,
			)
		}
		// Set defaults.
		flag := &Flag{
			cmpl:  -1,
			field: i,
			help:  strings.TrimSpace(tag.Get("help")),
			label: strings.TrimSpace(tag.Get("label")),
		}
		lflag := name.ToKebab()
		if prev, ok := seen[lflag]; ok {
			return fmt.Errorf(
				"cli: the derived --%s flag for field %s conflicts with %s on %s",
				lflag, field.Name, prev, oriType,
			)
		}
		flag.long = append(flag.long, lflag)
		seen[lflag] = field.Name
		root := c.Root()
		if !root.skipenv {
			env := root.envprefix + name.ToScreamingSnake()
			if prev, ok := seen[env]; ok {
				return fmt.Errorf(
					"cli: the derived environment variable %s for field %s conflicts with %s on %s",
					env, field.Name, prev, oriType,
				)
			}
			flag.env = []string{env}
			seen[env] = field.Name
		}
		// If no label has been specified, see if the help text has an embedded
		// label.
		if flag.label == "" && flag.help != "" {
			flag.help, flag.label = extractLabel(flag.help)
		}
		// Process the cli tag.
		for _, opt := range strings.Split(tag.Get("cli"), " ") {
			opt = strings.TrimSpace(opt)
			if opt == "" {
				continue
			}
			if opt == "-" {
				continue outer
			}
			if opt == "hide" {
				flag.hide = true
				continue
			}
			if opt == "inherit" {
				flag.inherit = true
				continue
			}
			if opt == "require" {
				flag.req = true
				continue
			}
			if opt == "skip:env" {
				flag.env = flag.env[1:]
				continue
			}
			if opt == "skip:flag" {
				flag.long = flag.long[1:]
				continue
			}
			if strings.HasPrefix(opt, "-") {
				if strings.ToLower(opt) != opt {
					goto invalid
				}
				if len(opt) == 2 && isFlagChar(opt[1]) {
					sflag := opt[1:]
					if prev, ok := seen[sflag]; ok {
						return fmt.Errorf(
							"cli: the -%s flag from field %s conflicts with %s on %s",
							sflag, field.Name, prev, oriType,
						)
					}
					flag.short = append(flag.short, sflag)
					seen[sflag] = field.Name
					continue
				}
				if strings.HasPrefix(opt, "--") && len(opt) >= 4 {
					lflag := opt[2:]
					for j := 0; j < len(lflag); j++ {
						if !isFlagChar(lflag[j]) {
							goto invalid
						}
					}
					if prev, ok := seen[lflag]; ok {
						return fmt.Errorf(
							"cli: the --%s flag from field %s conflicts with %s on %s",
							lflag, field.Name, prev, oriType,
						)
					}
					flag.long = append(flag.long, lflag)
					seen[lflag] = field.Name
					continue
				}
			invalid:
				return fmt.Errorf(
					"cli: invalid flag value %q found for field %s on %s",
					opt, field.Name, oriType,
				)
			}
			if opt == strings.ToUpper(opt) {
				isEnv := true
				for j := 0; j < len(opt); j++ {
					if !isEnvChar(opt[i]) {
						isEnv = false
						break
					}
				}
				if isEnv {
					if prev, ok := seen[opt]; ok {
						return fmt.Errorf(
							"cli: the environment variable %s for field %s conflicts with %s on %s",
							opt, field.Name, prev, oriType,
						)
					}
					flag.env = append(flag.env, opt)
					seen[opt] = field.Name
					continue
				}
			}
			if strings.HasPrefix(opt, "Complete") {
				meth, ok := oriType.MethodByName(opt)
				if !ok {
					return fmt.Errorf(
						"cli: completer method %s not found for field %s on %s",
						opt, field.Name, oriType,
					)
				}
				if flag.cmpl != -1 {
					return fmt.Errorf(
						"cli: completer already set for field %s on %s",
						field.Name, oriType,
					)
				}
				flag.cmpl = meth.Index
			} else {
				return fmt.Errorf(
					"cli: invalid cli tag value %q for field %s on %s",
					opt, field.Name, oriType,
				)
			}
		}
		// Figure out the flag type.
		flag.typ = getFlagType(field.Type, false)
		if flag.typ == "" {
			return fmt.Errorf(
				"cli: unsupported flag type %s for field %s on %s",
				field.Type, field.Name, oriType,
			)
		}
		if strings.HasPrefix(flag.typ, "[]") {
			flag.multi = true
		}
		if flag.typ == "bool" {
			flag.label = ""
		} else if flag.label == "" {
			flag.label = flag.typ
		}
		c.flags = append(c.flags, flag)
	}
	return nil
}

func (c *Context) run() error {
	if err := c.init(); err != nil {
		return err
	}
	// root := c.root == nil
	return c.cmd.Run(c)
}

func (c *Context) usage() string {
	impl, ok := c.cmd.(Usage)
	if ok {
		return impl.Usage(c)
	}
	b := strings.Builder{}
	b.WriteByte('\n')
	return b.String()
}

func extractLabel(help string) (string, string) {
	end := len(help)
	for i := 0; i < end; i++ {
		if help[i] == '{' {
			for j := i + 1; j < end; j++ {
				char := help[j]
				if char == ' ' {
					break
				}
				if char == '}' {
					if j-i == 1 {
						break
					}
					label := help[i+1 : j]
					return help[:i] + label + help[j+1:], label
				}
			}
		}
	}
	return help, ""
}

func getFlagType(rt reflect.Type, slice bool) string {
	switch kind := rt.Kind(); kind {
	case reflect.Bool:
		if slice {
			return ""
		}
		return "bool"
	case reflect.Float32:
		return "float32"
	case reflect.Float64:
		return "float64"
	case reflect.Int:
		return "int"
	case reflect.Int8:
		return "int8"
	case reflect.Int16:
		return "int16"
	case reflect.Int32:
		return "int32"
	case reflect.Int64:
		switch rt {
		case typeDuration:
			return "duration"
		default:
			return "int64"
		}
	case reflect.Interface, reflect.Ptr, reflect.Struct:
		if rt == typeTime {
			return "rfc3339"
		}
		switch kind {
		case reflect.Ptr:
			if rt.Elem() == typeTime {
				return "rfc3339"
			}
		case reflect.Struct:
			rt = reflect.PtrTo(rt)
		}
		if rt.Implements(typeTextUnmarshaler) {
			return "value"
		}
		return ""
	case reflect.Slice:
		if slice {
			// Only byte slices are supported as a potential slice type within a
			// slice.
			if rt.Elem().Kind() == reflect.Uint8 {
				return "string"
			}
			return ""
		}
		if rt.Elem().Kind() == reflect.Uint8 {
			return "string"
		}
		elem := getFlagType(rt.Elem(), true)
		if elem == "" {
			return elem
		}
		return "[]" + elem
	case reflect.String:
		return "string"
	case reflect.Uint:
		return "int"
	case reflect.Uint8:
		return "uint8"
	case reflect.Uint16:
		return "uint16"
	case reflect.Uint32:
		return "uint32"
	case reflect.Uint64:
		return "uint64"
	default:
		return ""
	}
}

func isEnvChar(char byte) bool {
	return char == '_' || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
}

func isFlagChar(char byte) bool {
	return char == '-' || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
}

func newContext(name string, cmd Command, args []string, parent *Context) (*Context, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if cmd == nil {
		fname := name
		if parent != nil {
			fname = parent.FullName() + " " + name
		}
		return nil, fmt.Errorf("cli: the Command instance for %q is nil", fname)
	}
	c := &Context{
		args: args,
		cmd:  cmd,
		name: name,
	}
	if parent != nil {
		c.parent = parent
		c.root = parent.root
	}
	return c, nil
}

func validateName(name string) error {
	for i := 0; i < len(name); i++ {
		char := name[i]
		if char == '-' || (char >= 'a' && char <= 'z') {
			continue
		}
		return fmt.Errorf("cli: invalid command name: %q", name)
	}
	return nil
}
