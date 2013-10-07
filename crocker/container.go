package crocker

import (
	"io"
	"os"
	. "polydawn.net/gosh/psh"
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
	Streams out a tar as produced by `docker export`.
*/
func (c *Container) Export(writer io.Writer) {
	c.dock.cmd()("export", c.id)(Opts{Out: writer})()
	writer.(io.WriteCloser).Close() //... this might be fixed by updating gosh
}

/*
	Convenience wrapper for Export(io.Writer) but writing to a file.
*/
func (c *Container) ExportToFilename(path string) {
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	c.Export(out)
}
