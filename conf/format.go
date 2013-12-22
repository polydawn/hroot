package conf

import (
	"strings"
	"path/filepath"
	. "polydawn.net/docket/util"
)

const DockFolder     = "dock"
const GraphFolder    = "graph"

//Image and parent image
type Image struct {
	//What image to use
	Name        string     `toml:"name"`

	//What image to build from
	Upstream    string     `toml:"upstream"`

	//What the upstream image is called in the docker index
	Index       string     `toml:"index"`
}

//A container's settings
type Container struct {
	//What command to run
	Command     []string   `toml:"command"`

	//Which folder to start
	Folder      string     `toml:"folder"`

	//Run in privileged mode?
	Privileged  bool       `toml:"privileged"`

	//Array of mounts (each an array of strings: hostfolder, guestfolder, "ro"/"rw" permission)
	Mounts      [][]string `toml:"mounts"`

	//What ports do you want to forward? (each an array of ints: hostport, guestport)
	Ports       [][]string `toml:"ports"`

	//Do you want to use custom DNS servers?
	DNS         []string   `toml:"dns"`

	//Attach interactive terminal?
	Attach      bool       `toml:"attach"`

	//Delete when done?
	Purge       bool       `toml:"purge"`

	//Env variables (each an array of strings: variable, value)
	Environment [][]string `toml:"environment"`
}

//Localize a container object to a given folder
func (c *Container) Localize(dir string) {
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
	Image    Image                `toml:"image"`

	//The settings struct
	Settings Container            `toml:"settings"`

	//A map of named targets, each representing another set of container settings
	Targets  map[string]Container `toml:"target"`
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
