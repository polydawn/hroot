package commands

import (
	. "fmt"
)

type BuildCmdOpts struct {
	Source      string `short:"s" long:"source"      description:"Container source.      (default: graph)"`
	Destination string `short:"d" long:"destination" description:"Container destination. (default: graph)"`
	NoOp bool          `long:"noop" description:"Set the container command to /bin/true and do not modify destination image name."`
}

const DefaultBuildTarget = "build"

//Transforms a container
func (opts *BuildCmdOpts) Execute(args []string) error {
	//Load settings
	docket := LoadDocket(args, DefaultBuildTarget, opts.Source, opts.Destination)
	docket.image = docket.config.Upstream //We're building; launch upstream image
	Println("Building from", docket.config.Upstream, "to", docket.config.Image)

	//If desired, set the command to /bin/true and do not modify destination image name
	//We'd love to not launch the container at all, but docker's export is completely broken.
	// 'docker export ubuntu' --> 'Error: No such container: ubuntu' --> :(
	if opts.NoOp {
		docket.config.Command = []string{ "/bin/true" }
		docket.config.Image = docket.config.Upstream
	}

	//Prepare source & destination
	docket.PrepareInput()
	docket.PrepareOutput()

	//Start or connect to a docker daemon
	docket.StartDocker()
	docket.PrepareCache()
	docket.Launch()

	//Perform any destination operations required
	docket.ExportBuild()

	docket.Cleanup()
	return nil
}
