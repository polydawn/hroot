package commands

//Helper struct holds all the state & shared functionality you need to run a docket command.

import (
	. "fmt"
	"io"
	"sync"
	"time"
	"polydawn.net/docket/confl"
	"polydawn.net/docket/crocker"
	"polydawn.net/docket/dex"
	. "polydawn.net/docket/util"
	"polydawn.net/guitar/stream"
)

//Holds everything needed to load/save docker images
type ImagePath struct {
	scheme string    //URI scheme
	path   string    //URI path
	graph *dex.Graph //Graph (if desired)
}

//Holds everything needed to run a docket command
type Docket struct {
	//Source and destination URIs
	source    ImagePath
	dest      ImagePath

	//Docker instance
	dock      *crocker.Dock

	//Container instance
	container *crocker.Container

	//Configuration
	folders  confl.Folders
	image    confl.Image
	settings confl.Container
	launchImage     string //Stored separately so we don't modify config if needed later for export.
}

//Create a docket struct
func LoadDocket(args []string, defaultTarget, sourceURI, destURI string) *Docket {
	//If there was no target specified, override it
	target   := GetTarget(args, defaultTarget)

	//Load toml parser
	parser := &confl.TomlConfigParser{}

	//Parse config file
	configuration, folders := confl.LoadConfigurationFromDisk(".", parser)
	config := configuration.GetTargetContainer(target)

	//Docket struct
	d := &Docket {
		folders:     *folders,
		image:       configuration.Image,
		settings:    config,
		launchImage: configuration.Image.Name, //Stored separately (see above)
	}

	//Parse input URI
	sourceScheme, sourcePath := ParseURI(sourceURI)
	d.source = ImagePath {
		scheme: sourceScheme,
		path:   sourcePath,
	}

	//If there's a destination URI, parse that as well
	if destURI != "" {
		destScheme, destPath     := ParseURI(destURI)

		d.dest = ImagePath {
			scheme: destScheme,
			path:   destPath,
		}
	}

	//Image name required
	if d.launchImage == "" {
		ExitGently("No image name specified.")
	}

	//Specifying a command in the settings section has confusing implications
	if len(d.settings.Command) > 0 {
		ExitGently("Cannot specify a command in settings; instead, put them in a target!")
	}

	return d
}

//Prepare the docket input
func (d *Docket) PrepareInput() {
	switch d.source.scheme {
		case "graph":
			//Look up the graph, and clear any unwanted state
			d.source.graph = dex.NewGraph(d.folders.Graph)
			Println("Opening source repository", d.source.graph.GetDir())
			d.source.graph.Cleanse()
		case "file":
			//If the user did not specify an image path, set one
			if d.source.path == "" {
				d.source.path = "./image.tar"
			}
		case "index":
			//If pulling from the index, use the index key instead (protect URL namespace from docker)
			d.launchImage = d.image.Index
	}
}

//Prepare the docket output
func (d *Docket) PrepareOutput() {
	switch d.dest.scheme {
		case "graph":
			//Look up the graph, and clear any unwanted state
			d.dest.graph = dex.NewGraph(d.folders.Graph)

			//Cleanse the graph unless it'd be redundant.
			if !(d.source.scheme == "graph" && d.source.graph.GetDir() == d.dest.graph.GetDir()) {
				Println("Opening destination repository", d.dest.graph.GetDir())
				d.dest.graph.Cleanse()
			}
		case "file":
			//If the user did not specify an image path, set one
			if d.dest.path == "" {
				d.dest.path = "./image.tar"
			}

			//If the user is insane and wants to overwrite his source tar, stop him.
			//	Not at all robust (absolute paths? what are those? etc)
			if d.source.scheme == "file" && d.source.path == d.dest.path {
				ExitGently("Tar location is same for source and destination:", d.source.path)
			}
		case "index":
			ExitGently("Destination", d.dest.scheme, "is not supported yet.")
	}
}

//Starts the docker daemon
func (d *Docket) StartDocker() {
	d.dock = crocker.NewDock(d.folders.Dock)

	//Announce the docker
	if d.dock.IsChildProcess() {
		Println("Started a docker in", d.dock.Dir())
	} else {
		Println("Connecting to docker", d.dock.Dir())
	}

}

//Behavior when docker cache has the image
func (d *Docket) prepareCacheWithImage() {
	image := d.image.Name

	switch d.source.scheme {
		case "graph":
			Println("Docker already has", image, "loaded, not importing from graph.")
		case "file":
			Println(
				"\n"   + "Warning: your docker cache already has " + image + " loaded." +
				"\n"   + "Importing will overwrite the saved image." +
				"\n\n" + "Continuing in 10 seconds, hit Ctrl-C to cancel..")
			time.Sleep(time.Second * 10)
		case "index":
			Println(
				"\n"   + "Warning: your docker cache already has " + d.image.Index + " loaded." +
				"\n"   + "Pulling from the index may modify the saved image." +
				"\n\n" + "Continuing in 10 seconds, hit Ctrl-C to cancel...")
			time.Sleep(time.Second * 10)
	}
}

//Behavior when docker cache doesn't have the image
func (d *Docket) prepareCacheWithoutImage() {
	image := d.image.Name

	switch d.source.scheme {
		case "docker":
			//Can't continue; specified docker as source and it doesn't have it
			ExitGently("Docker does not have", image, "loaded.")
		case "graph":
			//Check if the image is in the graph
			if !d.source.graph.HasBranch(image) {
				ExitGently("Image branch name", image, "not found in graph.")
			}

			//Pipe for I/O, and a waitgroup to make async action blocking
			importReader, importWriter := io.Pipe()
			var wait sync.WaitGroup
			wait.Add(1)

			//Closure to run the docker import
			go func() {
				d.dock.Import(importReader, image, "latest")
				wait.Done()
			}()

			//Run the guitar import
			err := stream.ImportFromFilesystem(importWriter, d.source.graph.GetDir())
			if err != nil { ExitGently("Import from graph failed:", err) }

			wait.Wait() //Block on our gofunc
	}
}

//Prepare the docker cache
func (d *Docket) PrepareCache() {
	//Behavior based on if the docker cache already has an image
	if d.dock.CheckCache(d.image.Name) {
		d.prepareCacheWithImage()
	} else {
		d.prepareCacheWithoutImage()
	}

	//Now that's taken care of, normal behavior
	switch d.source.scheme  {
		case "file":
			d.dock.ImportFromFilenameTagstring(d.source.path, d.image.Name) //Load image from file
		case "index":
			d.dock.Pull(d.image.Index) //Download from index
	}
}

//Lanuch the container and wait for it to complete
func (d *Docket) Launch() {
	Println("Launching container.")
	c := d.settings

	//Map the struct values to crocker function params
	d.container = crocker.Launch(d.dock, d.launchImage, c.Command, c.Attach, c.Privileged, c.Folder, c.DNS, c.Mounts, c.Ports, c.Environment)

	//Wait for container
	d.container.Wait()
}

//Prepare the docket export
func (d *Docket) ExportBuild() error {
	switch d.dest.scheme {
		case "graph":
			//Create new branches as needed
			d.dest.graph.PreparePublish(d.image.Name, d.image.Upstream)

			// Export a tar of the filesystem
			exportReader, exportWriter := io.Pipe()
			go d.container.Export(exportWriter)

			// Use guitar to write the tar's contents to the graph
			err := stream.ExportToFilesystem(exportReader, d.dest.graph.GetDir())
			if err != nil { return err }

			// Commit changes
			Println("Comitting to graph...")
			d.dest.graph.Publish(d.image.Name, d.image.Upstream)
		case "file":
			//Export a tar
			Println("Exporting to", d.dest.path)
			d.container.ExportToFilename(d.dest.path)
	}

	return nil
}

//Clean up after ourselves
func (d *Docket) Cleanup() {
	//Commit the image name to the docker cache.
	//	This is so if you run:
	//		docket build -s index  -d graph --noop
	//		docket build -s docker -d graph
	//	Docker will already know about your (much cooler) image name :)
	name, tag := crocker.SplitImageName(d.image.Name)
	Println("Exporting to docker cache:", name, tag)
	d.container.Commit(name, tag)

	//Remove the container from cache if desired
	if d.settings.Purge {
		d.container.Purge()
	}

	//Stop the docker daemon if it's a child process
	d.dock.Slay()
}
