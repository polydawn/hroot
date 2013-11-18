package commands

import (
	. "fmt"
	"io"
	"polydawn.net/docket/confl"
	"polydawn.net/docket/crocker"
	"polydawn.net/docket/dex"
	. "polydawn.net/docket/util"
	"polydawn.net/guitar/stream"
	"sync"
	"time"
)

type BuildCmdOpts struct {
	Source      string `short:"s" long:"source"      description:"Container source.      (default: graph)"`
	Destination string `short:"d" long:"destination" description:"Container destination. (default: graph)"`
	NoOp bool          `long:"noop" description:"Set the container command to /bin/true and do not modify destination image name."`
}

const DefaultBuildTarget = "build"

//Transforms a container
func (opts *BuildCmdOpts) Execute(args []string) error {
	//Get configuration
	target   := GetTarget(args, DefaultBuildTarget)
	settings := confl.NewConfigLoad(".")
	config := settings.GetConfig(target)
	var sourceGraph, destinationGraph *dex.Graph

	Println("Building from", config.Upstream, "to", config.Image)

	//If desired, set the command to /bin/true and do not modify destination image name
	//We'd love to not launch the container at all, but docker's export is completely broken.
	// 'docker export ubuntu' --> 'Error: No such container: ubuntu' --> :(
	if opts.NoOp {
		config.Command = []string{ "/bin/true" }
		config.Image = config.Upstream
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

	//Copy of config for 'docker run'
	runConfig := config
	runConfig.Image = config.Upstream

	//Prepare input
	switch sourceScheme {
		case "graph":
			//Look up the graph, and clear any unwanted state
			sourceGraph = dex.NewGraph(settings.Graph)
			Println("Opening source repository", sourceGraph.GetDir())
			sourceGraph.Cleanse()
		case "file":
			//If the user did not specify an image path, set one
			if sourcePath == "" {
				sourcePath = "./image.tar"
			}
		case "index":
			//If pulling from the index, use the index key instead (protect URL namespace from docker)
			runConfig.Image = config.Index
	}

	//Prepare output
	switch destinationScheme {
		case "docker":
			//Nothing required here until container has ran
		case "graph":
			//Look up the graph, and clear any unwanted state
			destinationGraph = dex.NewGraph(settings.Graph)

			//Cleanse the graph unless it'd be redundant.
			if !(sourceScheme == "graph" && sourceGraph.GetDir() == destinationGraph.GetDir()) {
				Println("Opening destination repository", destinationGraph.GetDir())
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
	hasImage := dock.CheckCache(runConfig.Image)
	switch sourceScheme {
		case "docker":
			//Check that docker has the image needed
			if !hasImage {
				return Errorf("Docker does not have " + runConfig.Image + " loaded.")
			}
		case "graph":
			//Import a tar from the filesystem
			if hasImage {
				Println("Docker already has", runConfig.Image, "loaded, not importing from graph.")
			} else {
				//Check if the image is in the graph
				if !sourceGraph.HasBranch(runConfig.Image) {
					return Errorf("Image branch name " + runConfig.Image + " not found in graph.")
				}

				//Run import
				importReader, importWriter := io.Pipe()
				var wait sync.WaitGroup
				wait.Add(1)
				go func() {
					dock.Import(importReader, runConfig.Image, "latest")
					wait.Done()
				}()
				err := stream.ImportFromFilesystem(importWriter, sourceGraph.GetDir())
				if err != nil {
					return Errorf("Import from graph failed: " + err.Error())
				}

				wait.Wait()
			}
		case "file":
			//If docker already has the image loaded, warn & wait first!
			if hasImage {
				Println(
					"\n"   + "Warning: your docker cache already has " + runConfig.Image + " loaded." +
					"\n"   + "Importing will overwrite the saved image." +
					"\n\n" + "Continuing in 10 seconds, hit Ctrl-C to cancel..")
				time.Sleep(time.Second * 10)
			}

			//Load image from file
			dock.ImportFromFilenameTagstring(sourcePath, runConfig.Image)
		case "index":
			//If docker already has the image loaded, warn & wait first!
			if hasImage {
				Println(
					"\n"   + "Warning: your docker cache already has " + runConfig.Index + " loaded." +
					"\n"   + "Pulling from the index may modify the saved image." +
					"\n\n" + "Continuing in 10 seconds, hit Ctrl-C to cancel...")
				time.Sleep(time.Second * 10)
			}

			//Download from index
			dock.Pull(config.Index)
	}

	//Launch the container and wait for it to finish
	container := Launch(dock, runConfig)
	container.Wait()

	//Perform any destination operations required
	name, tag := crocker.SplitImageName(config.Image)
	switch destinationScheme {
		case "graph":
			//Create new branches as needed
			destinationGraph.PreparePublish(config.Image, config.Upstream)

			// Export a tar of the filesystem
			exportReader, exportWriter := io.Pipe()
			go container.Export(exportWriter)

			// Use guitar to write the tar's contents to the graph
			err := stream.ExportToFilesystem(exportReader, destinationGraph.GetDir())
			if err != nil {
				return err
			}

			// Commit changes
			Println("Comitting to graph...")
			destinationGraph.Publish(config.Image, config.Upstream)
		case "file":
			//Export a tar
			Println("Exporting to", destinationPath)
			container.ExportToFilename(destinationPath)
	}

	//Commit the image name to the cache.
	//	This is so if you run:
	//		docket build -s index  -d graph --noop
	//		docket build -s docker -d graph
	//	Docker will already know about your (much cooler) image name :)
	Println("Exporting to", name, tag)
	container.Commit(name, tag)

	//Remove if desired
	if config.Purge {
		container.Purge()
	}

	//Stop the docker daemon if it's a child process
	dock.Slay()

	return nil
}
