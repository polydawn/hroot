// Configuration struct and methods

package trion

import (
	. "fmt"
	"os"
	"path/filepath"
	"strings"
)

//Configuration data
type TrionConfig struct {
	//What docker image to use
	Image       string     `toml:image`

	//What command to run
	Command     []string   `toml:command`

	//Which folder to start
	Folder     string     `toml:folder`

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

//Default configuration
var DefaultTrionConfig = TrionConfig {
	Image:       "ubuntu",
	Command:     []string{"/bin/echo", "Hello from docker!"},
	Folder:      "/",
	Privileged:  false,
	Mounts:      [][]string{},
	Ports:       [][]string{},
	DNS:         []string{},
	Attach:      false,
	Purge:       false,
	Environment: [][]string{},
}

//Preprocess a configuration object
func (c *TrionConfig) Prepare(dir string) {
	//Get the absolute directory this config is relative to
	cwd, err := filepath.Abs(dir)
	if err != nil {
		Println("Fatal: Cannot determine absolute path:", dir)
		os.Exit(1)
	}

	//Handle mounts
	for i := range c.Mounts {

		//Check for triple-dot ... notation, which is relative to that config's directory, not the CWD
		if strings.Index(c.Mounts[i][0], "...") != 1 {
			c.Mounts[i][0] = strings.Replace(c.Mounts[i][0], "...", cwd, 1)
		}

		//Find the absolute path for each host mount
		abs, err := filepath.Abs(c.Mounts[i][0])
		if err != nil {
			Println("Fatal: Cannot determine absolute path:", c.Mounts[i][0])
			os.Exit(1)
		}
		c.Mounts[i][0] = abs
	}
}

//Target name > configration mapping (used in the 'docker.toml' config files)
type TrionConfigs map[string]TrionConfig
