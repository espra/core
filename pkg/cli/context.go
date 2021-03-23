// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package cli

import (
	"fmt"
	"reflect"
	"strings"

	"web4.cc/pkg/ident"
)

var typeSubcommands = reflect.TypeOf(Subcommands{})

func (c *Context) init() error {
	ptr := false
	rv := reflect.ValueOf(c.cmd)
	typ := rv.Type()
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
	// Skip processing of options if the command isn't a struct pointer.
	if !ptr {
		return nil
	}
	// Process command line options from the struct definition.
	flen := rt.NumField()
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
				"cli: could not convert field name %q on %s: %s",
				field.Name, rt, err,
			)
		}
		// Set defaults.
		cmpl := -1
		env := []string{name.ToScreamingSnake()}
		long := []string{name.ToKebab()}
		required := false
		short := []string{}
		// Process the flags tag.
		for _, flag := range strings.Split(tag.Get("flags"), " ") {
			if strings.TrimSpace(flag) == "" {
				continue
			}
			if strings.ToLower(flag) != flag {
				return fmt.Errorf(
					"cli: invalid flag %q found for field name %q on %s",
					flag, field.Name, rt,
				)
			}
			if strings.HasPrefix(flag, "--") && len(flag) >= 4 {
				long = append(long, flag[2:])
				continue
			}
			if strings.HasPrefix(flag, "-") && len(flag) == 2 {
				short = append(short, flag[1:])
				continue
			}
			return fmt.Errorf(
				"cli: invalid flag %q found for field name %q on %s",
				flag, field.Name, rt,
			)
		}
		// Process the opts tag.
		for _, opt := range strings.Split(tag.Get("opts"), ",") {
			if strings.TrimSpace(opt) == "" {
				continue
			}
			if opt == "required" {
				required = true
				continue
			}
			if opt == strings.ToUpper(opt) {
				env = append(env, opt)
				continue
			}
			meth, ok := typ.MethodByName(opt)
			if !ok {
				return fmt.Errorf(
					"cli: could not find method %q for completing field name %q on %s",
					opt, field.Name, rt,
				)
			}
			if cmpl != -1 {
				return fmt.Errorf(
					"cli: completer already set for field name %q on %s",
					field.Name, rt,
				)
			}
			cmpl = meth.Index
		}
		c.opts = append(c.opts, &Option{
			cmpl:  cmpl,
			env:   env,
			field: i,
			help:  tag.Get("help"),
			long:  long,
			req:   required,
			short: short,
		})
	}
	return nil
}

func (c *Context) run() error {
	return c.cmd.Run(c)
}

func (c *Context) usage() string {
	return ""
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
	if err := c.init(); err != nil {
		return nil, err
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
