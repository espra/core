// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// Package cli provides an easy way to build command line applications.
//
// If the value for a subcommand is nil, it is treated as if the command didn't
// even exist. This is useful for disabling the builtin subcommands like
// completion and help.
package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"web4.cc/pkg/process"
)

var (
	_ cmdrunner = (*Version)(nil)
	_ cmdrunner = (*plain)(nil)
)

// Command specifies the basic interface that a command needs to implement. For
// more fine-grained control, commands can also implement the Completer, Helper,
// and Runner interfaces.
type Command interface {
	Info() *Info
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
	sub    Subcommands
}

// Args returns the command line arguments for the current context.
func (c *Context) Args() []string {
	return clone(c.args)
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
		path = append(path, c.name)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(path)))
	return strings.Join(path, " ")
}

// Help returns the help text for a command. Commands wishing to override the
// auto-generated help text, must implement the Helper interface.
func (c *Context) Help() string {
	return c.help()
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
	fmt.Print(c.help())
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

// Flag defines a command line flag derived from a Command struct.
type Flag struct {
	cmpl    int
	env     []string
	field   int
	help    string
	hide    bool
	inherit bool
	label   string
	long    []string
	multi   bool
	req     bool
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

// Inherited returns whether the flag will be inherited by any subcommands.
func (f *Flag) Inherited() bool {
	return f.inherit
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
// auto-generated from the command name, Info() output, and struct fields.
type Helper interface {
	Help(c *Context) string
}

// Info
type Info struct {
	Short string
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

func (v Version) Info() *Info {
	return &Info{
		Short: "Show the #{Program} version info",
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

func (p plain) Info() *Info {
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

// EnvPrefix overrides the default prefix of the program name when automatically
// deriving environment variables. Use an empty string if the environment
// variables should be unprefixed.
//
// This function will panic if the given prefix is not empty or made up of
// uppercase letters and underscores. Non-empty values must not have a trailing
// underscore. One will be appended automatically.
func EnvPrefix(s string) func(*Context) {
	if !isEnv(s) {
		panic(fmt.Errorf("cli: invalid env prefix: %q", s))
	}
	if s != "" {
		s += "_"
	}
	return func(c *Context) {
		c.opts.envprefix = s
	}
}

// FromFunc will define a new Command from the given run function and short info
// string. It's useful for defining commands where there's no need to handle any
// command line flags.
func FromFunc(run func(c *Context) error, info string) Command {
	return plain{
		info: &Info{Short: info},
		run:  run,
	}
}

// NoAutoEnv disables the automatic derivation of environment variable names
// from the exported field names of Command structs.
func NoAutoEnv(c *Context) {
	c.opts.autoenv = false
}

// NoValidate disables the automatic validation of all commands and subcommands.
// Validation adds to the startup time, and can be instead done by calling the
// Validate function directly from within tests.
func NoValidate(c *Context) {
	c.opts.validate = false
}

// ShowEnvHelp emits the associated environment variable names when
// auto-generating help text.
func ShowEnvHelp(c *Context) {
	c.opts.showenv = true
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

// RunThenExit provides a utility function for the common case of calling Run
// with os.Args, printing the error on failure, and exiting with a status code
// of 1 on failure, and 0 on success.
//
// The function will use process.Exit instead of os.Exit so that registered exit
// handlers will run.
func RunThenExit(name string, cmd Command, opts ...Option) {
	err := Run(name, cmd, os.Args, opts...)
	if err != nil {
		printErrorf("%s failed: %s", name, err)
		process.Exit(1)
	}
	process.Exit(0)
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
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
