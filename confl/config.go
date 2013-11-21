package confl

import (
	// "path/filepath"
	// "strings"
	// . "polydawn.net/docket/util"
)
/*
//Extract a configuration target from loaded settings.
//Each config will override the one after it.
func (cs *ConfigFile) GetConfig(target string) *Settings {
	config := DefaultSettings

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

	return &config
}


//Preprocess a configuration object
func preprocess(c *Settings, dir string) {
	//Get the absolute directory this config is relative to
	cwd, err := filepath.Abs(dir)
	if err != nil { ExitGently("Cannot determine absolute path: ", dir) }

	//Handle mounts
	for i := range c.Mounts {

		//Check for triple-dot ... notation, which is relative to that config's directory, not the CWD
		if strings.Index(c.Mounts[i][0], "...") != 1 {
			c.Mounts[i][0] = strings.Replace(c.Mounts[i][0], "...", cwd, 1)
		}

		//Find the absolute path for each host mount
		abs, err := filepath.Abs(c.Mounts[i][0])
		if err != nil { ExitGently("Cannot determine absolute path:", c.Mounts[i][0]) }
		c.Mounts[i][0] = abs
	}
}

//Loads a configuration object, overriding the base
func addConfig(base *Settings, inc Settings, meta toml.MetaData, target string) {

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
*/
