// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package cli

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"web4.cc/pkg/ident"
)

var (
	typeCompletion      = reflect.TypeOf(&Completion{})
	typeContext         = reflect.TypeOf(&Context{})
	typeDuration        = reflect.TypeOf(time.Duration(0))
	typeSubcommands     = reflect.TypeOf(Subcommands{})
	typeTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeTime            = reflect.TypeOf(time.Time{})
)

func (c *Context) contextualHelp(ia *InvalidArg) string {
	b := strings.Builder{}
	b.WriteString("Contextual Usage: \n")
	return b.String()
}

func (c *Context) defaultHelp() string {
	b := strings.Builder{}
	b.WriteString("Default Usage: ")
	b.WriteString(c.FullName())
	b.WriteByte('\n')
	return b.String()
}

func (c *Context) help() string {
	impl, ok := c.cmd.(Helper)
	if ok {
		return impl.Help(c)
	}
	x := c
	for x.parent != nil {
		x = x.parent
		impl, ok := x.cmd.(Helper)
		if ok {
			return impl.Help(c)
		}
	}
	return c.defaultHelp()
}

func (c *Context) init() error {
	ptr := false
	rv := reflect.ValueOf(c.cmd)
	oriType := rv.Type()
	if rv.Kind() == reflect.Ptr {
		ptr = true
		rv = rv.Elem()
	}
	// If it's a struct, ensure the original type is a pointer, and extract the
	// subcommands mapping if a field with the right name and type exists.
	rt := rv.Type()
	if rv.Kind() == reflect.Struct {
		if !ptr {
			if c.parent == nil {
				return fmt.Errorf(
					"cli: invalid Command for %q: Command structs must be pointers, not %s",
					c.name, oriType,
				)
			}
			return fmt.Errorf(
				"cli: invalid Command for the %q subcommand: Command structs must be pointers, not %s",
				c.FullName(), oriType,
			)
		}
		field, ok := rt.FieldByName("Subcommands")
		if ok {
			if field.Type != typeSubcommands {
				return fmt.Errorf(
					"cli: the Subcommands field on %s is not cli.Subcommands",
					oriType,
				)
			}
			subs := rv.FieldByName("Subcommands").Interface().(Subcommands)
			if c.parent == nil {
				for name, sub := range subs {
					c.subs[name] = sub
				}
			} else {
				c.subs = subs
			}
		}
		// Check for potential typo.
		_, ok = rt.FieldByName("SubCommands")
		if ok {
			return fmt.Errorf(
				"cli: invalid field SubCommands on %s: did you mean Subcommands?",
				oriType,
			)
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
		// Skip invalid fields and special fields.
		if field.PkgPath != "" || field.Anonymous || tag == "" {
			continue
		}
		if field.Name == "Subcommands" {
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
		optspec := c.Root().opts
		if optspec.autoenv {
			env := optspec.envprefix + name.ToScreamingSnake()
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
		skipenv := false
		for _, opt := range strings.Split(tag.Get("cli"), " ") {
			opt = strings.TrimSpace(opt)
			if opt == "" {
				continue
			}
			if opt == "-" {
				continue outer
			}
			if opt == "!autoenv" {
				if optspec.autoenv {
					flag.env = flag.env[1:]
					skipenv = true
				}
				continue
			}
			if opt == "!autoflag" {
				flag.long = flag.long[1:]
				continue
			}
			if opt == "hidden" {
				flag.hide = true
				continue
			}
			if opt == "required" {
				flag.req = true
				continue
			}
			if strings.HasPrefix(opt, "-") {
				if strings.ToLower(opt) != opt {
					goto invalid
				}
				if len(opt) == 2 && isShortFlag(opt[1]) {
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
					if !isLongFlag(lflag) {
						goto invalid
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
				if !isEnv(opt) {
					return fmt.Errorf(
						"cli: invalid environment variable %q found for field %s on %s",
						opt, field.Name, oriType,
					)
				}
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
			if strings.HasPrefix(opt, "Complete") {
				meth, ok := oriType.MethodByName(opt)
				if !ok {
					return fmt.Errorf(
						"cli: completer method %s not found for field %s on %s",
						opt, field.Name, oriType,
					)
				}
				if errmsg := isCompleter(meth.Type); errmsg != "" {
					return fmt.Errorf(
						"cli: invalid completer method %s for field %s on %s: %s",
						opt, field.Name, oriType, errmsg,
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
			max := 0
			if optspec.autoenv {
				max++
			}
			if skipenv {
				max--
			}
			if len(flag.env) > max {
				return fmt.Errorf(
					"cli: environment variables are not supported for slice types, as used for field %s on %s",
					field.Name, oriType,
				)
			}
			flag.env = nil
		}
		if flag.typ == "bool" {
			flag.label = ""
		} else if flag.label == "" {
			flag.label = flag.typ
		}
		// Error on missing env/flags.
		if len(flag.long) == 0 && len(flag.short) == 0 && len(flag.env) == 0 {
			if flag.multi {
				return fmt.Errorf(
					"cli: missing flags for field %s on %s", field.Name, oriType,
				)
			}
			return fmt.Errorf(
				"cli: missing flags or environment variables for field %s on %s",
				field.Name, oriType,
			)
		}
		c.flags = append(c.flags, flag)
	}
	return nil
}

func (c *Context) run() (err error) {
	// Initialize the Context.
	root := c.parent == nil
	if !root || !c.opts.validate {
		if err := c.init(); err != nil {
			return err
		}
	}
	// Process the environment variables.
	for _, flag := range c.flags {
		for _, env := range flag.env {
			val := os.Getenv(env)
			if val == "" {
				continue
			}
			if err := c.setEnv(flag, val); err != nil {
				return c.InvalidArg(InvalidEnv, val, nil, err)
			}
		}
	}
	// Process the command line arguments.
	var (
		help  bool
		fList []string
		fName string
		lArgs []string
		long  bool
		pArg  string
		pFlag *Flag
		rArgs []string
	)
outer:
	for i := 0; i < len(c.args); i++ {
		arg := c.args[i]
		// Handle any pending flag value.
		if pFlag != nil {
			if len(arg) > 0 && arg[0] == '-' {
				return c.InvalidArg(MissingValue, pArg, nil, nil)
			}
			if err := c.setFlag(pFlag, arg); err != nil {
				return err
			}
			pFlag = nil
			continue outer
		}
		// Skip flag processing if we see a double-dash.
		if arg == "--" {
			i++
			for ; i < len(c.args); i++ {
				rArgs = append(rArgs, c.args[i])
			}
			break outer
		}
		// Handle new flags.
		if strings.HasPrefix(arg, "-") {
			if strings.HasPrefix(arg, "--") {
				if len(arg) == 3 {
					return c.InvalidArg(InvalidFlag, arg, nil, nil)
				}
				if arg == "--help" {
					help = true
					continue outer
				}
				fName = arg[2:]
				long = true
			} else if len(arg) == 1 {
				rArgs = append(rArgs, "-")
				continue outer
			} else {
				if len(arg) != 2 {
					return c.InvalidArg(InvalidFlag, arg, nil, nil)
				}
				fName = arg[1:]
				long = false
			}
			for _, flag := range c.flags {
				if long {
					fList = flag.long
				} else {
					fList = flag.short
				}
				for _, name := range fList {
					if name != fName {
						continue
					}
					if flag.typ == "bool" {
						if err := c.setFlag(flag, "1"); err != nil {
							return err
						}
						continue outer
					}
					pArg = arg
					pFlag = flag
					continue outer
				}
			}
			return c.InvalidArg(InvalidFlag, arg, nil, nil)
		}
		// Accumulate all arguments after the first non-flag/value argument.
		for ; i < len(c.args); i++ {
			lArgs = append(lArgs, c.args[i])
		}
		break outer
	}
	// Check all required flags have been set.
	for _, flag := range c.flags {
		if flag.req && !(flag.setEnv || flag.setFlag) {
			return c.InvalidArg(MissingFlag, "", flag, nil)
		}
	}
	// Run the subcommand if there's a match.
	if len(lArgs) > 0 {
		name := lArgs[0]
		for sub, cmd := range c.subs {
			if name != sub {
				continue
			}
			if cmd == nil {
				break
			}
			if help {
				lArgs[0] = "--help"
				lArgs = append(lArgs, rArgs...)
			} else {
				lArgs = append(lArgs[1:], rArgs...)
			}
			csub, err := newContext(name, cmd, lArgs, c)
			if err != nil {
				return err
			}
			return csub.run()
		}
	}
	// Print the help text if --help was specified.
	if help {
		c.PrintHelp()
		return nil
	}
	cmd, runner := c.cmd.(Runner)
	// Handle non-Runners.
	if !runner {
		if len(lArgs) == 0 {
			c.PrintHelp()
			return nil
		}
		return c.InvalidArg(UnknownSubcommand, lArgs[0], nil, nil)
	}
	c.args = append(lArgs, rArgs...)
	return cmd.Run(c)
}

func (c *Context) setEnv(flag *Flag, val string) error {
	flag.setEnv = true
	return nil
}

func (c *Context) setFlag(flag *Flag, val string) error {
	// if seen[name] && !flag.multi {
	// 	return c.InvalidArg(RepeatedFlag, arg, nil, nil)
	// }
	flag.setFlag = true
	return nil
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

// NOTE(tav): These checks need to be kept in sync with any changes to the
// Completer interface.
func isCompleter(rt reflect.Type) string {
	if n := rt.NumIn(); n != 2 {
		return fmt.Sprintf("method must have 1 argument, not %d", n-1)
	}
	if in := rt.In(1); in != typeContext {
		return fmt.Sprintf("method's argument must be a *cli.Context, not %s", in)
	}
	if rt.NumOut() != 1 {
		return "method must have only one return value"
	}
	if out := rt.Out(0); out != typeCompletion {
		return fmt.Sprintf("method's return value must be a *cli.Completion, not %s", out)
	}
	return ""
}

func isEnv(env string) bool {
	last := len(env) - 1
	for i := 0; i < len(env); i++ {
		char := env[i]
		if i == 0 {
			if char < 'A' || char > 'Z' {
				return false
			}
			continue
		}
		if char == '_' {
			if i == last {
				return false
			}
			continue
		}
		if (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			continue
		}
		return false
	}
	return true
}

func isLongFlag(flag string) bool {
	last := len(flag) - 1
	for i := 0; i < len(flag); i++ {
		char := flag[i]
		if i == 0 {
			if char < 'a' || char > 'z' {
				return false
			}
			continue
		}
		if char == '-' {
			if i == last {
				return false
			}
			continue
		}
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			continue
		}
		return false
	}
	return true
}

func isShortFlag(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}

func isValidName(name string) bool {
	for i := 0; i < len(name); i++ {
		char := name[i]
		if char == '-' || (char >= 'a' && char <= 'z') {
			continue
		}
		return false
	}
	return true
}

func newContext(name string, cmd Command, args []string, parent *Context) (*Context, error) {
	if !isValidName(name) {
		if parent == nil {
			return nil, fmt.Errorf("cli: invalid program name: %q", name)
		}
		fname := parent.FullName()
		return nil, fmt.Errorf("cli: invalid name %q for the %q subcommand", name, fname)
	}
	c := &Context{
		args: args,
		cmd:  cmd,
		name: name,
	}
	if parent == nil {
		c.opts = &optspec{
			validate: true,
		}
	} else {
		c.parent = parent
	}
	return c, nil
}

func newRoot(name string, cmd Command, args []string, opts ...Option) (*Context, error) {
	if cmd == nil {
		return nil, fmt.Errorf("cli: the Command instance for %q is nil", name)
	}
	c, err := newContext(name, cmd, args, nil)
	if err != nil {
		return nil, err
	}
	upper := strings.ToUpper(name)
	c.opts.envprefix = strings.ReplaceAll(upper, "-", "_") + "_"
	for _, opt := range opts {
		opt(c)
	}
	c.subs = Subcommands{
		"completion": builtinCompletion,
		"help":       builtinHelp,
	}
	return c, nil
}

func validate(c *Context) error {
	if err := c.init(); err != nil {
		return err
	}
	for name, cmd := range c.subs {
		if cmd == nil {
			continue
		}
		sub, err := newContext(name, cmd, nil, c)
		if err != nil {
			return err
		}
		if err := validate(sub); err != nil {
			return err
		}
	}
	return nil
}
