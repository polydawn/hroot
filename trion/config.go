package trion

import (
	. "fmt"
	. "polydawn.net/dockctrl/util"
	"io/ioutil"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const FileName = "docker.json"

type TrionConfig struct {
	Image          string   //What docker image to use
	StartIn        string   //Which folder to start
	Privileged     bool     //Run in privileged mode?
	Mounts     [][]string   //Array of mounts (each an array of strings: host-folder, dock-folder, "ro"/"rw" permission setting)
	Command      []string   //What command to run
	Attach         bool     //Attach interactive terminal?
	Quiet          bool     //Suppress docker output entirely?
	DNS          []string   //Do you want to use custom DNS servers?
	Ports      [][]string   //What ports do you want to forward? (each an array of ints: host-port, guest-port)
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
	[][]string{},           //Mounts
	[]string{"launch.sh"},  //Command
	false,                  //Attach
	false,                  //Quiet
	[]string{},             //DNS
	[][]string{},           //Ports
	[]string{"build.sh"},   //Build
	"ubuntu",               //Upstream
	false,                  //Purge
	[][]string{},           //Environment
}

//Recursively finds configuration files and loads them top-down.
//This lets you have a base configuration in the parent directory and override it for specific containers.
func FindConfig(dir string) TrionConfig {
	file, stack, stackDir, config, loaded := FileName, new(Stack), new(Stack), DefaultTrionConfig, 0

	//recurse up the file tree looking for configuration
	for {
		data, err := ioutil.ReadFile(dir+"/"+file)
		if err != nil {
			break
		} else {
			stack.Push(data)
			stackDir.Push(dir)
			dir = "../" + dir
		}
	}

	//Apply the configuration file(s)
	for stack.Len() > 0 {
		content := LoadConfigFromJSON(stack.Pop().([]byte), stackDir.Pop().(string))
		AddConfig(&content, &config)
		loaded++
	}

	if loaded > 0 && !config.Quiet {
		Println("Loaded", loaded, "config files.")
	}

	return config
}

//Load data into struct from a JSON byte array
func LoadConfigFromJSON(data []byte, dir string) TrionConfig {
	var config TrionConfig
	err := json.Unmarshal(data, &config)

	//Check the unmarshalling was successful and that filepath succeeded
	if err != nil {
		Println("Cannot decode JSON:", err.Error())
		os.Exit(1)
	}

	PrepareConfig(&config, dir)
	return config
}

//Pre-process a configuration object
func PrepareConfig(config *TrionConfig, dir string) {
	//Get the absolute directory this config is relative to
	cwd, err := filepath.Abs(dir)
	if err != nil {
		Println("Fatal: Cannot determine absolute path:", dir)
		os.Exit(1)
	}

	//Handle mounts
	for i := range config.Mounts {

		//Check for triple-dot ... notation, which is relative to that config's directory, not the CWD
		if strings.Index(config.Mounts[i][0], "...") != -1 {
			config.Mounts[i][0] = strings.Replace(config.Mounts[i][0], "...", cwd, -1)
		}

		//Find the absolute path for each host mount
		abs, err := filepath.Abs(config.Mounts[i][0])
		if err != nil {
			Println("Fatal: Cannot determine absolute path:", config.Mounts[i][0])
			os.Exit(1)
		}
		config.Mounts[i][0] = abs
	}
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
	base.Mounts = append(base.Mounts, inc.Mounts ...)
	base.Ports = append(base.Ports, inc.Ports ...)
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
