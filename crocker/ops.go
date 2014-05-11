package crocker

//All the high-level functionality

import (
	. "fmt"
	"encoding/json"
	"io"
	"os"
	"strings"
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

	// Docker really hates its own domain. I know, whatever.
	nameTemp := strings.Replace(name, "docker.io", "docker.IO", -1)

	dock.Cmd()("import", "-", nameTemp + ":" + tag)(Opts{In: reader})()
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
	for _, img := range images {
		//Docker image listings are now grouped by tag, iterate over those
		for _, curTag := range img.RepoTags {
			name2, tag2 := SplitImageName(curTag)

			// Docker really hates its own domain. I know, whatever.
			nameTemp := strings.Replace(name2, "docker.io", "docker.IO", -1)

			if nameTemp == name && tag2 == tag { return true }
		}
	}

	return false
}

// Run the simple docker version command for debugging
func (dock *Dock) PrintVersion() {
	dock.Cmd()("version")()
}
