package confl

import (
	"path/filepath"
	// "github.com/BurntSushi/toml"
)

const DockFolder     = "dock"
const GraphFolder    = "graph"

//Image and parent image
type Image struct {
	//What image to use
	Name        string     `toml:name`

	//What image to build from
	Upstream    string     `toml:upstream`

	//What the upstream image is called in the docker index
	Index       string     `toml:index`
}

//A container's settings
type Container struct {
	//What command to run
	Command     []string   `toml:command`

	//Which folder to start
	Folder      string     `toml:folder`

	//Run in privileged mode?
	Privileged  bool       `toml:privileged`

	//Array of mounts (each an array of strings: hostfolder, guestfolder, "ro"/"rw" permission)
	Mounts      [][]string `toml:mounts`

	//What ports do you want to forward? (each an array of ints: hostport, guestport)
	Ports       [][]string `toml:ports`

	//Do you want to use custom DNS servers?
	DNS         []string   `toml:dns`

	//Attach interactive terminal?
	Attach      bool       `toml:attach`

	//Delete when done?
	Purge       bool       `toml:purge`

	//Env variables (each an array of strings: variable, value)
	Environment [][]string `toml:environment`
}

//Default container
var DefaultContainer = Container {
	Command:     []string{"/bin/echo", "Hello from docket!"},
	Folder:      "/",
	Privileged:  false,
	Mounts:      [][]string{},
	Ports:       [][]string{},
	DNS:         []string{},
	Attach:      false,
	Purge:       false,
	Environment: [][]string{},
}

//Docket configuration
type Configuration struct {
	//The image struct
	Image    Image                `toml:image`

	//The settings struct
	Settings Container            `toml:settings`

	//A map of named targets, each representing another set of container settings
	Targets  map[string]Container `toml:targets`
}

func (c *Configuration) GetTargetContainer(target string) Container {
	return c.Targets[target]
}

//Default configuration
var DefaultConfiguration = Configuration {
	Settings: DefaultContainer,
}

//Folder location
type Folders struct {
	//Where we've decided the dock  folder is or should be
	Dock string

	//Where we've decided the graph folder is or should be
	Graph string
}

//Default folders
func DefaultFolders(dir string) *Folders {
	return &Folders {
		Dock:  filepath.Join(dir, DockFolder),
		Graph: filepath.Join(dir, GraphFolder),
	}
}
