package confl

import (
	. "fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"path/filepath"
	"polydawn.net/docket/crocker"
	"strings"
)

//Hold the metadata
type ConfigLoad struct {
	configs []ContainerConfigs
	metas   []toml.MetaData
	Dock    string
	Graph   string
}

//Target name > configration mapping (used in the 'docker.toml' config files)
type ContainerConfigs map[string]crocker.ContainerConfig

const ConfigFileName = "docker.toml"
const DefaultTarget  = "default"
const DockFolder     = "dock"
const GraphFolder    = "graph"

//Recursively finds configuration files & folders
func NewConfigLoad(dir string) *ConfigLoad {
	load := &ConfigLoad{
		Dock:  dir + "/" + DockFolder,
		Graph: dir + "/" + GraphFolder,
	}

	//Recurse up the file tree looking for configuration
	for {
		buf, err := ioutil.ReadFile(dir + "/" + ConfigFileName)

		if err != nil {
			break
		} else {
			//Parse file, store in slices
			config, metadata := loadToml(string(buf), dir)
			load.configs = append(load.configs, config)
			load.metas   = append(load.metas,   metadata)

			//Check for folders
			if _, err = ioutil.ReadDir(dir + "/" + DockFolder) ; err == nil {
				load.Dock = dir + "/" + DockFolder
			}
			if _, err = ioutil.ReadDir(dir + "/" + GraphFolder) ; err == nil {
				load.Graph = dir + "/" + GraphFolder
			}

			dir = "../" + dir
		}
	}

	if n := len(load.configs); n > 0 {
		Println("Loaded", n, "config files.")
	}

	return load
}

//Extract a configuration target from loaded settings.
//Each config will override the one after it.
func (cs *ConfigLoad) GetConfig(target string) crocker.ContainerConfig {
	config := crocker.DefaultContainerConfig

	//For each config
	for i := len(cs.configs) - 1; i >= 0; i-- {
		newConfig := cs.configs[i]
		meta   := cs.metas[i]

		//If default target is provided, load it unconditionally
		if meta.IsDefined(DefaultTarget) {
			addConfig(&config, newConfig[DefaultTarget], meta, DefaultTarget)
		}

		//If default target is provided, load it unconditionally
		if target != DefaultTarget && meta.IsDefined(target) {
			addConfig(&config, newConfig[target], meta, target)
		}
	}

	return config
}

//Return the default image name (convenience function)
func (cs *ConfigLoad) GetDefaultImage() string {
	return cs.GetConfig(DefaultTarget).Image
}

func loadToml(data, dir string) (ContainerConfigs, toml.MetaData) {
	//Decode the file
	var set ContainerConfigs
	md, err := toml.Decode(data, &set)

	//Check for errors
	if err != nil {
		Println("Fatal: could not decode file in ", dir, err)
		os.Exit(1)
	}

	//Prepare all targets
	for _, conf := range set {
		preprocess(&conf, dir)
	}

	return set, md
}

//Preprocess a configuration object
func preprocess(c *crocker.ContainerConfig, dir string) {
	//Get the absolute directory this config is relative to
	cwd, err := filepath.Abs(dir)
	if err != nil {
		Println("Fatal: Cannot determine absolute path:", dir)
		os.Exit(1)
	}

	//Handle mounts
	for i := range c.Mounts {

		//Check for triple-dot ... notation, which is relative to that config's directory, not the CWD
		if strings.Index(c.Mounts[i][0], "...") != 1 {
			c.Mounts[i][0] = strings.Replace(c.Mounts[i][0], "...", cwd, 1)
		}

		//Find the absolute path for each host mount
		abs, err := filepath.Abs(c.Mounts[i][0])
		if err != nil {
			Println("Fatal: Cannot determine absolute path:", c.Mounts[i][0])
			os.Exit(1)
		}
		c.Mounts[i][0] = abs
	}
}

//Loads a configuration object, overriding the base
func addConfig(base *crocker.ContainerConfig, inc crocker.ContainerConfig, meta toml.MetaData, target string) {

	if meta.IsDefined(target, "image") {
		base.Image = inc.Image
	}

	if meta.IsDefined(target, "upstream") {
		base.Upstream = inc.Upstream
	}

	if meta.IsDefined(target, "index") {
		base.Index = inc.Index
	}

	if meta.IsDefined(target, "command") {
		base.Command = inc.Command
	}

	if meta.IsDefined(target, "folder") {
		base.Folder = inc.Folder
	}

	if meta.IsDefined(target, "privileged") {
		base.Privileged = inc.Privileged
	}

	if meta.IsDefined(target, "mounts") {
		base.Mounts = append(base.Mounts, inc.Mounts...)
	}

	if meta.IsDefined(target, "ports") {
		base.Ports = append(base.Ports, inc.Ports...)
	}

	if meta.IsDefined(target, "dns") {
		base.DNS = append(base.DNS, inc.DNS...)
	}

	if meta.IsDefined(target, "attach") {
		base.Attach = inc.Attach
	}

	if meta.IsDefined(target, "purge") {
		base.Purge = inc.Purge
	}

	if meta.IsDefined(target, "environment") {
		base.Environment = append(base.Environment, inc.Environment...)
	}
}
