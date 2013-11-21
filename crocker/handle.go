//Handles the weird bits of the docker CLI

package crocker

import (
	"io/ioutil"
	"strings"
	"time"
	. "polydawn.net/docket/util"
)

//Where to place & call CIDfiles
const TempDir = "/tmp"
const TempPrefix = "trion-"

//The default docker tag
const DefaultTag = "latest"

//Create a temporary file for docker to print its CID to
func CreateCIDfile() string {
	//Create a temporary file
	CIDfileFD, err := ioutil.TempFile(TempDir, TempPrefix)
	if err != nil { ExitGently("Error: could not create cidfile in", TempDir) }

	//Stat the file to get the name. Yes, this is dumb.
	info, err := CIDfileFD.Stat()
	if err != nil { ExitGently("Error: could not stat cidfile.") }

	//Release the file descriptor. This *has to happen* so docker can write to it. Yes, it is a little sad.
	CIDfileFD.Close()
	return TempDir + "/" + TempPrefix + info.Name()
}

//Given a filename that docker prints a ContainerID to, poll for file and write that CID to a channel.
//	This defeats the rather insane problem caused by docker not really writing the CIDfile at any particular time...
func PollCid(filename string) chan string {
	containerID := ""
	getCID := make(chan string)

	go func() {
		for i := 0; i <= 20; i++ {
			if out, err := ioutil.ReadFile(filename); err == nil {
				containerID = strings.Trim(string(out), "\n")
				getCID <- containerID
				return
			}
			time.Sleep(100 * time.Millisecond)
		}

		ExitGently("Error: could not read cidfile", filename)
	}()

	return getCID
}

//Given an image string, returns the image name and tag.
//	'ubuntu:12.10' -> 'ubuntu', '12.10'
//	'ubuntu' -> 'ubuntu', 'latest'
func SplitImageName(image string) (string, string) {
	sp := strings.Split(image, ":")

	if len(sp) == 2 {
		return sp[0], sp[1]
	} else {
		return image, DefaultTag
	}
}
