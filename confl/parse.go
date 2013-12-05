package confl

// The only file that imports toml.
// Keeps our chosen file format isolated from the rest of the system.

import (
	. "fmt"
	"github.com/BurntSushi/toml"
	. "polydawn.net/docket/util"
)

//Holds arrays of ParseData
type TomlConfigParser struct {
	configs  []*Container
	metas    []*toml.MetaData
	image    *Image
	targets   map[string]Container `toml:targets`
	firstFlag bool
}

//Called from lowest folder --> highest, as configuration files are discovered
func (p *TomlConfigParser) AddConfig(data, dir string) {
	//Parse toml and expand relative paths
	conf, meta := ParseString(data)
	conf.Settings.Localize(dir)

	//Add settings to stack
	p.configs = append(p.configs, &conf.Settings)
	p.metas   = append(p.metas,   meta)

	//Save image & targets (from first config file only, image & targets section does not inherit)
	if !p.firstFlag {
		p.image  = &conf.Image
		p.targets = conf.Targets
		p.firstFlag = true
	}
}

func (p *TomlConfigParser) GetConfig() *Configuration {
	//Load default configuration, and return it if no configs were added
	config := DefaultConfiguration
	if !p.firstFlag { return &config }

	//Load image name
	config.Image = *p.image

	//Inheritance starts at the highest file in your folder tree, then works down.
	for n := len(p.configs) - 1; n >= 0; n-- {
		//Load the settings of the new file
		LoadContainerSettings(&config.Settings, p.configs[n], p.metas[n])

		//Last conf file - load target settings
		if n == 0 {

		}
	}

	if n := len(p.configs); n > 0 {
		Println("Loaded", n, "config files.")
	}

	return &config
}

//Parse a TOML-formatted string into a configuration struct.
func ParseString(data string) (*Configuration, *toml.MetaData) {
	var set Configuration

	//Decode the file
	md, err := toml.Decode(data, &set)
	if err != nil { ExitGently("Could not decode file:", err) }

	return &set, &md
}

const prefix = "settings"

//Loads a container configuration object, overriding a base
//This function prevents empty TOML keys (anything you didn't specify) from overriding a preset value.
func LoadContainerSettings(base *Container, inc *Container, meta *toml.MetaData) {
	if meta.IsDefined(prefix, "command") {
		base.Command = inc.Command
	}

	if meta.IsDefined(prefix, "folder") {
		base.Folder = inc.Folder
	}

	if meta.IsDefined(prefix, "privileged") {
		base.Privileged = inc.Privileged
	}

	if meta.IsDefined(prefix, "mounts") {
		base.Mounts = append(base.Mounts, inc.Mounts...)
	}

	if meta.IsDefined(prefix, "ports") {
		base.Ports = append(base.Ports, inc.Ports...)
	}

	if meta.IsDefined(prefix, "dns") {
		base.DNS = append(base.DNS, inc.DNS...)
	}

	if meta.IsDefined(prefix, "attach") {
		base.Attach = inc.Attach
	}

	if meta.IsDefined(prefix, "purge") {
		base.Purge = inc.Purge
	}

	if meta.IsDefined(prefix, "environment") {
		base.Environment = append(base.Environment, inc.Environment...)
	}
}
