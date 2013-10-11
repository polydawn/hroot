package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"path/filepath"
	"polydawn.net/docket/confl"
	"polydawn.net/docket/crocker"
	"polydawn.net/docket/dex"
)

type unpackCmdOpts struct {
	Graph string `long:"graph" optional:"true" default:"defaults to ./graph/" description:"Path to the working directory of the git repo storing the image graph."`
}

func (opts *unpackCmdOpts) Execute(args []string) error {
	// Process positional args
	if len(args) != 1 {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: "expected one positional argument, for which image to import",
		}
	}
	image := args[0]

	return WithDocker(func(dock *crocker.Dock, settings *confl.ConfigLoad) error {
		return Unpack(dock, settings, image, opts.Graph)
	})
}

func Unpack(dock *crocker.Dock, settings *confl.ConfigLoad, image string, graphDir string) error {
	//If the user asked for a specific graph folder, use it, else find one
	if graphDir == "" {
		graphDir = settings.Graph
	}
	graphDir, err := filepath.Abs(graphDir)
	if err != nil {
		return fmt.Errorf("expected graph to exist: %v", err.Error())
	}

	// Look up the graph
	graph := dex.NewGraph(graphDir)

	// Import the latest of the base lineage
	dock.Import(graph.Load(image), image, "latest")

	return nil
}
