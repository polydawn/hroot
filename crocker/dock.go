package crocker

import (
	"bytes"
	"encoding/json"
	. "fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
	. "polydawn.net/pogo/gosh"
	. "polydawn.net/hroot/util"
)

const defaultSock = "unix:///var/run/docker.sock" //Where a docker daemon will run by default
const waitDuraction = 500 * time.Millisecond      //How long to wait for docker to create a socket
const pollDuraction = 10 * time.Millisecond       //How long to wait between polling for socket
const ApiVersion = "1.10"                          //Docker api version
const ServerVersion = "0.10.0"                     //Docker header version so its log can complain :)

//A struct representing a connection to a docker daemon
type Dock struct {
	//A socket connection to the docker for API calls
	sock *httputil.ClientConn

	//The docket URI (used in -H flag passthrough for exec wrapping)
	sockURI string
}

// Engage chevrons
func Dial(uri string) *Dock {
	//If no socket path was specified, use the default
	if uri == "" {
		uri = defaultSock
	}
	Println("Connecting to", uri)

	//Golang wants a network type and path.
	//Docker's -H flag wants a full URI.
	//Dial takes a URI and temporarily converts for Golang's sake.
	sp := strings.Split(uri, "://")
	if len(sp) != 2 { ExitGently("Socket path must be a full URI, example: unix:///var/run/docker.sock") }
	sockType := sp[0]
	sockPath := sp[1]

	//Set flags for expiry
	timeout := time.Now().Add(waitDuraction)
	done := false

	//Loop until success or timeout
	for !done {
		//Update timeout flag
		done = time.Now().After(timeout)

		//If the URI is a unix socket, do some extra sanity checks
		if sockType == "unix" {
			//Attempt to stat the socket
			sockStat, err := os.Stat(sockPath)
			if os.IsNotExist(err) {
				// The socket does not exist yet, continue waiting
				continue
			} else if err != nil {
				//Some other stat error, should not happen
				panic(err)
			} else if (sockStat.Mode() & os.ModeSocket) == 0 {
				//That's no sock!
				ExitGently("The path", uri, "is not a socket!")
			}
		}

		// Try to dial out
		dial, err := net.Dial(sockType, sockPath)

		//If the socket is live, we're finished
		if err == nil {
			return &Dock {
				sock: httputil.NewClientConn(dial, nil), //open http connection
				sockURI: uri,                            //socket URI
			}
		} else if strings.Contains(err.Error(), "permission denied"){
			ExitGently("You don't have permission to write to the docker socket. Try running as root.")
		}

		//Wait for a bit before checking again
		if !done {
			time.Sleep(pollDuraction)
		}
	}

	ExitGently("Can't connect to docker daemon. Is 'docker -d' running on this host?")
	return nil
}

//Close the socket if it's open
func (dock *Dock) Close() {
	if dock.sock != nil {
		dock.sock.Close()
	}
}


//Returns a gosh command struct for use in exec-wrapping docker
func (dock *Dock) Cmd() Command {
	template := Sh("docker")(DefaultIO)("-H=" + dock.sockURI)

	// If debug mode is set, print every command before executing
	if len(os.Getenv("DEBUG")) > 0 {
		template = template.Debug(func (ct *CommandTemplate) {
			Println("Exec:", ct.Cmd, ct.Args)
		})
	}

	return template
}

// Hit the docker daemon with an HTTP request, returns response byte array
func (dock *Dock) Call(method, path string, data interface{}) ([]byte, int) {
	// Print network traffic to terminal if DEBUG env var exists
	networkDebug := (len(os.Getenv("DEBUG")) > 0)
	if (networkDebug) { Println("Calling: " + method + " " + path) }

	//Encode data if needed
	var params io.Reader
	if data != nil {
			buf, err := json.Marshal(data)
			if err != nil { ExitGently("JSON marshalling failed: " + err.Error()) }

			if (networkDebug) { Println("Sending: " + string(buf)) }
			params = bytes.NewBuffer(buf)
	}

	//Create the request
	req, err := http.NewRequest(method, Sprintf("/v%s%s", ApiVersion, path), params)
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

	resp, err := dock.sock.Do(req)
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

	if (networkDebug) { Println("Network: " + string(body)) }

	return body, resp.StatusCode
}
