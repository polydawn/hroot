package main

import (
	. "fmt"
	"polydawn.net/dockctrl/confl"
	"polydawn.net/dockctrl/crocker"
)

/*
	Helps run anything that requires a docker connection.
	Handles creation & cleanup in one place.
	Docker daemon config is determined by looking around the cwd.
*/
func WithDocker(fn func(*crocker.Dock, *confl.ConfigLoad) error) error {
	//Load configuration, then find or start a docker
	settings := confl.NewConfigLoad(".")
	dock := crocker.NewDock("./dock")

	//Announce the docker
	if dock.IsChildProcess() {
		Println("Started a docker in", dock.Dir())
	} else {
		Println("Connecting to docker", dock.Dir())
	}

	//Run the closure, kill the docker if needed, and return any errors.
	err := fn(dock, settings)
	dock.Slay()
	return err
}

//Helper function: maps a TrionConfig struct to crocker function.
//Kinda ugly; this situation may improve once our config shenanigans solidifies a bit.
func Launch(dock *crocker.Dock, config crocker.ContainerConfig) *crocker.Container {
	return crocker.Launch(dock, config.Image, config.Command, config.Attach, config.Privileged, config.Folder, config.DNS, config.Mounts, config.Ports, config.Environment)
}
