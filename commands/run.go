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
	hroot := LoadDocket(args, DefaultRunTarget, opts.Source, "")
	Println("Running", hroot.image.Name)
	hroot.PrepareInput()

	//Start or connect to a docker daemon
	hroot.StartDocker()
	hroot.PrepareCache()
	hroot.Launch()

	hroot.Cleanup()
	return nil
}
