package crocker

import (
	. "fmt"
	"os"
	. "polydawn.net/gosh/psh"
	"strings"
)

/*
	Refers to a Container process.
	The container has been started and has an ID.
	Methods on this struct can be used to inspect and manipulate the container.
*/
type Container struct {
	// reference to the daemon we launched this container with
	dock *Dock

	// id of this container
	id string
}

/*
	Default tar filename amd image tag
*/
const TarFile = "image.tar"
const DefaultTag = "latest"

/*
	Launches a new Container in the given Dock.
	Punting on documentation while things are in flux; see command.go struct for details.
*/
func Launch(dock *Dock, image string, command []string, attach bool, privileged bool, startIn string, dns []string, mounts [][]string, ports [][]string, environment [][]string) *Container {
	dockRun := dock.cmd()("run")

	//Where should docker write the new CID?
	CIDfilename := CreateCIDfile()
	dockRun = dockRun("-cidfile", CIDfilename)

	//Where should the container start?
	dockRun = dockRun("-w", startIn)

	//Is the docker in privleged (pwn ur box) mode?
	if privileged {
		dockRun = dockRun("-privileged")
	}

	//Custom DNS servers?
	for i := range dns {
		dockRun = dockRun("-dns", dns[i])
	}

	//What folders get mounted?
	for i := range mounts {
		dockRun = dockRun("-v", mounts[i][0] + ":" + mounts[i][1] + ":" + mounts[i][2])
	}

	for i := range ports {
		dockRun = dockRun("-p", ports[i][0] + ":" + ports[i][1])
	}

	//What environment variables are set?
	for i:= range environment {
		dockRun = dockRun("-e", environment[i][0] + "=" + environment[i][1])
	}

	//Are we attaching?
	if attach {
		dockRun = dockRun("-i", "-t")
	}

	//Add image name
	dockRun = dockRun(image)

	//What command should it run?
	for i := range command {
		dockRun = dockRun(command[i])
	}

	//Poll for the CID and run the docker
	dockRun()
	getCID := PollCid(CIDfilename)

	return &Container{
		dock: dock,
		id:   <-getCID,
	}
}

/*
	Waits for the container's main process to exit (i.e., wraps `docker wait`).
*/
func (c *Container) Wait() {
	//TOOD:FUTURE: consider if wait/attach should be just built-in when launching a container, and this method just wraps checks against a promise-pattern.
	c.dock.cmd()("wait", c.id)()
}

/*
	Discards the container state and filesystem (i.e., wraps `docker rm`).

	After this method completes, many other inspection methods on this Container become
	invalid and will error, because the backing data has been, well, purged.

	This will error if called on a still-running container.
*/
func (c *Container) Purge() {
	c.dock.cmd()("rm", c.id)()
}

/*
	Executes 'docker export', after ensuring there is no image.tar in the way.
	This is because docker will *happily* export into an existing tar.
*/
func (c *Container) Export(path string) {
	tar := path + TarFile

	//Check for existing file
	file, err := os.Open(tar)
	if err == nil {
		_, err = file.Stat()
		file.Close()
	}

	//Delete tar if it exists
	if err == nil {
		Println("Warning: Output image.tar already exists. Overwriting...")
		err = os.Remove("./image.tar")
		if err != nil {
			Println("Fatal: Could not delete tar file.")
			os.Exit(1)
		}
	}

	out, err := os.OpenFile(tar, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	c.dock.cmd()("export", c.id)(Opts{Out: out})()
}

/*
	Import an image into docker's repository.
*/
func (c *Container) Import(path, name, tag string) {
	tar := path + TarFile

	//Open the file
	in, err := os.Open(tar)
	if err != nil {
		Println("Fatal: Could not open file for import:", tar)
	}

	Println("Importing", name + ":" + tag)
	c.dock.cmd()("import", "-", name, tag)(Opts{In: in, Out: os.Stdout })()
}

/*
	Import an image from a docker-style image string, such as 'ubuntu:latest'
*/
func (c *Container) ImportFromString(path, image string) {
	//Get the repository and tag
	name, tag := "", ""
	sp := strings.Split(image, ":")

	//If both a name and version are specified, use them, otherwise just tag it as 'latest'
	if len(sp) == 2 {
		name = sp[0]
		tag = sp[1]
	} else {
		name = image
		tag = DefaultTag
	}

	c.Import(path, name, tag)
}
