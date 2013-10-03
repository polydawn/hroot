package crocker

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
	Launches a new Container in the given Dock.
*/
func LaunchContainer(dock *Dock, command string) *Container {
	dockRun := dock.cmd()("run")

	// set up cidfile
	CIDfilename := CreateCIDfile()
	dockRun = dockRun("-cidfile", CIDfilename)

	//TODO: moar conf

	// launch
	dockRun()
	getCID := PollCid(CIDfilename)

	return &Container{
		dock: dock,
		id: <-getCID,
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