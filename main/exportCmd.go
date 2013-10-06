package main

import (
	"github.com/jessevdk/go-flags"
	"polydawn.net/dockctrl/crocker"
	"polydawn.net/dockctrl/confl"
)

type exportCmdOpts struct{}

func (opts *exportCmdOpts) Execute(args []string) error {
	return WithDocker(Export, args)
}

const ExportPath = "./" //Where to export docker images

//Exports the result of a target into docker.
func Export(dock *crocker.Dock, settings *confl.ConfigLoad, args []string) error {
	//Get the target
	if len(args) != 1 {
		return &flags.Error{
			Type: flags.ErrExpectedArgument,
			Message: "expected one positional argument, for which target to launch",
		}
	}
	target := args[0]

	//Get configuration
	config := settings.GetConfig(target)
	saveAs := settings.GetConfig(confl.DefaultTarget).Image

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
