package main

import (
	. "fmt"
	"polydawn.net/dockctrl/crocker"
	"polydawn.net/dockctrl/trion"
)

/*
	Helps run anything that requires a docker connection.
	Handles creation & cleanup in one place.
	Docker daemon config is determined by looking around the cwd.
*/
func WithDocker(fn func(*crocker.Dock, *trion.TrionSettings, []string) error, args []string) error {
	//Load configuration, then find or start a docker
	settings := trion.FindConfig(".")
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
func Launch(dock *crocker.Dock, config trion.TrionConfig) *crocker.Container {
	return crocker.Launch(dock, config.Image, config.Command, config.Attach, config.Privileged, config.Folder, config.DNS, config.Mounts, config.Ports, config.Environment)
}
