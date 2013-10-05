// Handles loading configuration from the filesystem

package trion

import (
	"io/ioutil"
	"os"
	. "fmt"
	"github.com/BurntSushi/toml"
)

const FileName = "docker.toml"

//Recursively finds configuration files.
func FindConfig(dir string) *TrionSettings {
	var configs []TrionConfigs
	var metas   []toml.MetaData

	//Recurse up the file tree looking for configuration
	for {
		buf, err := ioutil.ReadFile(dir + "/" + FileName)

		if err != nil {
			break
		} else {
			//Parse file, store in slices
			config, metadata := LoadToml(string(buf), dir)
			configs = append(configs, config)
			metas   = append(metas,   metadata)

			dir = "../" + dir
		}
	}

	if n := len(configs); n > 0 {
		Println("Loaded", n, "config files.")
	}

	return &TrionSettings{configs, metas}
}

//Load TOML file
func LoadToml(data, dir string) (TrionConfigs, toml.MetaData) {
	//Decode the file
	var set TrionConfigs
	md, err := toml.Decode(data, &set)

	//Check for errors
	if err != nil {
		Println("Fatal: could not decode file in ", dir, err)
		os.Exit(1)
	}

	//Prepare all targets
	for target := range set {
		temp := set[target]
		temp.Prepare(dir)
	}

	return set, md
}
