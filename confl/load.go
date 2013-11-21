package confl
import (
	"io/ioutil"
	"path/filepath"
	// "strings"
	// . "polydawn.net/docket/util"
)

const ConfigFileName = "docker.toml"

//A generic interface for loading configuration.
//Our implementation reads TOML files; roll your own!
type ConfigParser interface {
	//Called in order of file discovery (last entry is the highest parent).
	AddConfig(data, dir string)

	//Called to get the final configuration after loading.
	GetConfig() *Configuration
}

//Recursively finds configuration files & folders.
func LoadConfigurationFromDisk(dir string, parser ConfigParser) (*Configuration, *Folders) {
	//Default settings and folders
	folders := DefaultFolders(dir)

	//Recurse up the file tree looking for configuration
	for {

		//Try to read toml file
		buf, err := ioutil.ReadFile(dir + "/" + ConfigFileName)

		//Did we succeed?
		if err == nil {
			//Check for folders
			dockDir  := filepath.Join(dir, DockFolder)
			graphDir := filepath.Join(dir, DockFolder)

			if _, err := ioutil.ReadDir(dockDir) ; err == nil {
				folders.Dock = dockDir
			}
			if _, err := ioutil.ReadDir(graphDir) ; err == nil {
				folders.Graph = graphDir
			}

			//Parse file
			data := string(buf)
			parser.AddConfig(data, dir)
			dir = filepath.Join("../", dir)
		} else {
			break //If the file was not readable, done loading config
		}
	}

	return parser.GetConfig(), folders
}
