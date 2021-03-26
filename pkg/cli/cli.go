// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// Package cli provides an easy way to build command line applications.
//
// If the value for a subcommand is nil, it is treated as if the command didn't
// even exist. This is useful for disabling the builtin subcommands like
// completion and help.
package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"web4.cc/pkg/process"
)

// Invalid argument types.
const (
	UnspecifiedInvalidArg InvalidArgType = iota
	InvalidEnv
	InvalidFlag
	InvalidValue
	MissingFlag
	MissingValue
	RepeatedFlag
	UnknownSubcommand
)

// ErrInvalidArg indicates that there was an invalid command line argument. It
// can be used as the target to errors.Is to test if the returned error from Run
// calls was as a result of mistyped command line arguments.
var ErrInvalidArg = errors.New("cli: invalid command line argument")

var (
	_ cmdrunner = (*Version)(nil)
	_ cmdrunner = (*plain)(nil)
)

// Command specifies the basic interface that a command needs to implement. For
// more fine-grained control, commands can also implement any of the Completer,
// Helper, InvalidArgHelper, and Runner interfaces.
type Command interface {
	About() *Info
}

// Completer defines the interface that a command should implement if it wants
// to provide custom autocompletion on command line arguments.
type Completer interface {
	Complete(c *Context) Completion
}

type Completion struct {
}

// Context provides a way to access processed command line info at specific
// points within the command hierarchy.
type Context struct {
	args   []string
	cmd    Command
	flags  []*Flag
	name   string
	opts   *optspec
	parent *Context
	subs   Subcommands
}

// Args returns the command line arguments for the current context.
func (c *Context) Args() []string {
	return clone(c.args)
}

// ChildContext tries to create a child Context for a subcommand.
func (c *Context) ChildContext(subcommand string) (*Context, error) {
	cmd := c.subs[subcommand]
	if cmd == nil {
		return nil, c.InvalidArg(UnknownSubcommand, subcommand, nil, nil)
	}
	sub := &Context{
		cmd:    cmd,
		name:   subcommand,
		parent: c,
	}
	if err := sub.init(); err != nil {
		return nil, err
	}
	return sub, nil
}

// Command returns the Command associated with the current context. By doing a
// type assertion on the returned value, this can be used to access field values
// of the parent or root context.
func (c *Context) Command() Command {
	return c.cmd
}

// Flags returns the command line flags for the current context.
func (c *Context) Flags() []*Flag {
	flags := make([]*Flag, len(c.flags))
	copy(flags, c.flags)
	return flags
}

// FullName returns the space separated sequence of command names, all the way
// from the root to the current context.
func (c *Context) FullName() string {
	path := []string{c.name}
	for c.parent != nil {
		c = c.parent
		path = append([]string{c.name}, path...)
	}
	return strings.Join(path, " ")
}

// Help returns the help text for a command. Commands wishing to override the
// auto-generated help text, must implement the Helper interface.
func (c *Context) Help() string {
	return c.help()
}

func (c *Context) InvalidArg(typ InvalidArgType, arg string, flag *Flag, err error) error {
	ia := &InvalidArg{
		Arg:     arg,
		Context: c,
		Err:     err,
		Flag:    flag,
		Type:    typ,
	}
	impl, ok := c.cmd.(InvalidArgHelper)
	if ok {
		printHelp(impl.InvalidArg(ia))
		return ia
	}
	x := c
	for x.parent != nil {
		x = x.parent
		impl, ok := x.cmd.(InvalidArgHelper)
		if ok {
			printHelp(impl.InvalidArg(ia))
			return ia
		}
	}
	printErrorf(ia.Details())
	help := c.contextualHelp(ia)
	if help != "" {
		fmt.Println("")
		printHelp(help)
	}
	return ia
}

// Name returns the command name for the current context.
func (c *Context) Name() string {
	return c.name
}

// Parent returns the parent of the current context.
func (c *Context) Parent() *Context {
	return c.parent
}

// PrintHelp outputs the command's help text to stdout.
func (c *Context) PrintHelp() {
	printHelp(c.help())
}

// Program returns the program name, i.e. the command name for the root context.
func (c *Context) Program() string {
	root := c.Root()
	if root == nil {
		return c.name
	}
	return root.name
}

// Root returns the root context.
func (c *Context) Root() *Context {
	for c.parent != nil {
		c = c.parent
	}
	return c
}

// Default makes it super easy to create tools with subcommands. Just
// instantiate the struct, with the relevant Info, Subcommands, and pass it to
// RunThenExit.
type Default struct {
	Info        *Info `cli:"-"`
	Subcommands Subcommands
}

func (d *Default) About() *Info {
	return d.Info
}

// Flag defines a command line flag derived from a Command struct.
type Flag struct {
	cmpl    int
	env     []string
	field   int
	help    string
	hide    bool
	label   string
	long    []string
	multi   bool
	req     bool
	setEnv  bool
	setFlag bool
	short   []string
	typ     string
}

// Env returns the environment variables associated with the flag.
func (f *Flag) Env() []string {
	return clone(f.env)
}

// Help returns the help info for the flag.
func (f *Flag) Help() string {
	return f.help
}

// Hidden returns whether the flag should be hidden from help output.
func (f *Flag) Hidden() bool {
	return f.hide
}

// Label returns the descriptive label for the flag option. This is primarily
// used to generate the help text, e.g.
//
//     --input-file path
//
// Boolean flags will always result in an empty string as the label. For all
// other types, the following sources are used in priority order:
//
// - Any non-empty value set using the "label" struct tag on the field.
//
// - Any labels that can be extracted from the help info by looking for the
// first non-whitespace separated set of characters enclosed within {braces}
// within the "help" struct tag on the field.
//
// - The field type, e.g. string, int, duration, etc. For non-builtin types,
// this will simply state "value".
func (f *Flag) Label() string {
	return f.label
}

// LongFlags returns the associated long flags.
func (f *Flag) LongFlags() []string {
	return clone(f.long)
}

// Multi returns whether the flag can be set multiple times.
func (f *Flag) Multi() bool {
	return f.multi
}

// Required returns whether the flag has been marked as required.
func (f *Flag) Required() bool {
	return f.req
}

// ShortFlags returns the associated short flags.
func (f *Flag) ShortFlags() []string {
	return clone(f.short)
}

// Helper defines the interface that a command should implement if it wants
// fine-grained control over the help text. Otherwise, the text is
// auto-generated from the command name, About() output, and struct fields.
type Helper interface {
	Help(c *Context) string
}

// Info
type Info struct {
	Short string
}

// InvalidArg
type InvalidArg struct {
	Arg     string
	Context *Context
	Err     error
	Flag    *Flag
	Type    InvalidArgType
}

func (i *InvalidArg) Details() string {
	// root := i.Context.parent == nil
	name := i.Context.FullName()
	// if !root {
	// name = "subcommand " + name
	// }
	switch i.Type {
	case InvalidFlag:
		return fmt.Sprintf("%s: invalid flag %q", name, i.Arg)
	case MissingFlag:
		flag := i.Flag
		if len(flag.long) > 0 {
			return fmt.Sprintf("%s: missing required flag --%s", name, flag.long[0])
		}
		if len(flag.short) > 0 {
			return fmt.Sprintf("%s: missing required flag -%s", name, flag.short[0])
		}
		return fmt.Sprintf("%s: missing required env %s", name, flag.env[0])
	case UnknownSubcommand:
		return fmt.Sprintf("%s: unknown command %q", name, i.Arg)
	default:
		return fmt.Sprintf("%#v\nType: %s", i, i.Type)
	}
	return "boom"
}

func (i *InvalidArg) Error() string {
	return fmt.Sprintf("cli: invalid command line argument: %s", i.Details())
}

func (i *InvalidArg) Is(target error) bool {
	return target == ErrInvalidArg
}

// InvalidArgHelper defines the interface that a command should implement to
// control the error output when an invalid command line argument is
// encountered.
//
// The returned string is assumed to be contextual help based on the InvalidArg,
// and will be emitted to stdout. Non-empty strings will have a newline appended
// to them if they don't already include one.
//
// If this interface isn't implemented, commands will default to printing an
// error about the invalid argument to stderr, followed by auto-generated
// contextual help text.
type InvalidArgHelper interface {
	InvalidArg(ia *InvalidArg) string
}

type InvalidArgType int

func (i InvalidArgType) String() string {
	switch i {
	case UnspecifiedInvalidArg:
		return "UnspecifiedInvalidArg"
	case InvalidEnv:
		return "InvalidEnv"
	case InvalidFlag:
		return "InvalidFlag"
	case InvalidValue:
		return "InvalidValue"
	case MissingFlag:
		return "MissingFlag"
	case MissingValue:
		return "MissingValue"
	case RepeatedFlag:
		return "RepeatedFlag"
	case UnknownSubcommand:
		return "UnknownSubcommand"
	default:
		return "UnknownInvalidArg"
	}
}

// Option configures the root context.
type Option func(c *Context)

// Runner defines the interface that a command should implement to handle
// command line arguments.
type Runner interface {
	Run(c *Context) error
}

// Subcommands defines the field type for defining subcommands on a struct.
type Subcommands map[string]Command

// Version provides a default implementation to use as a subcommand to output
// version info.
type Version string

func (v Version) About() *Info {
	return &Info{
		Short: "Show the {Program} version info",
	}
}

func (v Version) Run(c *Context) error {
	fmt.Println(v)
	return nil
}

type cmdrunner interface {
	Command
	Runner
}

type plain struct {
	info *Info
	run  func(c *Context) error
}

func (p plain) About() *Info {
	return p.info
}

func (p plain) Run(c *Context) error {
	return p.run(c)
}

type optspec struct {
	autoenv   bool
	envprefix string
	showenv   bool
	validate  bool
}

// AutoEnv enables the automatic derivation of environment variable names from
// the exported field names of Command structs. By default, the program name
// will be converted to SCREAMING_CASE with a trailing underscore, and used as a
// prefix for all generated environment variables. This can be controlled using
// the EnvPrefix Option.
func AutoEnv(c *Context) {
	c.opts.autoenv = true
}

// EnvPrefix enables AutoEnv and overrides the default prefix of the program
// name when automatically deriving environment variables. Use an empty string
// if the environment variables should be unprefixed.
//
// This function will panic if the given prefix is not empty or is invalid, i.e.
// not made up of uppercase letters and underscores. Non-empty values must not
// have a trailing underscore. One will be appended automatically.
func EnvPrefix(s string) func(*Context) {
	if !isEnv(s) {
		panic(fmt.Errorf("cli: invalid env prefix: %q", s))
	}
	if s != "" {
		s += "_"
	}
	return func(c *Context) {
		c.opts.autoenv = true
		c.opts.envprefix = s
	}
}

// FromFunc will define a new Command from the given run function and short info
// string. It's useful for defining commands where there's no need to handle any
// command line flags.
func FromFunc(run func(c *Context) error, info string) Command {
	return &plain{
		info: &Info{Short: info},
		run:  run,
	}
}

// NoValidate disables the automatic validation of all commands and subcommands.
// Validation adds to the startup time, and can be instead done by calling the
// Validate function from within tests.
func NoValidate(c *Context) {
	c.opts.validate = false
}

// Run processes the command line arguments in the context of the given Command.
// The given program name will be used to auto-generate help text and error
// messages.
func Run(name string, cmd Command, args []string, opts ...Option) error {
	if len(args) < 1 {
		return fmt.Errorf("cli: missing executable path in the given args slice")
	}
	c, err := newRoot(name, cmd, args[1:], opts...)
	if err != nil {
		return err
	}
	if c.opts.validate {
		if err := validate(c); err != nil {
			return err
		}
	}
	return c.run()
}

// RunThenExit provides a utility function that:
//
// * Calls Run with os.Args.
//
// * If Run returns an error, prints the error as long as it's not InvalidArg
// related.
//
// * Exits with a status code of 0 on success, 2 on InvalidArg, and 1 otherwise.
//
// The function will use process.Exit instead of os.Exit so that any registered
// exit handlers will run.
func RunThenExit(name string, cmd Command, opts ...Option) {
	err := Run(name, cmd, os.Args, opts...)
	if err != nil {
		if errors.Is(err, ErrInvalidArg) {
			process.Exit(2)
		}
		printErrorf("%s: %s", name, err)
		process.Exit(1)
	}
	process.Exit(0)
}

// ShowEnvHelp emits the associated environment variable names when
// auto-generating help text.
func ShowEnvHelp(c *Context) {
	c.opts.showenv = true
}

// Validate ensures that the given Command and all descendants have compliant
// struct tags and command names. Without this, validation only happens for the
// specific commands when they are executed on the command line.
func Validate(name string, cmd Command, opts ...Option) error {
	c, err := newRoot(name, cmd, nil, opts...)
	if err != nil {
		return err
	}
	return validate(c)
}

// NOTE(tav): We return copies of slices to callers so that they don't
// accidentally mutate them.
func clone(xs []string) []string {
	ys := make([]string, len(xs))
	copy(ys, xs)
	return ys
}

func printErrorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR\nERROR\t"+format+"\nERROR\n", args...)
}

func printHelp(help string) {
	if help == "" {
		return
	}
	if help[len(help)-1] != '\n' {
		fmt.Print(help + "\n")
	} else {
		fmt.Print(help)
	}
}
