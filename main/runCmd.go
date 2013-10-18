package main

import (
	. "fmt"
	"polydawn.net/docket/confl"
	"polydawn.net/docket/dex"
)

type runCmdOpts struct {
	Source      string `short:"s" long:"source" default:"graph" description:"Container source."`
}

const DefaultRunTarget = "default"

//Runs a container
func (opts *runCmdOpts) Execute(args []string) error {
	//Get configuration
	target   := GetTarget(args, DefaultRunTarget)
	settings := confl.NewConfigLoad(".")
	config   := settings.GetConfig(target)
	var sourceGraph *dex.Graph

	//Right now, go-flags' default announation does not appear to work when in a sub-command.
	//	Will investigate and hopefully remove this later.
	if opts.Source == "" {
		opts.Source = "graph"
	}

	//Parse input URI
	sourceScheme, sourcePath := ParseURI(opts.Source)
	_ = sourcePath //remove later

	//Prepare input
	switch sourceScheme {
		case "docker":
			//TODO: check that docker has the image loaded
		case "graph":
			//Look up the graph, and clear any unwanted state
			sourceGraph = dex.NewGraph(settings.Graph)
			sourceGraph.Cleanse()
		case "file":
			//If the user did not specify an image path, set one
			if sourcePath == "" {
				sourcePath = "./image.tar"
			}
		case "index":
			return Errorf("Source " + sourceScheme + " is not supported yet.")
	}

	//Start or connect to a docker daemon
	dock := StartDocker(settings)

	//Prepare cache
	switch sourceScheme {
		case "graph":
			//Import the latest lineage
			dock.Import(sourceGraph.Load(config.Image), config.Image, "latest")
		case "file":
			//Load image from file
			dock.ImportFromFilenameTagstring(sourcePath, config.Image)
	}

	//Run the container and wait for it to finish
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
