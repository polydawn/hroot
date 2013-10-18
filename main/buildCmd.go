package main

import (
	. "fmt"
	"os"
	"polydawn.net/docket/confl"
)

type buildCmdOpts struct {
	Source      string `short:"s" long:"source"      default:"graph" description:"Container source."`
	Destination string `short:"d" long:"destination" default:"graph" description:"Container destination."`
}

const DefaultBuildTarget = "build"

//Transforms a container
func (opts *buildCmdOpts) Execute(args []string) error {
	//Get configuration
	target   := GetTarget(args, DefaultBuildTarget)
	settings := confl.NewConfigLoad(".")
	config := settings.GetConfig(target)
	saveAs := settings.GetDefaultImage()
	_ = saveAs

	//Right now, go-flags' default announation does not appear to work when in a sub-command.
	//	Will investigate and hopefully remove this later.
	if opts.Source == "" {
		opts.Source = "graph"
	}
	if opts.Destination == "" {
		opts.Destination = "graph"
	}

	//Parse input/output URIs
	sourceScheme, sourcePath           := ParseURI(opts.Source)
	destinationScheme, destinationPath := ParseURI(opts.Destination)

	_, _ = sourcePath, destinationPath //remove later

	//Prepare input
	switch sourceScheme {
		case "docker":
			//TODO: check that docker has the image loaded
		case "graph", "file", "index":
			Println("Input scheme", sourceScheme, "is not supported yet.")
			os.Exit(1)
	}

	//Prepare output
	switch destinationScheme {
		case "docker":
			//TODO: tag image when done
		case "graph", "file", "index":
			Println("Input scheme", sourceScheme, "is not supported yet.")
			os.Exit(1)
	}

	//Start or connect to a docker daemon
	dock := StartDocker(settings)

	// Launch the container and wait for it to finish
	container := Launch(dock, config)
	container.Wait()

	//Remove if desired
	if config.Purge {
		container.Purge()
	}

	//Stop the docker daemon if it's a child process
	dock.Slay()

	return nil
}
