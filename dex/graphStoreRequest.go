package dex

import (
	"archive/tar"
	"io"
	"polydawn.net/hroot/crocker"
	"polydawn.net/guitar/stream"
)

type GraphStoreRequest interface {
	// it's a little hard to decide what goes in here.
	// i don't *want* to have it talk about writing to a folder on disk, because that's a *current* implementation detail, and one that i actively want to go away.
	// but that's what it needs to do at the end of the day
	// unless we leave that detail to Graph and make this focused around tar streams,
	// which isn't really any less wrong
	// so really, if at all possible, we should just make the details of what this interface is not actually visible or implementable by others outside of this package.
	place(path string)
}

type GraphStoreRequest_Tar struct {
	Tarstream *tar.Reader
}

func (gr *GraphStoreRequest_Tar) place(path string) {
	// Use guitar to write the tar's contents to the graph
	err := stream.ExportToFilesystem(gr.Tarstream, path)
	if err != nil { panic(err); }
}

type GraphStoreRequest_Container struct {
	Container *crocker.Container
}

func (gr *GraphStoreRequest_Container) place(path string) {
	// Ask the container to become a tar byte stream
	exportReader, exportWriter := io.Pipe()
	go gr.Container.Export(exportWriter)

	// See it as a tarstream and punt that kind of store request
	wat := GraphStoreRequest_Tar{
		Tarstream: tar.NewReader(exportReader),
	} // golang, you're bad.  why can't i one-line this.
	wat.place(path)
}


