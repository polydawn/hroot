package trion

import (
	. "fmt"
	"io/ioutil"
	"encoding/json"
	"os"
)

const FileName = "docker.json"

type TrionConfig struct {
	Image          string   //What docker image to use
	StartIn        string   //Which folder to start
	Privileged     bool     //Run in privileged mode?
	Mount      [][]string   //Array of mounts (each an array of strings: host-folder, dock-folder, "ro"/"rw" permission setting)
	Command      []string   //What command to run
	Attach         bool     //Attach interactive terminal?
	Quiet          bool     //Suppress docker output entirely?
	DNS          []string   //Do you want to use custom DNS servers?
	Build        []string   //What command to run when building
	Upstream       string   //What image to use when building
	Purge          bool     //Delete when done?
	Environment [][]string  //Env variables (each an array of strings: variable, value)
	             //DAT ALIGNNMENT. SO GOOD.
}

var DefaultTrionConfig = TrionConfig {
	"ubuntu",               //Image
	"/",                    //StartIn
	false,                  //Privileged
	[][]string{},           //Mount
	[]string{"launch.sh"},  //Command
	false,                  //Attach
	false,                  //Quiet
	[]string{},             //DNS
	[]string{"build.sh"},   //Build
	"ubuntu",               //Upstream
	false,                  //Purge
	[][]string{},           //Environment
}

//Recursively finds configuration files and loads them top-down.
//This lets you have a base configuration in the parent directory and override it for specific containers.
func FindConfig(dir string) TrionConfig {
	file, stack, config, loaded := FileName, new(Stack), DefaultTrionConfig, 0

	//recurse up the file tree looking for configuration
	for {
		data, err := ioutil.ReadFile(dir+"/"+file)
		if err != nil {
			break
		} else {
			stack.Push(data)
			file = "../" + file
		}
	}


	//Apply the configuration file(s)
	for stack.Len() > 0 {
		content := LoadConfig(stack.Pop().([]byte))
		AddConfig(&content, &config)
		loaded++
	}

	if loaded > 0 && !config.Quiet {
		Println("Loaded", loaded, "config files.")
	}

	return config
}

//Load data into struct
func LoadConfig(data []byte) TrionConfig {
	var content TrionConfig
	err := json.Unmarshal(data, &content)

	if (err != nil) {
		Println("Cannot decode JSON:", err.Error())
		os.Exit(1)
	}

	return content
}

//Loads a configuration object, overriding the base
func AddConfig(inc, base *TrionConfig) {
	if inc.Image != "" {
		base.Image = inc.Image
	}
	if inc.Upstream != "" {
		base.Upstream = inc.Upstream
	}
	if inc.StartIn != "" {
		base.StartIn = inc.StartIn
	}
	base.Privileged = inc.Privileged
	base.Mount = append(base.Mount, inc.Mount ...)
	base.Environment = append(base.Environment, inc.Environment ...)
	if len(inc.Command) != 0 {
		base.Command = inc.Command
	}
	if len(inc.Build) != 0 {
		base.Build = inc.Build
	}
	base.Purge = inc.Purge
	base.Attach = inc.Attach
	base.Quiet = inc.Quiet
	base.DNS = append(base.DNS, inc.DNS ...)
}
