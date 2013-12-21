package conf

import (
	"io/ioutil"
	"path/filepath"
)

const ConfigFileName = "docker.toml"

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
			dockDir  := filepath.Join(dir, DockFolder)
			graphDir := filepath.Join(dir, GraphFolder)

			if _, err := ioutil.ReadDir(dockDir) ; err == nil {
				folders.Dock = dockDir
			}
			if _, err := ioutil.ReadDir(graphDir) ; err == nil {
				folders.Graph = graphDir
			}

			//Parse file, increment folder
			data := string(buf)
			dir = filepath.Join("../", dir)

			//Hold data
			files = append(files, data)
			dirs  = append(dirs, dir)
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
