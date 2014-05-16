package commands

//Helper struct holds all the state & shared functionality you need to run a hroot command.

import (
	. "fmt"
	"os"
	"time"
	"strings"
	"polydawn.net/hroot/conf"
	"polydawn.net/hroot/crocker"
	"polydawn.net/hroot/dex"
	. "polydawn.net/hroot/util"
	guitarconf "polydawn.net/guitar/conf"
)

//Holds everything needed to load/save docker images
type ImagePath struct {
	scheme string    //URI scheme
	path   string    //URI path
	graph *dex.Graph //Graph (if desired)
}

//Holds everything needed to run a hroot command
type Hroot struct {
	//Source and destination URIs
	source    ImagePath
	dest      ImagePath

	//Docker instance
	dock      *crocker.Dock

	//Container instance
	container *crocker.Container

	//Configuration
	folders  conf.Folders
	image    conf.Image
	settings conf.Container
	launchImage     string //Stored separately so we don't modify config if needed later for export.
}

//Create a hroot struct
func LoadHroot(args []string, defaultTarget, sourceURI, destURI string) *Hroot {
	//If there was no target specified, override it
	target   := GetTarget(args, defaultTarget)

	//Load toml parser
	parser := &conf.TomlConfigParser{}

	//Parse config file
	configuration, folders := conf.LoadConfigurationFromDisk(".", parser)
	config := configuration.Targets[target]

	//Hroot struct
	d := &Hroot {
		folders:     *folders,
		image:       configuration.Image,
		settings:    config,
		launchImage: configuration.Image.Name, //Stored separately (see above)
	}

	//If the user did not explicitly ask for a source type, try a smart default
	if sourceURI == "" {
		if d.image.Index != "" {
			sourceURI = "index"
		} else if d.image.Upstream != "" {
			sourceURI = "graph"
		}
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
	if len(configuration.Settings.Command) > 0 {
		ExitGently("Cannot specify a command in settings; instead, put them in a target!")
	}

	return d
}

//Prepare the hroot input
func (d *Hroot) PrepareInput() {

	//If you're using an index key with a non-index source, or upstream key with index source, reject.
	//Runs here (not LoadHroot) so commands have a chance to change settings.
	if d.source.scheme == "index" && d.image.Index == "" {
		ExitGently("You asked to pull from the index but have no index key configured.")
	} else if d.source.scheme != "index" && d.image.Upstream == "" {
		if d.source.scheme == "docker" {
			Println("Running an index image from docker cache.")
		} else {
			ExitGently("You asked to run from from", d.source.scheme, "but have no upstream key configured.")
		}
	}

	switch d.source.scheme {
		case "graph":
			//Look up the graph, and clear any unwanted state
			d.source.graph = dex.NewGraph(d.folders.Graph)
			Println("Opening source repository")
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

//Prepare the hroot output
func (d *Hroot) PrepareOutput() {
	switch d.dest.scheme {
		case "graph":
			//Look up the graph, and clear any unwanted state
			d.dest.graph = dex.NewGraph(d.folders.Graph)

			//If the user's git config isn't ready, we want to tell them *before* building.
			if !d.dest.graph.IsConfigReady() {
				ExitGently("\n" +
					"Git could not find a user name & email."                 + "\n"   +
					"You'll need to set up git with the following commands:"  + "\n\n" +
					"git config --global user.email \"you@example.com\""      + "\n"   +
					"git config --global user.name \"Your Name\"")
			}

			//Cleanse the graph unless it'd be redundant.
			Println("Opening destination repository")
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

//Connects to the docker daemon
func (d *Hroot) StartDocker(socketURI string) {
	d.dock = crocker.Dial(socketURI)

	// If debug mode is set, print docker version
	if len(os.Getenv("DEBUG")) > 0 {
		d.dock.PrintVersion()
	}
}

//Behavior when docker cache has the image
func (d *Hroot) prepareCacheWithImage(image string) {
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
func (d *Hroot) prepareCacheWithoutImage(image string) {
	switch d.source.scheme {
		case "docker":
			//Can't continue; specified docker as source and it doesn't have it
			ExitGently("Docker does not have", image, "loaded.")
		case "graph":
			d.source.graph.Load(
				image,
				&dex.GraphLoadRequest_Image{
					Dock: d.dock,
					ImageName: image,
				},
			)
	}
}

//Prepare the docker cache
func (d *Hroot) PrepareCache() {
	image := d.launchImage

	//Behavior based on if the docker cache already has an image
	if d.dock.CheckCache(image) {
		d.prepareCacheWithImage(image)
	} else {
		d.prepareCacheWithoutImage(image)
	}

	//Now that the docker cache has the image, run normal behavior
	//Both these actions take place unconditionally, but warn the user if the cache is hot.
	switch d.source.scheme  {
		case "file":
			d.dock.ImportFromFilenameTagstring(d.source.path, image) //Load image from file
		case "index":
			d.dock.Pull(d.image.Index)
	}
}

//Lanuch the container and wait for it to complete
func (d *Hroot) Launch() {
	Println("Launching container.")
	c := d.settings

	//Map the struct values to crocker function params
	d.container = crocker.Launch(d.dock, d.launchImage, c.Command, c.Attach, c.Privileged, c.Folder, c.DNS, c.Mounts, c.Ports, c.Environment)

	//Wait for container
	d.container.Wait()
}

//Prepare the hroot export
func (d *Hroot) ExportBuild(forceEpoch bool) error {
	switch d.dest.scheme {
		case "graph":
			Println("Committing to graph...")

			//Don't give ancestor name to graph publish if source was not the graph.
			ancestor := d.image.Upstream
			if d.source.scheme != "graph" {
				ancestor = ""
			}

			d.dest.graph.Publish(
				d.image.Name,
				ancestor,
				&dex.GraphStoreRequest_Container{
					Container: d.container,
					Settings: guitarconf.Settings{
						Epoch: forceEpoch,
					},
				},
			)
		case "file":
			//Export a tar
			Println("Exporting to", d.dest.path)
			d.container.ExportToFilename(d.dest.path)
	}

	//Commit the image name to the docker cache.
	//	This is so if you run:
	//		hroot build -s index  -d graph --noop
	//		hroot build -s docker -d graph
	//	Docker will already know about your (much cooler) image name :)
	name, tag := crocker.SplitImageName(d.image.Name)
	// Docker really hates its own domain. I know, whatever.
	nameTemp := strings.Replace(name, "docker.io", "docker.IO", -1)
	Println("Exporting to docker cache:", name, tag)
	d.container.Commit(nameTemp, tag)

	return nil
}

//Clean up after ourselves
func (d *Hroot) Cleanup() {
	//Remove the container from cache if desired
	if d.settings.Purge {
		d.container.Purge()
	}

	//Close the docker connection
	d.dock.Close()
}
