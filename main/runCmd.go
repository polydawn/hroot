package main

import (
	"github.com/jessevdk/go-flags"
	"polydawn.net/dockctrl/crocker"
	"polydawn.net/dockctrl/confl"
)

type runCmdOpts struct{}

func (opts *runCmdOpts) Execute(args []string) error {
	//Get the target
	if len(args) != 1 {
		return &flags.Error{
			Type: flags.ErrExpectedArgument,
			Message: "expected one positional argument, for which target to launch",
		}
	}
	target := args[0]

	return WithDocker(func(dock *crocker.Dock, settings *confl.ConfigLoad) error {
		return Run(dock, settings, target)
	})
}

//Launches a docker
func Run(dock *crocker.Dock, settings *confl.ConfigLoad, target string) error {
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
