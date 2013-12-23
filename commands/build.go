package commands

import (
	. "fmt"
)

type BuildCmdOpts struct {
	Source      string `short:"s" long:"source"      description:"Container source.      (default: graph)"`
	Destination string `short:"d" long:"destination" description:"Container destination. (default: graph)"`
	NoOp bool          `long:"noop" description:"Set the container command to /bin/true."`
}

const DefaultBuildTarget = "build"

//Transforms a container
func (opts *BuildCmdOpts) Execute(args []string) error {
	//Check if the user explicitly asked for a source type
	sourceEmpty := false
	if opts.Source == "" {
		sourceEmpty = true
		opts.Source = "graph" //set this so LoadDocket runs correctly
	}

	//Load settings
	docket := LoadDocket(args, DefaultBuildTarget, opts.Source, opts.Destination)

	//We're building; launch upstream image
	docket.launchImage = docket.image.Upstream
	Println("Building from", docket.image.Upstream, "to", docket.image.Name)

	//If the user did not explicitly ask for a source type, try a smart default
	if sourceEmpty && docket.image.Index != "" {
		docket.source.scheme = "index"
	} else if sourceEmpty && docket.image.Upstream != "" {
		docket.source.scheme = "graph"
	}

	//If desired, set the command to /bin/true and do not modify destination image name
	//We'd love to not launch the container at all, but docker's export is completely broken.
	// 'docker export ubuntu' --> 'Error: No such container: ubuntu' --> :(
	if opts.NoOp {
		docket.settings.Command = []string{ "/bin/true" }
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
