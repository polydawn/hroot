package crocker

// Functions that have to do with talking to a docker instance via API calls

import (
	"encoding/json"
	"net"
	"net/http/httputil"
	"strings"
	. "polydawn.net/hroot/util"
)

// Engage chevrons
func (dock *Dock) Dial() {
	//Connect (if we're not already connected)
	if dock.sock != nil { return }
	dial, err := net.Dial("unix", dock.GetSockPath())

	// Error handling stolen directly from docker
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			ExitGently("Can't connect to docker daemon. Is 'docker -d' running on this host?")
		} else {
			ExitGently("Connection Error:", err.Error())
		}
	}

	dock.sock = httputil.NewClientConn(dial, nil)
}

// Check if an image is loaded in docker's cache.
func (dock *Dock) CheckCache(image string) bool {
	dock.Dial()
	var images []APIImages
	name, tag := SplitImageName(image)

	//API call
	data, _ := call(dock.sock, "GET", "/images/json", nil)
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
