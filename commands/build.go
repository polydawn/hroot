package commands

import (
	. "fmt"
)

type BuildCmdOpts struct {
	Source      string `short:"s" long:"source"      description:"Container source.      (default: graph)"`
	Destination string `short:"d" long:"destination" description:"Container destination. (default: graph)"`
	NoOp bool          `long:"noop" description:"Set the container command to /bin/true."`
	Epoch       bool   `long:"epoch" description:"Force all file modtimes to epoch."`
}

const DefaultBuildTarget = "build"

//Transforms a container
func (opts *BuildCmdOpts) Execute(args []string) error {
	//Check if the user explicitly asked for a source type
	sourceEmpty := false
	if opts.Source == "" {
		sourceEmpty = true
		opts.Source = "graph" //set this so LoadHroot runs correctly
	}

	//Load settings
	hroot := LoadHroot(args, DefaultBuildTarget, opts.Source, opts.Destination)

	//We're building; launch upstream image
	hroot.launchImage = hroot.image.Upstream
	Println("Building from", hroot.image.Upstream, "to", hroot.image.Name)

	//If the user did not explicitly ask for a source type, try a smart default
	if sourceEmpty && hroot.image.Index != "" {
		hroot.source.scheme = "index"
	} else if sourceEmpty && hroot.image.Upstream != "" {
		hroot.source.scheme = "graph"
	}

	//If desired, set the command to /bin/true and do not modify destination image name
	//We'd love to not launch the container at all, but docker's export is completely broken.
	// 'docker export ubuntu' --> 'Error: No such container: ubuntu' --> :(
	if opts.NoOp {
		hroot.settings.Command = []string{ "/bin/true" }
	}

	//Prepare source & destination
	hroot.PrepareInput()
	hroot.PrepareOutput()

	//Start or connect to a docker daemon
	hroot.StartDocker()
	hroot.PrepareCache()
	hroot.Launch()

	//Perform any destination operations required
	hroot.ExportBuild(opts.Epoch)

	hroot.Cleanup()
	return nil
}
