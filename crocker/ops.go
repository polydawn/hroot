package crocker

//All the high-level functionality

import (
	. "fmt"
	"encoding/json"
	"io"
	"os"
	. "polydawn.net/pogo/gosh"
	. "polydawn.net/hroot/util"
)

func (dock *Dock) Pull(image string) {
	dock.Cmd()("pull", image)()
}

/*
	Import an image into repository, caching the expanded form so that it's
	ready to be used as a base filesystem for containers.
*/
func (dock *Dock) Import(reader io.Reader, name string, tag string) {
	Println("Importing", name + ":" + tag)
	dock.Cmd()("import", "-", name, tag)(Opts{In: reader})()
}

func (dock *Dock) ImportFromFilename(path string, name string, tag string) {
	in, err := os.Open(path)
	if err != nil { panic(err) }
	dock.Import(in, name, tag)
}

/*
	Import an image from a docker-style image string, such as 'ubuntu:latest'
*/
func (dock *Dock) ImportFromFilenameTagstring(path, image string) {
	name, tag := SplitImageName(image)
	dock.ImportFromFilename(path, name, tag)
}

// Check if an image is loaded in docker's cache.
func (dock *Dock) CheckCache(image string) bool {
	var images []APIImages
	name, tag := SplitImageName(image)

	//API call
	data, _ := dock.Call("GET", "/images/json", nil)
	err := json.Unmarshal(data, &images)
	if err != nil { ExitGently("Docker API error:", err.Error()) }

	//Check if docker has image & tag
	for _, cur := range images {
		if cur.Repository == name && cur.Tag == tag {
			return true
		}
	}

	return false
}
