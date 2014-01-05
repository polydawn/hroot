package dex

import (
	"archive/tar"
	"io"
	"polydawn.net/hroot/crocker"
	"polydawn.net/guitar/stream"
	"polydawn.net/guitar/conf"
)

type GraphStoreRequest interface {
	// it's a little hard to decide what goes in here.
	// i don't *want* to have it talk about writing to a folder on disk, because that's a *current* implementation detail, and one that i actively want to go away.
	// but that's what it needs to do at the end of the day
	// unless we leave that detail to Graph and make this focused around tar streams,
	// which isn't really any less wrong
	// so really, if at all possible, we should just make the details of what this interface is not actually visible or implementable by others outside of this package.
	place(path string)

	settings() conf.Settings
}

type GraphStoreRequest_Tar struct {
	Tarstream *tar.Reader
	Settings conf.Settings
}

func (gr *GraphStoreRequest_Tar) place(path string) {
	// Use guitar to write the tar's contents to the graph
	err := stream.ExportToFilesystem(gr.Tarstream, path, gr.Settings)
	if err != nil { panic(err); }
}

func (gr *GraphStoreRequest_Tar) settings() conf.Settings {
	return gr.Settings
}

type GraphStoreRequest_Container struct {
	Container *crocker.Container
	Settings conf.Settings
}

func (gr *GraphStoreRequest_Container) place(path string) {
	// Ask the container to become a tar byte stream
	exportReader, exportWriter := io.Pipe()
	go gr.Container.Export(exportWriter)

	// See it as a tarstream and punt that kind of store request
	wat := GraphStoreRequest_Tar{
		Tarstream: tar.NewReader(exportReader),
		Settings: gr.Settings,
	} // golang, you're bad.  why can't i one-line this.
	wat.place(path)
}

func (gr *GraphStoreRequest_Container) settings() conf.Settings {
	return gr.Settings
}


