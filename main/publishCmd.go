package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"io"
	"path/filepath"
	"polydawn.net/dockctrl/confl"
	"polydawn.net/dockctrl/crocker"
	"polydawn.net/dockctrl/dex"
)

type publishCmdOpts struct {
	Graph string `long:"graph" optional:"true" default:"defaults to ./graph/" description:"Path to the working directory of the git repo storing the image graph."`
}

func (opts *publishCmdOpts) Execute(args []string) error {
	//Get the target
	if len(args) != 1 {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: "expected one positional argument, for which target to launch",
		}
	}
	target := args[0]

	//Assign defaults over any zero values in the opts, and canonicalize values
	var err error
	if opts.Graph == "" {
		opts.Graph = "./graph"
	}
	opts.Graph, err = filepath.Abs(opts.Graph)
	if err != nil {
		return fmt.Errorf("expected graph to exist: %v", err.Error())
	}

	//Look up the graph
	graph := dex.NewGraph(opts.Graph)

	return WithDocker(func(dock *crocker.Dock, settings *confl.ConfigLoad) error {
		return Publish(dock, settings, target, graph)
	})
}

func Publish(dock *crocker.Dock, settings *confl.ConfigLoad, target string, graph *dex.Graph) error {
	// Cleanse git working tree in case of unwanted unknown state
	graph.Cleanse()

	// Get configuration
	config := settings.GetConfig(target)
	saveAs := settings.GetConfig(confl.DefaultTarget).Image

	// Import the latest of the base lineage
	dock.Import(graph.Load(config.Image), config.Image, "latest")

	// Launch the transition and wait for it to finish
	container := Launch(dock, config)
	container.Wait()

	// Export a tar of the filesystem
	exportStreamOut, exportStreamIn := io.Pipe()
	go container.Export(exportStreamIn)

	// Commit it to the image graph
	graph.Publish(exportStreamOut, saveAs, config.Image)

	// Destroy temp container.  You just exported it as an image, what could you possibly need it for.
	container.Purge()

	return nil
}
