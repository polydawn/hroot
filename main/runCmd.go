package main

import (
	"polydawn.net/docket/confl"
	// "polydawn.net/docket/crocker"
)

type runCmdOpts struct {
	Source      string `short:"s" long:"source"      default:"graph" description:"Container source."`
}

const DefaultRunTarget = "default"

//Runs a container
func (opts *runCmdOpts) Execute(args []string) error {
	//Get configuration
	target   := GetTarget(args, DefaultRunTarget)
	settings := confl.NewConfigLoad(".")
	config   := settings.GetConfig(target)

	//Start or connect to a docker daemon
	dock := StartDocker(settings)

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
