package trion

import (
	"github.com/BurntSushi/toml"
)

//Hold the metadata
type TrionSettings struct {
	configs []TrionConfigs
	metas   []toml.MetaData
}

const DefaultTarget = "default"

//Extract a configuration target from loaded settings.
//	Each config will override the one after it.
func (cs *TrionSettings) GetConfig(target string) TrionConfig {
	config := DefaultTrionConfig

	//For each config
	for i := len(cs.configs) -1 ; i >= 0; i-- {
		newConfig := cs.configs[i]
		meta   := cs.metas[i]

		//If default target is provided, load it unconditionally
		if meta.IsDefined(DefaultTarget) {
			AddConfig(&config, newConfig[DefaultTarget], meta, DefaultTarget)
		}

		//If default target is provided, load it unconditionally
		if target != DefaultTarget && meta.IsDefined(target) {
			AddConfig(&config, newConfig[target], meta, target)
		}
	}

	return config
}

//Loads a configuration object, overriding the base
func AddConfig(base *TrionConfig, inc TrionConfig, meta toml.MetaData, target string) {

	if meta.IsDefined(target, "image") {
		base.Image = inc.Image
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
