package main

import (
	. "fmt"
	"polydawn.net/docket/confl"
	"polydawn.net/docket/crocker"
)

//If the user specified a target, use that, else use the command's default target
func GetTarget(args []string, defaultTarget string) string {
	if len(args) >= 1 {
		return args[0]
	} else {
		return defaultTarget
	}
}

//Find or start a docker
func StartDocker(settings *confl.ConfigLoad) *crocker.Dock {
	dock := crocker.NewDock(settings.Dock)

	//Announce the docker
	if dock.IsChildProcess() {
		Println("Started a docker in", dock.Dir())
	} else {
		Println("Connecting to docker", dock.Dir())
	}

	return dock
}

//Helper function: maps a TrionConfig struct to crocker function.
//Kinda ugly; this situation may improve once our config shenanigans solidifies a bit.
func Launch(dock *crocker.Dock, config crocker.ContainerConfig) *crocker.Container {
	return crocker.Launch(dock, config.Image, config.Command, config.Attach, config.Privileged, config.Folder, config.DNS, config.Mounts, config.Ports, config.Environment)
}
