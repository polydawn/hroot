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
	LoadContainerSettings(&p.config.Settings, &conf.Settings, meta, "settings")

	//Load image names
	p.config.Image = conf.Image

	//Load any target settings
	p.config.Targets = conf.Targets

	for x := range p.config.Targets {
		a := p.config.Settings
		b := p.config.Targets[x]
		LoadContainerSettings(&a, &b, meta, "target", x)
		p.config.Targets[x] = a
	}

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

//Loads a container configuration object, overriding a base
//This function prevents empty TOML keys (anything you didn't specify) from overriding a preset value.
func LoadContainerSettings(base *Container, inc *Container, meta *toml.MetaData, key ...string) {

	if meta.IsDefined(append(key, "command")...) {
		base.Command = inc.Command
	}

	if meta.IsDefined(append(key, "folder")...) {
		base.Folder = inc.Folder
	}

	if meta.IsDefined(append(key, "privileged")...) {
		base.Privileged = inc.Privileged
	}

	if meta.IsDefined(append(key, "mounts")...) {
		base.Mounts = append(base.Mounts, inc.Mounts...)
	}

	if meta.IsDefined(append(key, "ports")...) {
		base.Ports = append(base.Ports, inc.Ports...)
	}

	if meta.IsDefined(append(key, "dns")...) {
		base.DNS = append(base.DNS, inc.DNS...)
	}

	if meta.IsDefined(append(key, "attach")...) {
		base.Attach = inc.Attach
	}

	if meta.IsDefined(append(key, "purge")...) {
		base.Purge = inc.Purge
	}

	if meta.IsDefined(append(key, "environment")...) {
		base.Environment = append(base.Environment, inc.Environment...)
	}
}
