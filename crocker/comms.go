package crocker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	. "net/http/httputil"
	"os"
	"strings"
)

// Much of this logic is stolen directly from https://github.com/dotcloud/docker/blob/master/commands.go
//        Modified to use a single connection instead of opening many, among other things...

const ApiVersion = "1.4"
const ServerVersion = "0.6.3"

// Convenience call for requests that aren't expected to return any data.
func call(conn *ClientConn, method, path string, data interface{}) ([]byte, int) {
	//Fire the request, die on error for now
	response, code, err := callRaw(conn, method, path, data)
	if (err != nil) {
		fmt.Println(code, "Request Error:", err.Error())
		os.Exit(1)
	}
	return response, code
}

// Hit the docker daemon with an HTTP request, returns response byte array
func callRaw(conn *ClientConn, method, path string, data interface{}) ([]byte, int, error) {
	//Encode data if needed
	var params io.Reader
	if data != nil {
			buf, err := json.Marshal(data)
			if err != nil {
					return nil, -1, err
			}
			params = bytes.NewBuffer(buf)
	}

	//Create the request
	req, err := http.NewRequest(method, fmt.Sprintf("/v%s%s", ApiVersion, path), params)
	if err != nil {
			return nil, -2, err
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
					return nil, -3, fmt.Errorf("Can't connect to docker daemon. Is 'docker -d' running on this host?")
			}
			return nil, -4, err
	}

	//Read in response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
			return nil, -5, err
	}

	//Check error code
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			if len(body) == 0 {
					return nil, resp.StatusCode, fmt.Errorf("Error: %s", http.StatusText(resp.StatusCode))
			}
			return nil, resp.StatusCode, fmt.Errorf("Error: %s", body)
	}

	return body, resp.StatusCode, nil
}
