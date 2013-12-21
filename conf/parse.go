package conf

// The only file that imports toml.
// Keeps our chosen file format isolated from the rest of the system.

import (
	"github.com/BurntSushi/toml"
	. "polydawn.net/docket/util"
)

type TomlConfigParser struct {
	config *Configuration
}

func (p *TomlConfigParser) AddConfig(data, dir string) ConfigParser {
	//Load default configuration if no previous data
	if p.config == nil {
		a := DefaultConfiguration
		p.config = &a
	}

	//Parse toml, expand relative paths, and override settings
	conf, meta := ParseString(data)
	conf.Settings.Localize(dir)
	LoadContainerSettings(&p.config.Settings, &conf.Settings, meta)

	//Load image names
	p.config.Image = conf.Image

	//Load any target settings
	p.config.Targets = conf.Targets

	//Chain calls
	return p
}

func (p *TomlConfigParser) GetConfig() *Configuration {
	if p.config == nil {
		return &DefaultConfiguration
	} else {
		return p.config
	}
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
