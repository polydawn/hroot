package commands

import (
	. "fmt"
	"io"
	"polydawn.net/docket/confl"
	"polydawn.net/docket/crocker"
	"polydawn.net/docket/dex"
	. "polydawn.net/docket/util"
	"time"
)

type BuildCmdOpts struct {
	Source      string `short:"s" long:"source"      description:"Container source.      (default: graph)"`
	Destination string `short:"d" long:"destination" description:"Container destination. (default: graph)"`
	NoOp bool          `long:"noop" description:"Set the container command to /bin/true."`
}

const DefaultBuildTarget = "build"

//Transforms a container
func (opts *BuildCmdOpts) Execute(args []string) error {
	//Get configuration
	target   := GetTarget(args, DefaultBuildTarget)
	settings := confl.NewConfigLoad(".")
	config := settings.GetConfig(target)
	saveAs := settings.GetDefaultImage()
	var sourceGraph, destinationGraph *dex.Graph

	Println("Building from", config.Image, "to", saveAs)

	//If desired, set the command to /bin/true.
	//We'd love to not launch the container at all, but docker's export is completely broken.
	// 'docker export ubuntu' --> 'Error: No such container: ubuntu' --> :(
	if opts.NoOp {
		config.Command = []string{ "/bin/true" }
	}

	//Right now, go-flags' default announation does not appear to work when in a sub-command.
	//	Will investigate and hopefully remove this later.
	if opts.Source == "" {
		opts.Source = "graph"
	}
	if opts.Destination == "" {
		opts.Destination = "graph"
	}

	//Parse input/output URIs
	sourceScheme, sourcePath           := ParseURI(opts.Source)
	destinationScheme, destinationPath := ParseURI(opts.Destination)

	//Prepare input
	switch sourceScheme {
		case "graph":
			//Look up the graph, and clear any unwanted state
			sourceGraph = dex.NewGraph(settings.Graph)
			sourceGraph.Cleanse()
		case "file":
			//If the user did not specify an image path, set one
			if sourcePath == "" {
				sourcePath = "./image.tar"
			}
	}

	//Prepare output
	name, tag := crocker.SplitImageName(saveAs)
	switch destinationScheme {
		case "docker":
			//Nothing required here until container has ran
		case "graph":
			//Look up the graph, and clear any unwanted state
			destinationGraph = dex.NewGraph(settings.Graph)

			//Cleanse the graph unless it'd be redundant.
			if !(sourceScheme == "graph" && sourceGraph.GetDir() != destinationGraph.GetDir()) {
				destinationGraph.Cleanse()
			}
		case "file":
			//If the user did not specify an image path, set one
			if destinationPath == "" {
				destinationPath = "./image.tar"
			}

			//If the user is insane and wants to overwrite his source tar, stop him.
			//	Not at all robust (absolute paths? what are those? etc)
			if sourceScheme == "file" && sourcePath == destinationPath {
				return Errorf("Tar location is same for source and destination: " + sourcePath)
			}
		case "index":
			return Errorf("Destination " + destinationScheme + " is not supported yet.")
	}

	//Start or connect to a docker daemon
	dock := StartDocker(settings)

	//Prepare cache
	switch sourceScheme {
		case "docker":
			//Check that docker has the image needed
			if !dock.CheckCache(config.Image) {
				return Errorf("Docker does not have " + config.Image + " loaded.")
			}
		case "graph":
			//Import the latest lineage
			if dock.CheckCache(config.Image) {
				Println("Docker already has", config.Image, "loaded, not importing from graph.")
			} else {
				dock.Import(sourceGraph.Load(config.Image), config.Image, "latest")
			}
		case "file":
			//If docker already has the image loaded, warn & wait first!
			if dock.CheckCache(config.Image) {
				Println(
					"\n"   + "Warning: your docker cache already has " + config.Image + " loaded." +
					"\n"   + "Importing will overwrite the saved image." +
					"\n\n" + "Continuing in 10 seconds, hit Ctrl-C to cancel..")
				time.Sleep(time.Second * 10)
			}

			//Load image from file
			dock.ImportFromFilenameTagstring(sourcePath, config.Image)
		case "index":
			//If docker already has the image loaded, warn & wait first!
			if dock.CheckCache(config.Image) {
				Println(
					"\n"   + "Warning: your docker cache already has " + config.Image + " loaded." +
					"\n"   + "Pulling from the index may modify the saved image." +
					"\n\n" + "Continuing in 10 seconds, hit Ctrl-C to cancel..")
				time.Sleep(time.Second * 10)
			}

			//Download from index
			dock.Pull(config.Image)
	}

	//Launch the container and wait for it to finish
	container := Launch(dock, config)
	container.Wait()

	//Perform any destination operations required
	switch destinationScheme {
		//If we're not exporting to the graph, there is no commit hash from which to generate a tag.
		//	Thus the docker import will have either a static tag (from docker.toml) or the default 'latest' tag.
		case "docker":
			Println("Exporting to", name, tag)
			container.Commit(name, tag)
		case "graph":
			// Export a tar of the filesystem
			exportStreamOut, exportStreamIn := io.Pipe()
			go container.Export(exportStreamIn)

			// Commit it to the image graph
			destinationGraph.Publish(exportStreamOut, saveAs, config.Image)
		case "file":
			//Export a tar
			Println("Exporting to", destinationPath)
			container.ExportToFilename(destinationPath)
	}

	//Remove if desired
	if config.Purge {
		container.Purge()
	}

	//Stop the docker daemon if it's a child process
	dock.Slay()

	return nil
}
