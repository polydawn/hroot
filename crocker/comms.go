package crocker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	. "net/http/httputil"
	"strings"
	. "polydawn.net/hroot/util"
)

// Much of this logic is stolen directly from https://github.com/dotcloud/docker/blob/master/commands.go
//        Modified to use a single connection instead of opening many, among other things...

const ApiVersion = "1.4"
const ServerVersion = "0.6.3"

// Hit the docker daemon with an HTTP request, returns response byte array
func call(conn *ClientConn, method, path string, data interface{}) ([]byte, int) {
	//Encode data if needed
	var params io.Reader
	if data != nil {
			buf, err := json.Marshal(data)
			if err != nil { ExitGently("JSON marshalling failed: " + err.Error()) }
			params = bytes.NewBuffer(buf)
	}

	//Create the request
	req, err := http.NewRequest(method, fmt.Sprintf("/v%s%s", ApiVersion, path), params)
	if err != nil {
			ExitGently("Could not create request: " + err.Error())
	}

	//Headers
	req.Header.Set("User-Agent", "Docker-Client/" + ServerVersion)
	if data != nil {
			req.Header.Set("Content-Type", "application/json")
	} else if method == "POST" {
			req.Header.Set("Content-Type", "plain/text")
	}

	resp, err := conn.Do(req)
	if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
					ExitGently("Can't connect to docker daemon. Is 'docker -d' running on this host?")
			}
			ExitGently("Couldn't connect to docker: " + err.Error())
	}

	//Read in response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil { ExitGently("Could not read response: " + err.Error()) }

	//Check error code
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			if len(body) == 0 {
					ExitGently("Bad return: " + http.StatusText(resp.StatusCode))
			}
			ExitGently("Bad return: " + string(body))
	}

	return body, resp.StatusCode
}
