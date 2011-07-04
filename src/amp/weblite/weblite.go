// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package weblite

import (
	"amp/runtime"
	"amp/optparse"
	"fmt"
	"os"
	"path/filepath"
)

type Response interface{}
type Handler func(ctx *Context) Response

// -----------------------------------------------------------------------------
// Context
// -----------------------------------------------------------------------------

type Context struct {
	Args   []string
	Params map[string]string
	Host   string
	Method string
}

func (ctx *Context) Error(message string) {
}

func (ctx *Context) Redirect(url string) {
}

func (ctx *Context) GetCookie(key string) (value string) {
	return
}

func (ctx *Context) SetCookie(key string, value string) {
	return
}

func (ctx *Context) GetUser() (user string) {
	return
}

// -----------------------------------------------------------------------------
// Service
// -----------------------------------------------------------------------------

type Service struct {
	Handler   Handler
	Renderers []interface{}
	admin     bool
	auth      bool
	stream    bool
	xsrf      bool
}

func (service *Service) AdminOnly() *Service {
	service.admin = true
	service.auth = true
	return service
}

func (service *Service) AuthOnly() *Service {
	service.auth = true
	return service
}

func (service *Service) DisableXSRF() *Service {
	service.xsrf = true
	return service
}

func (service *Service) Stream() *Service {
	service.stream = true
	return service
}

// -----------------------------------------------------------------------------
// Application
// -----------------------------------------------------------------------------

type AppConfig struct {
	Debug               *bool
	GenConfig           *bool
	Host                *string
	Port                *int
	FrontendHost        *string
	FrontendPort        *int
	FrontendConnections *int
	FrontendTLS         *bool
	LogDirectory        *string
	LogRotate           *string
	NoConsoleLog        *bool
	RunDirectory        *string
	TemplatesDirectory  *string
}

type Application struct {
	Name     string
	Config   *AppConfig
	Debug    bool
	Opts     *optparse.OptionParser
	Path     string
	Registry map[string]*Service
}

func (app *Application) Register(path string, handler Handler, renderers ...interface{}) (service *Service) {
	service = &Service{Handler: handler, Renderers: renderers}
	app.Registry[path] = service
	return service
}

func (app *Application) ParseOpts() {

	opts := app.Opts
	conf := app.Config

	conf.Host = opts.StringConfig("weblite-host", "localhost",
		"the host to bind this weblite server to [localhost]")

	conf.Port = opts.IntConfig("weblite-port", 8080,
		"the port to bind this weblite server to [8080]")

	conf.FrontendHost = opts.StringConfig("frontend-host", "localhost",
		"the frontend host to connect to [localhost]")

	conf.FrontendPort = opts.IntConfig("frontend-port", 9040,
		"the frontend port to connect to [9040]")

	conf.FrontendConnections = opts.IntConfig("frontend-cxns", 5,
		"the number of frontend connections to maintain [5]")

	conf.FrontendTLS = opts.BoolConfig("frontend-tls", false,
		"use TLS when connecting to the frontend [false]")

	conf.RunDirectory = opts.StringConfig("run-dir", "run",
		"the path to the run directory to store locks, pid files, etc. [run]")

	conf.TemplatesDirectory = opts.StringConfig("templates-dir", "templates",
		"the path to the templates directory [templates]")

	conf.LogDirectory = opts.StringConfig("log-dir", "log",
		"the path to the log directory [log]")

	conf.LogRotate = opts.StringConfig("log-rotate", "never",
		"specify one of 'hourly', 'daily' or 'never' [never]")

	conf.NoConsoleLog = opts.BoolConfig("no-console-log", false,
		"disable server requests being logged to the console [false]")

	extraConfig := opts.StringConfig("extra-config", "",
		"path to a YAML config file with additional options")

	// Parse the command line options.
	os.Args[0] = app.Name
	args := opts.Parse(os.Args)

	// Print the default YAML config file if the ``-g`` flag was specified.
	if *conf.GenConfig {
		opts.PrintDefaultConfigFile()
		runtime.Exit(0)
	}

	var instanceDirectory string

	// Assume the parent directory of the config as the instance directory.
	if len(args) >= 1 {
		if args[0] == "help" {
			opts.PrintUsage()
			runtime.Exit(0)
		}
		configPath, err := filepath.Abs(filepath.Clean(args[0]))
		if err != nil {
			runtime.StandardError(err)
		}
		err = opts.ParseConfig(configPath, os.Args)
		if err != nil {
			runtime.StandardError(err)
		}
		instanceDirectory, _ = filepath.Split(configPath)
	} else {
		opts.PrintUsage()
		runtime.Exit(0)
	}

	// Load the extra config file with additional options if one has been
	// specified.
	if *extraConfig != "" {
		extraConfigPath, err := filepath.Abs(filepath.Clean(*extraConfig))
		if err != nil {
			runtime.StandardError(err)
		}
		extraConfigPath = runtime.JoinPath(instanceDirectory, extraConfigPath)
		err = opts.ParseConfig(extraConfigPath, os.Args)
		if err != nil {
			runtime.StandardError(err)
		}
	}

	app.Path = instanceDirectory

}

func (app *Application) Init(env map[string]string) {

	conf := app.Config

	// Set the debug mode flag if the ``-d`` flag was specified.
	app.Debug = *conf.Debug

	// Create the log directory if it doesn't exist.
	logPath := runtime.JoinPath(app.Path, *conf.LogDirectory)
	err := os.MkdirAll(logPath, 0755)
	if err != nil {
		runtime.StandardError(err)
	}

	// Create the run directory if it doesn't exist.
	runPath := runtime.JoinPath(app.Path, *conf.RunDirectory)
	err = os.MkdirAll(runPath, 0755)
	if err != nil {
		runtime.StandardError(err)
	}

	// Initialise the process-related resources.
	runtime.Init()
	runtime.InitProcess(app.Name, runPath)

	fmt.Printf("Running %s on %s:%d\n",
		app.Name, *app.Config.Host, *app.Config.Port)

	app.HandleRequests()

}

func (app *Application) HandleRequests() {
	for {

	}
}

func (app *Application) HandleRequest(path string) {
	ctx := &Context{}
	service := app.Registry[path]
	resp := service.Handler(ctx)
	renderers := service.Renderers
	var output map[string]interface{}
	switch typ := resp.(type) {
	case map[string]interface{}:
		output = resp.(map[string]interface{})
	case int:
		output = make(map[string]interface{})
		output["output"] = resp.(int)
	case string:
		output = make(map[string]interface{})
		output["output"] = resp.(string)
	case nil:
		output = make(map[string]interface{})
		output["output"] = nil
	}
	if len(renderers) != 0 {
	}
}

func Noop(input interface{}) {

}

// -----------------------------------------------------------------------------
// Constructor
// -----------------------------------------------------------------------------

func App(name, version string) (*Application, *optparse.OptionParser) {

	// Initialise the options parser.
	opts := optparse.Parser(
		fmt.Sprintf("Usage: %s <config.yaml> [options]\n", name),
		fmt.Sprintf("%s %s", name, version))

	conf := &AppConfig{}

	// Create the Application object.
	app := &Application{
		Config: conf,
		Name:   name,
		Opts:   opts,
	}

	app.Registry = make(map[string]*Service)

	// Setup default command line options.
	conf.Debug = opts.Bool([]string{"-d", "--debug"}, false,
		"enable debug mode")

	conf.GenConfig = opts.Bool([]string{"-g", "--gen-config"}, false,
		"show the default yaml config")

	return app, opts

}
