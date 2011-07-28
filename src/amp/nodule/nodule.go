// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package nodule

import (
	"amp/logging"
	"amp/master"
	"amp/runtime"
	"amp/yaml"
	"bytes"
	"exec"
	"exp/template"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// Find all nodules at or within the immediate subdirectories of the given
// ``directory``.
func Find(directory string) (nodules []*Nodule, err os.Error) {

	directory, err = filepath.Abs(filepath.Clean(directory))
	if err != nil {
		return
	}

	root, err := os.Open(directory)
	if err != nil {
		return
	}

	defer root.Close()

	stat, err := root.Stat()
	if err != nil {
		return
	}

	if !stat.IsDirectory() {
		return nil, fmt.Errorf("No directory found at %q.", directory)
	}

	file := filepath.Join(directory, "nodule.yaml")

	fileobj, err := os.Open(file)
	if err == nil {
		fileobj.Close()
		return []*Nodule{&Nodule{Name: stat.Name, Path: directory}}, nil
	}

	nodules = make([]*Nodule, 0)

	for {
		items, err := root.Readdirnames(100)
		if len(items) == 0 {
			if err != nil && err != os.EOF {
				return nil, err
			}
			break
		}
		for _, name := range items {
			path := filepath.Join(directory, name)
			stat, err := os.Stat(path)
			if err != nil {
				return nil, err
			}
			if stat.IsDirectory() {
				file := filepath.Join(path, "nodule.yaml")
				fileobj, err = os.Open(file)
				if err == nil {
					fileobj.Close()
					nodules = append(nodules, &Nodule{Name: name, Path: path})
				}
			}
		}
	}

	return nodules, nil

}

type ConfigEnv struct {
	Profile  string
	Platform string
	Darwin   bool
	Linux    bool
	FreeBSD  bool
}

var defaultEnv *ConfigEnv

func GetConfigEnv() *ConfigEnv {
	if defaultEnv != nil {
		return defaultEnv
	}
	env := &ConfigEnv{
		Profile:  runtime.Profile,
		Platform: runtime.Platform,
	}
	switch runtime.Platform {
	case "linux":
		env.Linux = true
	case "freebsd":
		env.FreeBSD = true
	case "darwin":
		env.Darwin = true
	}
	defaultEnv = env
	return env
}

func EvalStrings(name string, list []string, data interface{}) ([]string, os.Error) {
	result := make([]string, len(list))
	for idx, value := range list {
		if strings.IndexRune(value, '{') == -1 {
			result[idx] = value
		} else {
			tpl := template.New(name)
			buf := &bytes.Buffer{}
			err := tpl.Parse(value)
			if err != nil {
				return nil, err
			}
			err = tpl.Execute(buf, data)
			if err != nil {
				return nil, err
			}
			result[idx] = buf.String()
		}
	}
	return result, nil
}

type Config struct {
	Type    string
	Build   []string
	Run     []string
	Test    []string
	Depends []string
}

type Nodule struct {
	Name string
	Path string
	Conf *Config
}

func (nodule *Nodule) Error(process string, error os.Error) os.Error {
	logging.ErrorData(
		"node",
		fmt.Sprintf("Error %s the %s nodule: %s", process, nodule.Name, error))
	return error
}

func (nodule *Nodule) ConfigError(msg string, v ...interface{}) os.Error {
	err := fmt.Errorf(msg, v...)
	return nodule.Error("loading config for", err)
}

func (nodule *Nodule) LoadConf() (err os.Error) {

	filename := filepath.Join(nodule.Path, "nodule.yaml")
	conf, err := yaml.ParseFile(filename)
	if err != nil {
		return nodule.ConfigError(err.String())
	}

	typ, ok := conf.GetString("type")
	if !ok {
		return nodule.ConfigError("Couldn't get a config value for 'type' in %s", filename)
	}

	run, ok := conf.GetStringList("run")
	if !ok {
		switch typ {
		case "go":
			run = []string{"./" + nodule.Name, "{{$.Profile}}.yaml"}
		default:
			return nodule.ConfigError("Couldn't get a config value for 'run' in %s", filename)
		}
	}

	if len(run) == 0 {
		return nodule.ConfigError("Couldn't get a config value for 'run' in %s", filename)
	}

	env := GetConfigEnv()
	run, err = EvalStrings("run", run, env)
	if err != nil {
		return nodule.ConfigError(err.String())
	}

	build, ok := conf.GetStringList("build")
	if !ok {
		switch typ {
		case "go":
			build = []string{"{{if $.FreeBSD}}gmake{{else}}make{{end}}"}
		default:
			return nodule.ConfigError("Couldn't get a config value for 'build' in %s", filename)
		}
	}

	if len(build) == 0 {
		return nodule.ConfigError("Couldn't get a config value for 'build' in %s", filename)
	}

	build, err = EvalStrings("build", build, env)
	if err != nil {
		return nodule.ConfigError(err.String())
	}

	test, ok := conf.GetStringList("test")
	if !ok {
		switch typ {
		case "go":
			test = []string{"gotest", "-v"}
		default:
			return nodule.ConfigError("Couldn't get a config value for 'test' in %s", filename)
		}
	}

	if len(test) == 0 {
		return nodule.ConfigError("Couldn't get a config value for 'test' in %s", filename)
	}

	test, err = EvalStrings("test", test, env)
	if err != nil {
		return nodule.ConfigError(err.String())
	}

	depends, ok := conf.GetStringList("depends")
	if !ok {
		switch typ {
		case "go":
			depends = []string{"*.go"}
		default:
			depends = make([]string, 0)
		}
	}

	nodule.Conf = &Config{
		Type:    typ,
		Build:   build,
		Run:     run,
		Test:    test,
		Depends: depends,
	}

	return

}

func (nodule *Nodule) handleCommon(process string, args []string) (err os.Error) {
	exe, err := exec.LookPath(args[0])
	if err != nil {
		return nodule.Error(process, err)
	}
	cmd := &exec.Cmd{
		Path: exe,
		Args: args,
		Dir:  nodule.Path,
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		logging.ErrorData("node", fmt.Errorf("Error running %v: %s", args, err))
		logging.ErrorData("node", "\n\n"+string(output))
		return
	}
	return
}

func (nodule *Nodule) Build() (err os.Error) {
	Log("Building the %s nodule: %s", nodule.Name, nodule.Path)
	if nodule.Conf == nil {
		err = nodule.LoadConf()
		if err != nil {
			return
		}
	}
	return nodule.handleCommon("building", nodule.Conf.Build)
}

func (nodule *Nodule) Test() (err os.Error) {
	err = nodule.Build()
	if err != nil {
		return
	}
	Log("Testing the %s nodule: %s", nodule.Name, nodule.Path)
	return nodule.handleCommon("testing", nodule.Conf.Test)
}

func (nodule *Nodule) Review() (err os.Error) {
	err = nodule.Test()
	if err != nil {
		return
	}
	Log("Reviewing the %s nodule: %s", nodule.Name, nodule.Path)
	return
}

func (nodule *Nodule) Run() (err os.Error) {
	Log("Running the %s nodule: %s", nodule.Name, nodule.Path)
	return nodule.handleCommon("running", nodule.Conf.Run)
}

type Host struct {
	CachePath    string
	ControlTCP   net.Listener
	ControlUnix  net.Listener
	LastRef      uint64
	Listener     net.Listener
	Nodules      map[uint64]*Nodule
	ReadTimeout  int64
	WriteTimeout int64
}

func (host *Host) Run(debug bool) (err os.Error) {
	for {

	}
	return
}

func NewHost(runPath, hostAddress string, hostPort int, ctrlAddress string, ctrlPort int, initNodules string, nodulePaths []string, master *master.Client) (host *Host, err os.Error) {

	// Create the cache directory if it doesn't exist.
	cachePath := filepath.Join(runPath, "cache")
	err = os.MkdirAll(cachePath, 0755)
	if err != nil {
		return
	}

	ctrlTCP, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ctrlAddress, ctrlPort))
	if err != nil {
		return
	}

	socket := filepath.Join(runPath, "node.sock")
	_, err = os.Stat(socket)
	if err == nil {
		err = os.Remove(socket)
		if err != nil {
			return
		}
	}

	ctrlUnix, err := net.Listen("unix", socket)
	if err != nil {
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostAddress, hostPort))
	if err != nil {
		return
	}

	host = &Host{
		CachePath:    cachePath,
		ControlTCP:   ctrlTCP,
		ControlUnix:  ctrlUnix,
		Listener:     listener,
		ReadTimeout:  60 * 1e9,
		WriteTimeout: 60 * 1e9,
	}

	return

}

func Log(msg string, args ...interface{}) {
	if len(args) > 0 {
		logging.InfoData("node", fmt.Sprintf(msg, args...))
	} else {
		logging.InfoData("node", msg)
	}
}

func FilterConsoleLog(record *logging.Record) (write bool, data []interface{}) {
	if len(record.Items) > 0 {
		meta := record.Items[0]
		switch meta.(type) {
		case string:
			if meta.(string) == "node" {
				return true, record.Items[1:]
			}
		}
	}
	return true, nil
}
