package crocker

//Configuration data
type ContainerConfig struct {
	//What image to use
	Image       string     `toml:image`

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

//Default configuration
var DefaultContainerConfig = ContainerConfig{
	Image:       "ubuntu",
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
