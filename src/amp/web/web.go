// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package web

import (
	"amp/optparse"
	"amp/runtime"
	"amp/structure"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"template"
)

type Response interface{}
type Handler func(*Context) Response
type Renderer func(*Context, Response) Response

// -----------------------------------------------------------------------------
// Context
// -----------------------------------------------------------------------------

type Context struct {
	admin bool
	user  string
	xsrf  string

	AjaxRequest     bool
	Args            []string
	Kwargs          map[string]string
	Env             map[string]interface{}
	Host            string
	JSONCallback    string
	Method          string
	ResponseHeaders map[string]string
	StopRendering   bool
}

func (ctx *Context) Error(message string) {
	panic(os.NewError(message))
}

func (ctx *Context) SetResponseStatus(status int) {
}

func (ctx *Context) NotFound() {
}

func (ctx *Context) Redirect(url string) {
}

func (ctx *Context) InternalRedirect(url string) {
}

func (ctx *Context) GetCookie(key string) (value string) {
	return
}

func (ctx *Context) SetCookie(key string, value string) {
	return
}

func (ctx *Context) ExpireCookie(key string) {
	return
}

func (ctx *Context) GetUser() string {
	return ""
}

func (ctx *Context) IsAdmin() bool {
	return false
}

func (ctx *Context) XSRF() string {
	return ""
}

func (ctx *Context) DontCache() {
}

func (ctx *Context) ComputeURL() string {
	return ""
}

func (ctx *Context) ComputeHostURL() string {
	return ""
}

// -----------------------------------------------------------------------------
// Service
// -----------------------------------------------------------------------------

type Service struct {
	admin    bool
	auth     bool
	stream   bool
	xsrf     bool
	wildcard bool

	// A service is effectively a handler and a list of renderers registered for
	// a given path.
	Handler   Handler
	Renderers []Renderer
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

func (service *Service) XSRF() *Service {
	service.xsrf = true
	return service
}

func (service *Service) Stream() *Service {
	service.stream = true
	return service
}

// -----------------------------------------------------------------------------
// Templating Provider
// -----------------------------------------------------------------------------

type TemplatingProvider interface {
	Execute(string, interface{}) []byte
	GenerateRenderer(string) Renderer
}

type Templating struct {
	Debug        bool
	Cache        map[string]*template.Template
	Directory    string
	FormatterMap template.FormatterMap
}

func (templating *Templating) Load(path string) (tpl *template.Template) {
	tpl = template.New(templating.FormatterMap)
	return
}

func (templating *Templating) Init(assetManifest string) {
	templating.FormatterMap = template.FormatterMap{
		"":     template.StringFormatter,
		"str":  template.StringFormatter,
		"html": template.HTMLFormatter,
		"static": func(w io.Writer, format string, value ...interface{}) {
			if templating.Debug {
			} else {
			}
		},
	}
}

func (templating *Templating) GenerateRenderer(path string) Renderer {
	return func(ctx *Context, input Response) (resp Response) {
		return templating.Execute(path, input)
	}
}

func (templating *Templating) Execute(path string, data interface{}) []byte {
	template := templating.Load(path)
	buffer := &bytes.Buffer{}
	template.Execute(buffer, data)
	return buffer.Bytes()
}

// -----------------------------------------------------------------------------
// Standard Application Config
// -----------------------------------------------------------------------------

type AppConfig struct {
	assetManifest       *string
	debug               *bool
	genConfig           *bool
	host                *string
	port                *int
	errorDirectory      *string
	frontendHost        *string
	frontendPort        *int
	frontendConnections *int
	frontendTLS         *bool
	logDirectory        *string
	logRotate           *string
	noConsoleLog        *bool
	runDirectory        *string
	templatesDirectory  *string
}

// -----------------------------------------------------------------------------
// Application
// -----------------------------------------------------------------------------

type Application struct {
	Name       string
	Config     *AppConfig
	Debug      bool
	Hooks      []func()
	Opts       *optparse.OptionParser
	Path       string
	Services   *structure.PrefixTree
	Templating TemplatingProvider
}

func (app *Application) Register(path string, handler Handler, renderers ...interface{}) (service *Service) {
	list := make([]Renderer, len(renderers))
	for i, r := range renderers {
		switch r.(type) {
		case string:
			if app.Templating == nil {
				panic("Templating used without being enabled: " + r.(string))
			}
			list[i] = app.Templating.GenerateRenderer(r.(string))
		case Renderer:
			list[i] = r.(Renderer)
		default:
			panic("Unknown renderer type!")
		}
	}
	service = &Service{Handler: handler, Renderers: list}
	if path[len(path)-1] == '*' {
		service.wildcard = true
		path = path[:len(path)-1]
	}
	app.Services.Insert(path, service)
	return service
}

func (app *Application) RegisterTemplatingProvider(provider TemplatingProvider) {
	app.Templating = provider
}

func (app *Application) RegisterHook(hook func()) {
	app.Hooks = append(app.Hooks, hook)
}

func (app *Application) ParseOpts() {

	opts := app.Opts
	conf := app.Config

	conf.host = opts.StringConfig("weblite-host", "localhost",
		"the host to bind this weblite server to [localhost]")

	conf.port = opts.IntConfig("weblite-port", 8080,
		"the port to bind this weblite server to [8080]")

	conf.frontendHost = opts.StringConfig("frontend-host", "localhost",
		"the frontend host to connect to [localhost]")

	conf.frontendPort = opts.IntConfig("frontend-port", 9040,
		"the frontend port to connect to [9040]")

	conf.frontendConnections = opts.IntConfig("frontend-cxns", 5,
		"the number of frontend connections to maintain [5]")

	conf.frontendTLS = opts.BoolConfig("frontend-tls", false,
		"use TLS when connecting to the frontend [false]")

	conf.runDirectory = opts.StringConfig("run-dir", "run",
		"the path to the run directory to store locks, pid files, etc. [run]")

	conf.errorDirectory = opts.StringConfig("error-dir", "error",
		"the path to the error templates directory [error]")

	conf.logDirectory = opts.StringConfig("log-dir", "log",
		"the path to the log directory [log]")

	conf.logRotate = opts.StringConfig("log-rotate", "never",
		"specify one of 'hourly', 'daily' or 'never' [never]")

	conf.noConsoleLog = opts.BoolConfig("no-console-log", false,
		"disable server requests being logged to the console [false]")

	extraConfig := opts.StringConfig("extra-config", "",
		"path to a YAML config file with additional options")

	// Parse the command line options.
	os.Args[0] = app.Name
	args := opts.Parse(os.Args)

	// Print the default YAML config file if the ``-g`` flag was specified.
	if *conf.genConfig {
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

	// Set the debug mode flag if the ``-d`` flag was specified.
	app.Debug = *conf.debug
	app.Path = instanceDirectory

	for _, hook := range app.Hooks {
		hook()
	}

}

func (app *Application) EnableTemplating() *Application {

	templating := &Templating{}
	templatesDirectory := app.Opts.StringConfig("templates-dir", "templates",
		"the path to the templates directory [templates]")

	assetManifest := app.Opts.StringConfig("asset-manifest", "assets.json",
		"the path to the JSON asset manifest file [assets.json]")

	app.RegisterTemplatingProvider(templating)
	app.RegisterHook(func() {
		templating.Debug = app.Debug
		templating.Directory = runtime.JoinPath(app.Path, *templatesDirectory)
		templating.Init(*assetManifest)
	})

	return app

}

func (app *Application) Init(env map[string]interface{}) {

	conf := app.Config

	// Create the log directory if it doesn't exist.
	logPath := runtime.JoinPath(app.Path, *conf.logDirectory)
	err := os.MkdirAll(logPath, 0755)
	if err != nil {
		runtime.StandardError(err)
	}

	// Create the run directory if it doesn't exist.
	runPath := runtime.JoinPath(app.Path, *conf.runDirectory)
	err = os.MkdirAll(runPath, 0755)
	if err != nil {
		runtime.StandardError(err)
	}

	// Initialise the process-related resources.
	runtime.Init()
	runtime.InitProcess(app.Name, runPath)

	fmt.Printf("Running %s on %s:%d\n", app.Name, *conf.host, *conf.port)
	app.HandleRequests()

}

func (app *Application) HandleRequests() {
	for {

	}
}

func (app *Application) HandleRequest(path string) {
	ctx := &Context{}
	matches := app.Services.MatchPrefix(path)
	if len(matches) == 0 {
		panic("no match")
	}
	var service *Service
	for i := len(matches); i >= 0; i-- {
		match := matches[0]
		service = match.Value.(*Service)
		if match.Suffix != "" && match.Suffix[0] != '/' && !service.wildcard {
			continue
		}
		break
	}
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

// -----------------------------------------------------------------------------
// Constructor
// -----------------------------------------------------------------------------

func App(name, version string) *Application {

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

	app.Hooks = make([]func(), 0)
	app.Services = structure.NewPrefixTree()

	// Setup default command line options.
	conf.debug = opts.Bool([]string{"-d", "--debug"}, false,
		"enable debug mode")

	conf.genConfig = opts.Bool([]string{"-g", "--gen-config"}, false,
		"show the default yaml config")

	return app

}
