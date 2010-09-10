// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

// The optparse package provides utility functions for the parsing and
// autocompletion of command line arguments.
package optparse

import (
	"amp/runtime"
	"amp/slice"
	"amp/yaml"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type option struct {
	configflag     string
	shortflag      string
	longflag       string
	dest           string
	usage          string
	requiredFlag   bool
	requiredConfig bool
	defined        bool
	valueType      string
	intValue       *int
	boolValue      *bool
	stringValue    *string
}

type OptionParser struct {
	Usage          string
	Version        string
	ParseHelp      bool
	ParseVersion   bool
	Autocomplete   bool
	options        []*option
	config2options map[string]*option
	configflags    []string
	short2options  map[string]*option
	shortflags     []string
	long2options   map[string]*option
	longflags      []string
	helpAdded      bool
	versionAdded   bool
}

func (opt *option) String() (output string) {
	output = "  "
	if opt.configflag != "" {
		output += opt.configflag
		output += ": "
	} else {
		if opt.shortflag != "" {
			output += opt.shortflag
			if opt.longflag != "" {
				output += ", "
			}
		}
		if opt.longflag != "" {
			output += opt.longflag
		}
		if opt.dest != "" {
			output += " " + opt.dest
		}
	}
	length := len(output)
	if length >= 19 {
		output += "\n                    "
	} else {
		padding := make([]byte, 20-length)
		for i, _ := range padding {
			padding[i] = 32
		}
		output += string(padding)
	}
	output += opt.usage
	output += "\n"
	return
}

func (op *OptionParser) computeFlags(flags []string, opt *option) (configflag, shortflag, longflag string) {
	for _, flag := range flags {
		if strings.HasPrefix(flag, "--") {
			longflag = flag
			op.long2options[longflag] = opt
			slice.AppendString(&op.longflags, longflag)
		} else if strings.HasPrefix(flag, "-") {
			shortflag = flag
			op.short2options[shortflag] = opt
			slice.AppendString(&op.shortflags, shortflag)
		} else if strings.HasSuffix(flag, ":") {
			configflag = flag[0 : len(flag)-1]
			op.config2options[configflag] = opt
			slice.AppendString(&op.configflags, configflag)
		} else {
			longflag = flag
			op.long2options[longflag] = opt
			slice.AppendString(&op.longflags, longflag)
		}
	}
	return
}

func (op *OptionParser) Default(flags []string, usage string, displayDest bool, info ...interface{}) (opt *option) {
	opt = &option{}
	opt.usage = usage
	opt.configflag, opt.shortflag, opt.longflag = op.computeFlags(flags, opt)
	var required bool
	var dest string
	for _, prop := range info {
		switch prop := prop.(type) {
		case bool:
			required = prop
		case string:
			dest = prop
		}
	}
	if required {
		if opt.configflag == "" {
			opt.requiredFlag = true
		} else {
			opt.requiredConfig = true
		}
	}
	if displayDest {
		if dest != "" {
			opt.dest = dest
		} else {
			if opt.longflag != "" {
				opt.dest = strings.ToUpper(strings.TrimLeft(opt.longflag, "-"))
			} else {
				opt.dest = strings.ToUpper(strings.TrimLeft(opt.shortflag, "-"))
			}
		}
	}
	length := len(op.options)
	if cap(op.options) == length {
		temp := make([]*option, length, 2*(length+1))
		for idx, item := range op.options {
			temp[idx] = item
		}
		op.options = temp
	}
	op.options = op.options[0 : length+1]
	op.options[length] = opt
	return
}

func (op *OptionParser) Int(flags []string, value int, usage string, info ...interface{}) (result *int) {
	opt := op.Default(flags, usage, true, info)
	opt.valueType = "int"
	opt.intValue = &value
	return &value
}

func (op *OptionParser) String(flags []string, value string, usage string, info ...interface{}) (result *string) {
	opt := op.Default(flags, usage, true, info)
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

func (op *OptionParser) IntConfig(flag string, value int, usage string, info ...interface{}) (result *int) {
	opt := op.Default([]string{flag + ":", "--" + flag}, usage, false, info)
	opt.valueType = "int"
	opt.intValue = &value
	return &value
}

func (op *OptionParser) StringConfig(flag string, value string, usage string, info ...interface{}) (result *string) {
	opt := op.Default([]string{flag + ":", "--" + flag}, usage, false, info)
	opt.valueType = "string"
	opt.stringValue = &value
	return &value
}

func (op *OptionParser) BoolConfig(flag string, value bool, usage string) (result *bool) {
	opt := op.Default([]string{flag + ":", "--" + flag}, usage, false)
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

	// Command-line auto-completion support.
	autocomplete := os.Getenv("OPTPARSE_AUTO_COMPLETE")
	if autocomplete != "" {

		compWords := os.Getenv("COMP_WORDS")
		if compWords == "" {
			// zsh's bashcompinit does not pass COMP_WORDS, replace with
			// COMP_LINE for now...
			compWords = os.Getenv("COMP_LINE")
			if compWords == "" {
				runtime.Exit(1)
			}
		}
		compWordsList := strings.Split(compWords, " ", -1)
		compLine := os.Getenv("COMP_LINE")
		compPoint, err := strconv.Atoi(os.Getenv("COMP_POINT"))
		if err != nil {
			runtime.Exit(1)
		}
		compWord, err := strconv.Atoi(os.Getenv("COMP_CWORD"))
		if err != nil {
			runtime.Exit(1)
		}

		prefix := ""
		if compWord > 0 {
			if compWord < len(compWordsList) {
				prefix = compWordsList[compWord]
			}
		}

		// At some point in the future, make autocompletion customisable per
		// option flag and, at that point, make use of these variables.
		_ = compLine
		_ = compPoint

		// Pass to the shell completion if the previous word was a flag
		// expecting some parameter.
		if (compWord - 1) > 0 {
			prev := compWordsList[compWord-1]
			if strings.HasPrefix(prev, "--") {
				opt, ok := op.long2options[prev]
				if ok {
					if opt.dest != "" {
						runtime.Exit(1)
					}
				}
			} else if strings.HasPrefix(prev, "-") {
				opt, ok := op.short2options[prev]
				if ok {
					if opt.dest != "" {
						runtime.Exit(1)
					}
				}
			}
		}

		completions := make([]string, 0)
		for flag, _ := range op.long2options {
			if strings.HasPrefix(flag, prefix) {
				slice.AppendString(&completions, flag)
			}
		}

		for flag, _ := range op.short2options {
			if strings.HasPrefix(flag, prefix) {
				slice.AppendString(&completions, flag)
			}
		}

		fmt.Print(strings.Join(completions, " "))
		runtime.Exit(1)

	}

	if argLength == 0 {
		return
	}

	var opt *option
	var ok bool

	idx := 1

	for {
		arg := args[idx]
		noOpt := true
		if strings.HasPrefix(arg, "--") {
			opt, ok = op.long2options[arg]
			if ok {
				noOpt = false
			}
		} else if strings.HasPrefix(arg, "-") {
			opt, ok = op.short2options[arg]
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
			runtime.Error("%s: error: no such option: %s\n", args[0], arg)
		}
		if opt.dest != "" {
			if idx == argLength {
				runtime.Error("%s: error: %s option requires an argument\n", args[0], arg)
			}
		}
		if opt.valueType == "bool" {
			if opt.longflag == "--help" && op.ParseHelp {
				op.PrintUsage()
				runtime.Exit(1)
			} else if opt.longflag == "--version" && op.ParseVersion {
				fmt.Printf("%s\n", op.Version)
				runtime.Exit(0)
			}
			*opt.boolValue = true
			opt.defined = true
			idx += 1
		} else if opt.valueType == "string" {
			if idx == argLength {
				runtime.Error("%s: error: no value specified for %s\n", args[0], arg)
			}
			*opt.stringValue = args[idx+1]
			opt.defined = true
			idx += 2
		} else if opt.valueType == "int" {
			if idx == argLength {
				runtime.Error("%s: error: no value specified for %s\n", args[0], arg)
			}
			intValue, err := strconv.Atoi(args[idx+1])
			if err != nil {
				runtime.Error("%s: error: couldn't convert %s value '%s' to an integer\n", args[0], arg, args[idx+1])
			}
			*opt.intValue = intValue
			opt.defined = true
			idx += 2
		}
		if idx > argLength {
			break
		}
	}

	for _, opt := range op.options {
		if opt.requiredFlag && !opt.defined {
			runtime.Error("%s: error: required: %s", args[0], opt)
		}
	}

	return

}

func (op *OptionParser) ParseConfig(filename string, args []string) (err os.Error) {

	data, err := yaml.ParseFile(filename)
	if err != nil {
		return err
	}

	for config, opt := range op.config2options {
		if opt.defined {
			continue
		}
		value, ok := data[config]
		if !ok {
			if opt.requiredConfig {
				runtime.Error("%s: error: required: %s", args[0], opt)
			} else {
				continue
			}
		}
		if opt.valueType == "bool" {
			if value == "true" || value == "on" || value == "yes" {
				*opt.boolValue = true
			} else if value == "false" || value == "off" || value == "no" {
				*opt.boolValue = false
			} else {
				runtime.Error("%s: error: invalid boolean value for %s: %q\n", args[0], config, value)
			}
		} else if opt.valueType == "string" {
			*opt.stringValue = value
		} else if opt.valueType == "int" {
			intValue, err := strconv.Atoi(value)
			if err != nil {
				runtime.Error("%s: error: couldn't convert the %s value %q to an integer\n", args[0], config, value)
			}
			*opt.intValue = intValue
		}
	}

	return nil

}

func (op *OptionParser) PrintUsage() {
	fmt.Print(op.Usage)
	if len(op.configflags) > 0 {
		fmt.Print("\nConfig File Options:\n")
	}
	for _, opt := range op.options {
		if opt.configflag != "" {
			fmt.Printf("%v", opt)
		}
	}
	if len(op.options) > 0 {
		fmt.Print("\nOptions:\n")
	}
	for _, opt := range op.options {
		if opt.configflag == "" {
			fmt.Printf("%v", opt)
		}
	}
}

// Utility constructor.
func Parser(usage string, version ...string) (op *OptionParser) {
	op = &OptionParser{}
	op.long2options = make(map[string]*option)
	op.short2options = make(map[string]*option)
	op.config2options = make(map[string]*option)
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
