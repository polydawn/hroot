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
type Image struct {
	scheme string    //URI scheme
	path   string    //URI path
	graph *dex.Graph //Graph (if desired)
}

//Holds everything needed to run a docket command
type Docket struct {
	//Source and destination URIs
	source    Image
	dest      Image

	//Docker instance
	dock      *crocker.Dock

	//Container instance
	container *crocker.Container

	//Configuration
	settings  *confl.ConfigLoad
	config    *crocker.ContainerConfig
	image     string //Stored separately so we don't modify config if needed later for export.
}


//Create a docket struct
func LoadDocket(args []string, defaultTarget, sourceURI, destURI string) *Docket {
	//Parse config file
	target   := GetTarget(args, defaultTarget)
	settings := confl.NewConfigLoad(".")
	config := settings.GetConfig(target)

	//Parse input URI
	sourceScheme, sourcePath := ParseURI(sourceURI)

	//Docket struct
	docket := &Docket{
		source: Image{
			scheme: sourceScheme,
			path:   sourcePath,
		},
		settings: settings,
		config: config,
		image: config.Image, //Stored separately (see above)
	}

	//If there's a destination URI, parse that as well
	if destURI != "" {
		destScheme, destPath     := ParseURI(destURI)

		docket.dest = Image{
			scheme: destScheme,
			path:   destPath,
		}
	}

	return docket
}

//Prepare the docket input
func (d *Docket) PrepareInput() {
	switch d.source.scheme {
		case "graph":
			//Look up the graph, and clear any unwanted state
			d.source.graph = dex.NewGraph(d.settings.Graph)
			Println("Opening source repository", d.source.graph.GetDir())
			d.source.graph.Cleanse()
		case "file":
			//If the user did not specify an image path, set one
			if d.source.path == "" {
				d.source.path = "./image.tar"
			}
		case "index":
			//If pulling from the index, use the index key instead (protect URL namespace from docker)
			d.image = d.config.Index
	}
}

//Prepare the docket output
func (d *Docket) PrepareOutput() error {
	switch d.dest.scheme {
		case "graph":
			//Look up the graph, and clear any unwanted state
			d.dest.graph = dex.NewGraph(d.settings.Graph)

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
				return Errorf("Tar location is same for source and destination: " + d.source.path)
			}
		case "index":
			return Errorf("Destination " + d.dest.scheme + " is not supported yet.")
	}

	return nil
}

//Starts the docker daemon
func (d *Docket) StartDocker() {
	d.dock = crocker.NewDock(d.settings.Dock)

	//Announce the docker
	if d.dock.IsChildProcess() {
		Println("Started a docker in", d.dock.Dir())
	} else {
		Println("Connecting to docker", d.dock.Dir())
	}

}

//Behavior when docker cache has the image
func (d *Docket) prepareCacheWithImage() {
	image := d.config.Image

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
				"\n"   + "Warning: your docker cache already has " + d.config.Index + " loaded." +
				"\n"   + "Pulling from the index may modify the saved image." +
				"\n\n" + "Continuing in 10 seconds, hit Ctrl-C to cancel...")
			time.Sleep(time.Second * 10)
	}
}

//Behavior when docker cache doesn't have the image
func (d *Docket) prepareCacheWithoutImage() error {
	image := d.config.Image

	switch d.source.scheme {
		case "docker":
			//Can't continue; specified docker as source and it doesn't have it
			return Errorf("Docker does not have " + image + " loaded.")
		case "graph":
			//Check if the image is in the graph
			if !d.source.graph.HasBranch(image) {
				return Errorf("Image branch name " + image + " not found in graph.")
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
			if err != nil {
				return Errorf("Import from graph failed: " + err.Error())
			}

			wait.Wait() //Block on our gofunc
	}

	return nil
}

//Prepare the docker cache
func (d *Docket) PrepareCache() error {
	//Behavior based on if the docker cache already has an image
	if d.dock.CheckCache(d.config.Image) {
		d.prepareCacheWithImage()
	} else {
		err := d.prepareCacheWithoutImage()
		if err != nil { return err }
	}

	//Now that's taken care of, normal behavior
	switch d.source.scheme  {
		case "file":
			d.dock.ImportFromFilenameTagstring(d.source.path, d.config.Image) //Load image from file
		case "index":
			d.dock.Pull(d.config.Index) //Download from index
	}

	return nil
}

//Lanuch the container and wait for it to complete
func (d *Docket) Launch() {
	Println("Launching container.")
	c := d.config

	//Map the struct values to crocker function params
	d.container = crocker.Launch(d.dock, d.image, c.Command, c.Attach, c.Privileged, c.Folder, c.DNS, c.Mounts, c.Ports, c.Environment)

	//Wait for container
	d.container.Wait()
}

//Prepare the docket export
func (d *Docket) ExportBuild() error {
	switch d.dest.scheme {
		case "graph":
			//Create new branches as needed
			d.dest.graph.PreparePublish(d.config.Image, d.config.Upstream)

			// Export a tar of the filesystem
			exportReader, exportWriter := io.Pipe()
			go d.container.Export(exportWriter)

			// Use guitar to write the tar's contents to the graph
			err := stream.ExportToFilesystem(exportReader, d.dest.graph.GetDir())
			if err != nil { return err }

			// Commit changes
			Println("Comitting to graph...")
			d.dest.graph.Publish(d.config.Image, d.config.Upstream)
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
	name, tag := crocker.SplitImageName(d.config.Image)
	Println("Exporting to docker cache:", name, tag)
	d.container.Commit(name, tag)

	//Remove the container from cache if desired
	if d.config.Purge {
		d.container.Purge()
	}

	//Stop the docker daemon if it's a child process
	d.dock.Slay()
}
