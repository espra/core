// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package nodule

import (
	"amp/logging"
	"amp/yaml"
	"exec"
	"fmt"
	"os"
	"path/filepath"
)

type Nodule struct {
	Name    string
	Path    string
	Conf    *Config
	version []byte
}

func (nodule *Nodule) Error(process string, error os.Error) os.Error {
	logging.ErrorData(
		"node",
		fmt.Sprintf("Error %s the %s nodule: %s", process, nodule.Name, error))
	return error
}

func (nodule *Nodule) ConfigError(msg string, v ...interface{}) os.Error {
	return nodule.Error("loading config for", fmt.Errorf(msg, v...))
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
			build = goBuild
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
			test = goTest
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
			length := len(goDepends) + 1
			depends = make([]string, length)
			copy(depends, goDepends)
			depends[length-1] = nodule.Name
		default:
			depends = defaultDepends
		}
	}

	ignore, ok := conf.GetStringList("ignore")
	if !ok {
		switch typ {
		case "go":
			ignore = goIgnore
		default:
			ignore = defaultIgnore
		}
	}

	nodule.Conf = &Config{
		Type:    typ,
		Build:   build,
		Run:     run,
		Test:    test,
		Depends: depends,
		Ignore:  ignore,
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

func (nodule *Nodule) Test() (err os.Error) {
	err = nodule.Build()
	if err != nil {
		return
	}
	Log("Testing the %s nodule: %s", nodule.Name, nodule.Path)
	return nodule.handleCommon("testing", nodule.Conf.Test)
}

func (nodule *Nodule) Version() (version []byte, err os.Error) {
	if nodule.version != nil {
		return nodule.version, nil
	}
	if nodule.Conf == nil {
		err = nodule.LoadConf()
		if err != nil {
			return
		}
	}
	return
}

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
