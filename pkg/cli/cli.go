// Public Domain (-) 2010-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// Package cli provides an easy way to build command line applications.
package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"web4.cc/pkg/process"
)

// Command specifies the minimal set of methods that a Command needs to
// implement. Commands wishing to have more fine-grained control, can also
// implement the Completer and Usage interfaces.
type Command interface {
	Info() string
	Run(c *Context) error
}

// Completer defines the interface that a Command should implement if it wants
// to provide custom autocompletion on command line arguments.
type Completer interface {
	Complete()
}

// Context provides a way to access processed command line info at specific
// points within the command hierarchy.
type Context struct {
	args   []string
	cmd    Command
	name   string
	opts   []*Option
	parent *Context
	root   *Context
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

// Name returns the command name for the current context.
func (c *Context) Name() string {
	return c.name
}

// Options returns the command line arguments for the current context.
func (c *Context) Options() []*Option {
	opts := make([]*Option, len(c.opts))
	copy(opts, c.opts)
	return opts
}

// Parent returns the parent of the current context.
func (c *Context) Parent() *Context {
	return c.parent
}

// Root returns the root context.
func (c *Context) Root() *Context {
	if c.root == nil {
		return c
	}
	return c.root
}

// RootName returns the command name for the root context.
func (c *Context) RootName() string {
	if c.root == nil {
		return c.name
	}
	return c.root.name
}

// Usage returns the generated usage for the current context.
func (c *Context) Usage() string {
	return c.usage()
}

// Option defines the command line option derived from a Command struct.
type Option struct {
	cmpl  int
	env   []string
	field int
	help  string
	long  []string
	req   bool
	short []string
}

// Env returns the environment variables associated with the option.
func (o *Option) Env() []string {
	return clone(o.env)
}

// Help returns the help info for the option.
func (o *Option) Help() string {
	return o.help
}

// LongFlags returns the long flags associated with the option.
func (o *Option) LongFlags() []string {
	return clone(o.long)
}

// Required returns whether the option has been marked as required.
func (o *Option) Required() bool {
	return o.req
}

// ShortFlags returns the short flags associated with the option.
func (o *Option) ShortFlags() []string {
	return clone(o.short)
}

// Subcommands defines the field type for defining subcommands on a struct.
type Subcommands map[string]Command

// Usage defines the interface that a Command should implement if it wants
// fine-grained control over the usage output. Otherwise, the usage is
// auto-generated from the command name, Info() output, and struct fields.
type Usage interface {
	Usage(c *Context) string
}

// Version provides a default implementation to use as a subcommand to output
// version info.
type Version string

func (v Version) Info() string {
	return "Show the #{RootName} version info"
}

func (v Version) Run(c *Context) error {
	fmt.Println(v)
	return nil
}

type plain struct {
	info string
	run  func(c *Context) error
}

func (p plain) Info() string {
	return p.info
}

func (p plain) Run(c *Context) error {
	return p.run(c)
}

// FromFunc will define a new Command from the given run function and info
// string. It's useful for defining commands where there's no need to handle any
// command line options.
func FromFunc(run func(c *Context) error, info string) Command {
	return plain{info, run}
}

// Run processes the command line arguments in the context of the given Command.
// The given program name will be used to auto-generate usage text and error
// messages.
func Run(name string, cmd Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("cli: missing program name in the given args slice")
	}
	c, err := newContext(name, cmd, args, nil)
	if err != nil {
		return err
	}
	return c.run()
}

// RunThenExit provides a utility function for the common case of calling Run
// with os.Args, printing the error on failure, and exiting with a status code
// of 1 on failure, and 0 on success.
//
// The function will use process.Exit instead of os.Exit so that registered exit
// handlers will be triggered.
func RunThenExit(name string, cmd Command) {
	err := Run(name, cmd, os.Args)
	if err != nil {
		printErrorf("%s failed: %s", name, err)
		process.Exit(1)
	}
	process.Exit(0)
}

func clone(xs []string) []string {
	ys := make([]string, len(xs))
	copy(ys, xs)
	return ys
}

func printErrorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
