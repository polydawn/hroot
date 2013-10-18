package main

import (
	. "fmt"
	"os"
	"strings"
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

//Given a URI, return the scheme name separate from everything else
//	See: https://en.wikipedia.org/wiki/URI_scheme#Generic_syntax
func ParseURI(input string) (string, string) {
	//Parse input
	components := strings.SplitN(input, ":", 2)
	scheme := components[0]

	//Check if a path was provided
	path := ""
	if len(components) > 1 {
		path = components[1]
	}

	//Check that the scheme name is one we support
	switch scheme {
		case "graph", "file", "docker", "index": //pass
		case "":
			Println("Command source/destination is empty; must be one of (graph, file, docker, index)")
			os.Exit(1)
		default:
			Println("Unrecognized scheme '" + scheme + "': must be one of (graph, file, docker, index)")
			os.Exit(1)
	}

	return scheme, path
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
