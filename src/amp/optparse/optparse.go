// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The optparse package provides utility functions for the parsing and
// autocompletion of command line arguments.
package optparse

import (
	"amp/slice"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type option struct {
	shortflag   string
	longflag    string
	dest        string
	usage       string
	valueType   string
	intValue    *int
	boolValue   *bool
	stringValue *string
}

type OptionParser struct {
	Usage           string
	Version         string
	ParseHelp       bool
	ParseVersion    bool
	Autocomplete    bool
	options         []*option
	short2options   map[string]*option
	shortflags      []string
	long2options    map[string]*option
	longflags       []string
	helpAdded       bool
	versionAdded    bool
}

func (opt *option) String() (output string) {
	output = "  "
	if opt.shortflag != "" {
		output += "-" + opt.shortflag
		if opt.longflag != "" {
			output += ", "
		}
	}
	if opt.longflag != "" {
		output += "--" + opt.longflag
	}
	if opt.dest != "" {
		output += "=" + opt.dest
	}

	length := len(output)
	if length >= 19 {
		output += "\n                    "
	} else {
		padding := make([]byte, 20 - length)
		for i, _ := range(padding) {
			padding[i] = 32
		}
		output += string(padding)
	}
	output += opt.usage
	output += "\n"
	return
}

func (op *OptionParser) computeFlags(flags []string, opt *option) (shortflag, longflag string) {
	for _, flag := range flags {
		if strings.HasPrefix(flag, "--") {
			longflag = strings.TrimLeft(flag, "-")
			op.long2options[longflag] = opt
			slice.AppendString(&op.longflags, longflag)
		} else if strings.HasPrefix(flag, "-") {
			shortflag = strings.TrimLeft(flag, "-")
			op.short2options[shortflag] = opt
			slice.AppendString(&op.shortflags, shortflag)
		} else {
			longflag = strings.TrimLeft(flag, "-")
			op.long2options[longflag] = opt
			slice.AppendString(&op.longflags, longflag)
		}
	}
	return
}

func (op *OptionParser) Default(flags []string, usage string, displayDest bool, dest ...string) (opt *option) {
	opt = new(option)
	opt.usage = usage
	opt.shortflag, opt.longflag = op.computeFlags(flags, opt)
	destSlice := []string(dest)
	if displayDest {
		if len(destSlice) > 0 {
			opt.dest = destSlice[0]
 		} else {
			if opt.longflag != "" {
				opt.dest = strings.ToUpper(opt.longflag)
			} else {
				opt.dest = strings.ToUpper(opt.shortflag)
			}
		}
	}
	length := len(op.options)
	if cap(op.options) == length {
        temp := make([]*option, length, (2 * length) + 1)
        for idx, item := range op.options {
            temp[idx] = item
        }
        op.options = temp
	}
	op.options = op.options[0:length+1]
	op.options[length] = opt
	return
}

func (op *OptionParser) Int(flags []string, value int, usage string, dest ...string) (result *int) {
	opt := op.Default(flags, usage, true, dest)
	opt.valueType = "int"
	opt.intValue = &value
	return &value
}

func (op *OptionParser) String(flags []string, value string, usage string, dest ...string) (result *string) {
	opt := op.Default(flags, usage, true, dest)
	opt.valueType = "string"
	opt.stringValue = &value
	return &value
}

func (op *OptionParser) Bool(flags []string, value bool, usage string) (result *bool) {
	opt := op.Default(flags, usage, false)
	opt.valueType = "bool"
	opt.boolValue = &value
	return &value
}

func (op *OptionParser) Parse(args []string) (remainder []string) {

	if op.ParseHelp && !op.helpAdded {
		op.Bool([]string{"-h", "--help"}, false, "show this help and exit")
		op.helpAdded = true
	}
	if op.ParseVersion && !op.versionAdded {
		op.Bool([]string{"-v", "--version"}, false, "show the version and exit")
		op.versionAdded = true
	}

	argLength := len(args) - 1
	if argLength == 0 {
		return
	}

	var opt *option
	var ok bool
	var flag string

	idx := 1

	for {
		arg := args[idx]
		noOpt := true
		if strings.HasPrefix(arg, "--") {
			flag = strings.TrimLeft(arg, "-")
			opt, ok = op.long2options[flag]
			if ok {
				noOpt = false
			}
		} else if strings.HasPrefix(arg, "-") {
			flag = strings.TrimLeft(arg, "-")
			opt, ok = op.short2options[flag]
			if ok {
				noOpt = false
			}
		} else {
			slice.AppendString(&remainder, arg)
			if idx == argLength {
				break
			} else {
				idx += 1
				continue
			}
		}
		if noOpt {
			fmt.Printf("%s: error: no such option: %s\n\n", args[0], arg)
			op.PrintUsage()
			os.Exit(1)
		}
		if opt.dest != "" {
			if idx == argLength {
				fmt.Printf("%s: error: %s option requires an argument\n\n", args[0], arg)
				op.PrintUsage()
				os.Exit(1)
			}
		}
		if opt.valueType == "bool" {
			if opt.longflag == "help" && op.ParseHelp {
				op.PrintUsage()
				os.Exit(1)
			} else if opt.longflag == "version" && op.ParseVersion {
				fmt.Printf("%s\n", op.Version)
				os.Exit(0)
			}
			*opt.boolValue = true
			idx += 1
		} else if opt.valueType == "string" {
			*opt.stringValue = args[idx+1]
			idx += 2
		} else if opt.valueType == "int" {
			intValue, err := strconv.Atoi(args[idx+1])
			if err != nil {
				fmt.Printf("%s: error: couldn't convert %s value '%s' to an integer\n\n", args[0], arg, args[idx+1])
				op.PrintUsage()
				os.Exit(1)
			}
			*opt.intValue = intValue
			idx += 2
		}
		if idx > argLength {
			break
		}
	}

	return

}

func (op *OptionParser) PrintUsage() {
	fmt.Print(op.Usage)
	if len(op.options) > 0 {
		fmt.Print("\nOptions:\n")
	}
	for _, opt := range op.options {
		fmt.Printf("%v", opt)
	}
}

// Utility constructor.
func Parser(usage string, version ...string) (op *OptionParser) {
	op = new(OptionParser)
	op.long2options = make(map[string]*option)
	op.short2options = make(map[string]*option)
	op.Usage = usage
	op.Autocomplete = true
	op.ParseHelp = true
	verSlice := []string(version)
	if len(verSlice) > 0 {
		op.ParseVersion = true
		op.Version = verSlice[0]
	} else {
		op.ParseVersion = false
 	}
	return op
}
