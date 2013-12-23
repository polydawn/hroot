package commands

import (
	. "fmt"
)

type RunCmdOpts struct {
	Source      string `short:"s" long:"source" default:"graph" description:"Container source."`
}

const DefaultRunTarget = "run"

//Runs a container
func (opts *RunCmdOpts) Execute(args []string) error {
	//Load settings
	docket := LoadDocket(args, DefaultRunTarget, opts.Source, "")
	Println("Running", docket.image.Name)
	docket.PrepareInput()

	//Start or connect to a docker daemon
	docket.StartDocker()
	docket.PrepareCache()
	docket.Launch()

	docket.Cleanup()
	return nil
}
