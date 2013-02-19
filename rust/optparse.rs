// Public Domain (-) 2012-2013 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

//= The optparse module provides support for handling command line options.
//
// Just specify the options that you want your program to take and `optparse`
// will automatically generate usage messages, validate arguments, type cast,
// provide auto-completion support and even handle sub-commands.
//
// To start, initialise a handler with `usage` and `version` info for your
// command line program, e.g.
//
//     let opts = optparse::new(
//                  "Usage: jsonprint [options] <files>", "jsonprint 0.1"
//                  );
//
// Then add the various options that you want it to handle by specifying the
// `flags` and a brief help `usage` for the specific option type, e.g.
//
//     let indent = opts.int(["-i", "--indent"], "number of spaces to use for indentation");
//     let output = opts.str(["-o", "--output"], "the path to write the output to");
//     let pretty = opts.bool(["-p", "--pretty"], "pretty-print the generated output");
//
// And invoke the `optparse` machinery by calling the `parse()` method. This
// will return any left-over parameters as a string slice, e.g.
//
//     let files = opts.parse();
//
// You can then dereference the option variables to get their parsed values, e.g.
//
//     io::println(fmt!("Writing to the file: %s", *output));
//
//     if *pretty {
//         // pretty-print the output ...
//     }
//
// The builtin option types correspond to Rust's core data types and can be one
// of `bool`, `int`, 'i64', `str`, 'uint' or 'u64'. They default to the "zero"
// value for their type, i.e. `false`, `0` or `""`.
//
// You can override the default by specifying a third parameter when you
// initialise the options, e.g.
//
//     let indent = opts.int(
//         ["-i", "--indent"], "number of spaces to use for indentation", 4
//         );
//
// You can also specify that an option is required by calling `required()`
// before initialising the option, e.g.
//
//     let output = opts.required().str(["-o", "--output", "..."]);
//
// In order to be user-friendly, `-h`/`--help` and `-v`/`--version` options are
// automatically added when you call `parse()`. These use the `usage`, `version`
// and option `usage` parameters you specified to auto-generate helpful output,
// e.g.
//
// ```
// $ jsonprint --help
// Usage: jsonprint [options] <files>
//
//   -h, --help        show this help message and exit
//   -v, --version     show program's version number and exit
// ```
//
// This is often the desired behaviour and follows established best practice for
// command line tools. Sometimes though, especially when serving non-English
// locales, you might want to disable this behaviour:
//
//     opts.handle_help = false;
//     opts.handle_version = false;
//
// If you want, you can still add custom options and handle it yourself with the
// utility `print_help()` and `print_version()` methods, e.g.
//
//     let help = opts.bool(["-h", "--hilfe"], "diese hilfe anzeigen und beenden");
//
//     opts.handle_help = false;
//     opts.parse();
//
//     if *help {
//         opts.print_help();
//         os.set_exit_status(0);
//     }
//
// You can override the default parsing of the command line `os.args()` by
// passing an explicit string slice to `parse()`, e.g.
//
//     let files = opts.parse(["jsonprint", "items.json"]);
//
// Following convention, all arguments following a standalone `--` parameter are
// returned without being parsed as options, e.g in the following, `--pretty` is
// not treated as an option flag:
//
//     let files = opts.parse(["jsonprint", "items.json", "--", "--pretty"]);
//

// You can define custom option types by implementing the `Value` trait and
// specifying it in the explicit `option` initialiser, e.g. to aggregate
// multiple `--server` values into a vector, you might do something like:
//
//     type Servers = ~[~str];
//
//     impl Servers: optparse::Value {
//         fn set(arg: ~str) Option {
//             if arg.len() == 0 {
//                 return Some(~"server value cannot be empty");
//             }
//             self.push(arg);
//             return None;
//         }
//     }
//
//     let servers: Servers = ~[];
//
//     opts.option(["-s", "--server"], "address of upstream server", servers)

// auto-complete
// config file

use core::option::{Option, None, Some};

pub trait Completer {
    fn complete(self) -> ~[~str];
}

impl Completer for ~[~str] {
    fn complete(self) -> ~[~str] {
        copy self
    }
}

pub fn list_complete(items: &[&str]) -> Option<@Completer> {
    Some(items.map(|i| i.to_str()) as @Completer)
}

// The Value trait.
pub trait Value {
    fn set(&mut self, &str) -> Option<~str>;
    fn string(self) -> ~str;
}

type Bool = bool;

impl Value for Bool {
    fn set(&mut self, s: &str) -> Option<~str> {
        if s.len() != 0 {
            *self = true;
        }
        None
    }
    fn string(self) -> ~str {
        fmt!("%?", self)
    }
}

type I64 = i64;

impl Value for I64 {
    fn set(&mut self, s: &str) -> Option<~str> {
        match i64::from_str(s) {
            Some(x) => {
                *self = x;
                None
            }
            None => Some(fmt!("strconv: unable to convert %s to an i64", s))
        }
    }
    fn string(self) -> ~str {
        fmt!("%?", self)
    }
}

type Int = int;

impl Value for Int {
    fn set(&mut self, s: &str) -> Option<~str> {
        match int::from_str(s) {
            Some(x) => {
                *self = x;
                None
            }
            None => Some(fmt!("strconv: unable to convert %s to an int", s))
        }
    }
    fn string(self) -> ~str {
        fmt!("%d", self)
    }
}

type Str = @str;

impl Value for Str {
    fn set(&mut self, s: &str) -> Option<~str> {
        *self = s.to_managed();
        None
    }
    fn string(self) -> ~str {
        fmt!("\"%s\"", self)
    }
}

type U64 = u64;

impl Value for U64 {
    fn set(&mut self, s: &str) -> Option<~str> {
        match u64::from_str(s) {
            Some(x) => {
                *self = x;
                None
            }
            None => Some(fmt!("strconv: unable to convert %s to a u64", s))
        }
    }
    fn string(self) -> ~str {
        fmt!("%?", self)
    }
}

type Uint = uint;

impl Value for Uint {
    fn set(&mut self, s: &str) -> Option<~str> {
        match uint::from_str(s) {
            Some(x) => {
                *self = x;
                None
            }
            None => Some(fmt!("strconv: unable to convert %s to a uint", s))
        }
    }
    fn string(self) -> ~str {
        fmt!("%?", self)
    }
}

pub type ErrPrinter = &fn(&str, &str);

fn default_arg_required(prog: &str, arg: &str) {
    io::println(fmt!("%s: error: %s option requires an argument", prog, arg))
}

fn default_no_such_option(prog: &str, arg: &str) {
    io::println(fmt!("%s: error: no such option: %s", prog, arg))
}

fn default_required(prog: &str, arg: &str) {
    io::println(fmt!("%s: error: required: %s", prog, arg))
}

fn exit(print: &const ErrPrinter, prog: &str, arg: &str) {
    (*print)(prog, arg);
    unsafe {
        libc::exit(1);
    }
}

fn get_prog_name(arg: &str) -> ~str {
    match path::Path(arg).filename() {
        Some(x) => copy x,
        None => str::from_slice(arg)
    }
}

pub struct OptionParser {
    priv mut _help_added: bool,
    priv mut _version_added: bool,
    mut handle_help: bool,
    mut handle_version: bool,
    mut completer: Option<@Completer>,
    mut err_arg_required: ErrPrinter,
    mut err_no_such_option: ErrPrinter,
    mut err_required: ErrPrinter,
    mut next_completer: Option<@Completer>,
    mut next_dest: ~str,
    mut next_required: bool,
    mut opts: ~[@OptValue],
    mut print_defaults: bool,
    mut usage: ~str,
    mut version: ~str,
}

impl OptionParser {

    fn bool(&self, flags: &[&str], usage: &str) -> @mut bool {
        self._bool(flags, usage)
    }

    fn bool(&self, flag: &str, usage: &str) -> @mut bool {
        self._bool(~[flag], usage)
    }

    priv fn _bool(&self, flags: &[&str], usage: &str) -> @mut bool {
        let mut val = @mut false;
        self._option(flags, usage, true, val as Value);
        val
    }

    fn dest(&self, name: &str) -> &self/OptionParser {
        self.next_dest = str::from_slice(name);
        return self;
    }

    fn i64(&self, flags: &[&str], usage: &str) -> @mut i64 {
        self._i64(flags, usage, 0)
    }

    fn i64(&self, flag: &str, usage: &str) -> @mut i64 {
        self._i64(~[flag], usage, 0)
    }

    fn i64(&self, flags: &[&str], usage: &str, default: i64) -> @mut i64 {
        self._i64(flags, usage, default)
    }

    fn i64(&self, flag: &str, usage: &str, default: i64) -> @mut i64 {
        self._i64(~[flag], usage, default)
    }

    priv fn _i64(&self, flags: &[&str], usage: &str, default: i64) -> @mut i64 {
        let mut val = @mut default;
        self._option(flags, usage, false, val as Value);
        val
    }

    fn int(&self, flags: &[&str], usage: &str) -> @mut int {
        self._int(flags, usage, 0)
    }

    fn int(&self, flag: &str, usage: &str) -> @mut int {
        self._int(~[flag], usage, 0)
    }

    fn int(&self, flags: &[&str], usage: &str, default: int) -> @mut int {
        self._int(flags, usage, default)
    }

    fn int(&self, flag: &str, usage: &str, default: int) -> @mut int {
        self._int(~[flag], usage, default)
    }

    priv fn _int(&self, flags: &[&str], usage: &str, default: int) -> @mut int {
        let mut val = @mut default;
        self._option(flags, usage, false, val as Value);
        val
    }

    fn str(&self, flags: &[&str], usage: &str) -> @mut @str {
        self._str(flags, usage, "")
    }

    fn str(&self, flag: &str, usage: &str) -> @mut @str {
        self._str(~[flag], usage, "")
    }

    fn str(&self, flags: &[&str], usage: &str, default: &str) -> @mut @str {
        self._str(flags, usage, default)
    }

    fn str(&self, flag: &str, usage: &str, default: &str) -> @mut @str {
        self._str(~[flag], usage, default)
    }

    priv fn _str(&self, flags: &[&str], usage: &str, default: &str) -> @mut @str {
        let mut val = @mut default.to_managed();
        self._option(flags, usage, false, val as Value);
        val
    }

    fn option(&self, flags: &[&str], usage: &str, value: @Value) {
        self._option(flags, usage, false, value)
    }

    fn option(&self, flag: &str, usage: &str, value: @Value) {
        self._option(~[flag], usage, false, value)
    }

    priv fn _option(&self, flags: &[&str], usage: &str, implicit: bool, value: @Value) {
        let mut flag_config = ~"";
        let mut flag_long = ~"";
        let mut flag_short = ~"";
        for flags.each |f| {
            let flag = str::from_slice(*f);
            if flag.starts_with("--") {
                flag_long = flag;
            } else if flag.starts_with("-") {
                flag_short = flag;
            } else {
                flag_long = ~"--" + flag;
                flag_config = flag;
                // fail fmt!("optparse: invalid flag: %s", flag);
            }
        }
        self.opts.push(@OptValue{
            completer: self.next_completer,
            defined: false,
            dest: copy self.next_dest,
            flag_long: flag_long,
            flag_short: flag_short,
            implicit: implicit,
            usage: str::from_slice(usage),
            required_flag: if self.next_required && flag_config.len() != 0 {
                false
            } else {
                self.next_required
            },
            required_conf: self.next_required,
            value: value,
            flag_config: flag_config,
        });
        self.next_completer = None;
        self.next_dest = ~"";
        self.next_required = false;
    }

    fn parse(&self) -> ~[~str] {
        self._parse(os::args())
    }

    fn parse(&self, args: &[~str]) -> ~[~str] {
        self._parse(args)
    }

    priv fn _parse(&self, args: &[~str]) -> ~[~str] {
        if self.handle_help && !self._help_added {
            self.bool(["-h", "--help"], "show this help and exit");
            self._help_added = true;
        }
        if self.handle_version && !self._version_added {
            self.bool(["-v", "--version"], "show the version and exit");
            self._version_added = true;
        }
        let retargs: ~[~str] = ~[];
        let arglen = args.len();
        if arglen < 1 {
            return retargs;
        }
        let optslen = self.opts.len();
        let mut i = 0;
        while i < arglen {
            let arg = copy args[i];
            let mut j = 0;
            while j < optslen {
                let opt = self.opts[j];
                if opt.flag_long == arg {
                    if opt.implicit {
                        opt.value.set(arg);
                    } else if arglen > i + 1 {
                        opt.value.set(args[i+1]);
                    }
                    break;
                }
                j += 1;
            }
            i += 1;
        }
        // exit(&mut self.err_no_such_option, get_prog_name(args[0]), "name");
        retargs
    }

    fn print_config_file(self, name: &str) {
        io::println(name)
    }

    fn required(&self) -> &self/OptionParser {
        self.next_required = true;
        return self;
    }

    fn set_main_completer(&self, completer: Option<@Completer>) -> &self/OptionParser {
        self.completer = completer;
        return self;
    }

    fn u64(&self, flags: &[&str], usage: &str) -> @mut u64 {
        self._u64(flags, usage, 0)
    }

    fn u64(&self, flag: &str, usage: &str) -> @mut u64 {
        self._u64(~[flag], usage, 0)
    }

    fn u64(&self, flags: &[&str], usage: &str, default: u64) -> @mut u64 {
        self._u64(flags, usage, default)
    }

    fn u64(&self, flag: &str, usage: &str, default: u64) -> @mut u64 {
        self._u64(~[flag], usage, default)
    }

    priv fn _u64(&self, flags: &[&str], usage: &str, default: u64) -> @mut u64 {
        let mut val = @mut default;
        self._option(flags, usage, false, val as Value);
        val
    }

    fn uint(&self, flags: &[&str], usage: &str) -> @mut uint {
        self._uint(flags, usage, 0)
    }

    fn uint(&self, flag: &str, usage: &str) -> @mut uint {
        self._uint(~[flag], usage, 0)
    }

    fn uint(&self, flags: &[&str], usage: &str, default: uint) -> @mut uint {
        self._uint(flags, usage, default)
    }

    fn uint(&self, flag: &str, usage: &str, default: uint) -> @mut uint {
        self._uint(~[flag], usage, default)
    }

    priv fn _uint(&self, flags: &[&str], usage: &str, default: uint) -> @mut uint {
        let mut val = @mut default;
        self._option(flags, usage, false, val as Value);
        val
    }

    fn with_completer(&self, completer: Option<@Completer>) -> &self/OptionParser {
        self.next_completer = completer;
        return self;
    }

}

struct OptValue {
    completer: Option<@Completer>,
    mut defined: bool,
    dest: ~str,
    flag_config: ~str,
    flag_long: ~str,
    flag_short: ~str,
    implicit: bool,
    usage: ~str,
    required_conf: bool,
    required_flag: bool,
    mut value: @Value
}

pub fn new(usage: ~str, version: ~str) -> ~OptionParser {
    ~OptionParser{
        _help_added: false,
        _version_added: false,
        completer: None,
        err_arg_required: default_arg_required,
        err_no_such_option: default_no_such_option,
        err_required: default_required,
        next_completer: None,
        next_dest: ~"",
        next_required: false,
        opts: ~[],
        handle_help: true,
        handle_version: if version == ~"" {
            false
        } else {
            true
        },
        print_defaults: false,
        usage: copy usage,
        version: copy version,
    }
}