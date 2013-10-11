package main

import (
	// "github.com/jessevdk/go-flags"
	"polydawn.net/docket/confl"
	"polydawn.net/docket/crocker"
)

type runCmdOpts struct{}

func (opts *runCmdOpts) Execute(args []string) error {
	//Get the target
	target := ""
	if len(args) == 1 {
		target = args[0]
	} else {
		target = "default"
	}

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
