//High-level functionality

package trion

import (
	. "fmt"
	"polydawn.net/dockctrl/crocker"
)

const ExportPath = "./" //Where to export docker images

//Helps run anything that requires a docker connection.
//Handles creation & cleanup in one place.
func WithDocker(fn func(*crocker.Dock, *TrionSettings, []string) error, args []string) error {
	//Load configuration, then find or start a docker
	settings := FindConfig(".")
	dock := crocker.NewDock(".")

	//Announce the docker
	if dock.IsChildProcess() {
		Println("Started a docker in", dock.Dir())
	} else {
		Println("Connecting to docker", dock.Dir())
	}

	//Run the closure, kill the docker if needed, and return any errors.
	err := fn(dock, settings, args)
	dock.Slay()
	return err
}

//Helper function: maps a TrionConfig struct to crocker function.
//Kinda ugly; this situation may improve once our config shenanigans solidifies a bit.
func Launch(dock *crocker.Dock, config TrionConfig) *crocker.Container {
	return crocker.Launch(dock, config.Image, config.Command, config.Attach, config.Privileged, config.Folder, config.DNS, config.Mounts, config.Ports, config.Environment)
}

//Launches a docker
func Run(dock *crocker.Dock, settings *TrionSettings, args []string) error {
	//Get the target
	target := args[0] //TODO: replace the args with golflags!

	//Get configuration
	config := settings.GetConfig(target)

	//Start the docker and wait for it to finish
	container := Launch(dock, config)
	container.Wait()

	//Remove if desired
	if config.Purge {
		container.Purge()
	}

	return nil
}

//Exports the result of a target into docker.
func Export(dock *crocker.Dock, settings *TrionSettings, args []string) error {
	//Get the target
	target := args[0] //TODO: replace the args with golflags!

	//Get configuration
	config := settings.GetConfig(target)
	saveAs := settings.GetConfig(DefaultTarget).Image

	//Run the build
	container := Launch(dock, config)
	container.Wait()

	//Create a tar
	container.Export(ExportPath)

	//Import the built docker
	// Todo: add --noImport option to goflags
	container.ImportFromString(ExportPath, saveAs)

	//Remove if desired
	if config.Purge {
		container.Purge()
	}

	return nil
}
