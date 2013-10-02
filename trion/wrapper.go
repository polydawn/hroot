//High-level functionality

package trion

import (
	"polydawn.net/dockctrl/crocker"
	. "fmt"
)

const ExportPath = "./" //Where to export docker images

//Helps run anyything that requires a docker connection.
//Handles creation & cleanup in one place.
func WithDocker(fn func(TrionConfig, *Command) error ) error {
	//Load configuration, then find or start a docker
	config := FindConfig(".")
	dock, dir, ours := crocker.FindDock()
	cmd := &Command{dock.Client()}

	//Announce the docker
	if ours {
		Println("Started a docker in", dir)
	} else {
		Println("Connecting to docker", dir)
	}

	//Run the closure, kill the docker if needed, and return any errors.
	err := fn(config, cmd)
	dock.Slay()
	return err
}

//Launches a docker
func Launch(config TrionConfig, cmd *Command) error {
	//Start the docker and wait for it to finish
	CID := cmd.Run(config)
	cmd.Wait(CID)

	//Remove if desired
	if config.Purge {
		cmd.Purge(CID)
	}

	return nil
}

//Builds a docker
func Build(config TrionConfig, cmd *Command) error {
	//Use the build command and upstream image
	buildConfig        := config
	buildConfig.Command = config.Build
	buildConfig.Image   = config.Upstream

	//Run the build
	CID := cmd.Run(buildConfig)
	cmd.Wait(CID)

	//Create a tar
	cmd.Export(CID, ExportPath)

	//Import the built docker
	// Todo: add --noImport option to goflags
	cmd.Import(config, ExportPath)

	//Remove if desired
	if config.Purge {
		cmd.Purge(CID)
	}

	return nil
}
