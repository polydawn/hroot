package conf

import (
	"io/ioutil"
	"path/filepath"
)

const ConfigFileName = "hroot.toml"

//A generic interface for loading configuration.
//Our implementation reads TOML files; roll your own!
type ConfigParser interface {

	//Parses a new configuration string.
	//Each call should override configuration added before.
	AddConfig(data, dir string) ConfigParser

	//Called to get the final configuration after loading.
	GetConfig() *Configuration

}

//Recursively finds configuration files & folders.
func LoadConfigurationFromDisk(dir string, parser ConfigParser) (*Configuration, *Folders) {
	//Default settings, folders, and parsed data
	folders := DefaultFolders(dir)
	files := []string{}
	dirs  := []string{}

	//Recurse up the file tree looking for configuration
	for {

		//Try to read toml file
		buf, err := ioutil.ReadFile(dir + "/" + ConfigFileName)

		//Did we succeed?
		if err == nil {
			//Check for folders
			folders.Dock  = filepath.Join(dir, DockFolder)
			folders.Graph = filepath.Join(dir, GraphFolder)

			//Convert data to a string, save for later
			data := string(buf)
			files = append(files, data)
			dirs  = append(dirs, dir)

			//Increment folder for next stage
			dir = filepath.Join("../", dir)
		} else {
			break //If the file was not readable, done loading config
		}
	}

	//Unroll data - we discovered them in reverse order, send each to parser
	for n := len(files) - 1; n >= 0; n-- {
		parser.AddConfig(files[n], dirs[n])
	}

	return parser.GetConfig(), folders
}
