package dex;

import (
	"archive/tar"
	"io"
	"polydawn.net/hroot/crocker"
	"polydawn.net/guitar/stream"
	"sync"
)

type GraphLoadRequest interface {
	receive(path string)
}

type GraphLoadRequest_Tar struct {
	Tarstream *tar.Writer
}

func (gr *GraphLoadRequest_Tar) receive(path string) {
	// Use guitar to read the graph contents a tarstream
	err := stream.ImportFromFilesystem(gr.Tarstream, path)
	if err != nil { panic(err); }
}

type GraphLoadRequest_Image struct {
	Dock *crocker.Dock
	ImageName string
}

func (gr *GraphLoadRequest_Image) receive(path string) {
	//Pipe for I/O, and a waitgroup to make async action blocking
	importReader, importWriter := io.Pipe()
	var wait sync.WaitGroup
	wait.Add(1)

	//Closure to run the docker import
	go func() {
		gr.Dock.Import(importReader, gr.ImageName, "latest")
		wait.Done()
	}()

	//Run the guitar import
	wat := GraphLoadRequest_Tar{
		Tarstream: tar.NewWriter(importWriter),
	} // golang, you're bad.  why can't i one-line this.
	wat.receive(path);
	importWriter.Close()

	// wait for docker importing on the tar byte stream to return
	wait.Wait()
}


