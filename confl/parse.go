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
	configs []*Configuration
	metas []*toml.MetaData
}

func (p *TomlConfigParser) AddConfig(data, dir string) {

}

func (p *TomlConfigParser) GetConfig() *Configuration {
	config := DefaultConfiguration


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
