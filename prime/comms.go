package prime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	. "net/http/httputil"
	"os"
	"strings"
)

// Much of this logic is stolen directly from https://github.com/dotcloud/docker/blob/master/commands.go
//	Modified to use a single connection instead of opening many, among other things...

const ApiVersion = "1.4"
const ServerVersion = "0.5.2-dev"

// Engage chevrons
func dial(proto, addr string) *ClientConn {
	dial, err := net.Dial(proto, addr)

	if err != nil { // Error handling stolen directly from docker
		if strings.Contains(err.Error(), "connection refused") {
			fmt.Println("Can't connect to docker daemon. Is 'docker -d' running on this host?")
		} else {
			fmt.Println("Connection Error:", err.Error())
		}
		os.Exit(1)
	}

	return NewClientConn(dial, nil)
}

// Decode JSON cand cast to string -> object map
func decodeJSON(data []byte) (map[string]interface {}) {
	var raw interface{}
	err := json.Unmarshal(data, &raw)
	if (err != nil) {
		fmt.Println("Cannot decode JSON:", err.Error())
		os.Exit(1)
	}
	return raw.(map[string]interface{})
}

// Hit the docker daemon with an HTTP request, returns JSON-decoded string -> object array
func call(conn *ClientConn, method, path string, data interface{}) (map[string]interface {}, int) {
	//Fire the request, die on error for now
	response, code, err := callRaw(conn, method, path, data)
	if (err != nil) {
		fmt.Println(code, "Request Error:", err.Error())
		os.Exit(1)
	}
	return decodeJSON(response), code
}

// Convenience call for requests that aren't expected to return any data.
func callEmpty(conn *ClientConn, method, path string, data interface{}) ([]byte, int) {
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

	//Fire the request
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

//Stolen from golang docs, modified slightly BECAUSE THE JSON DOCUMENTATION LIES ABOUT INT VERSUS FLOATS
func debugJSON(data map[string]interface {}) {
	for k, v := range data {
		switch vv := v.(type) {
		case nil:
			fmt.Println(k, "(nil)", vv)
		case bool:
			fmt.Println(k, "(bool)", vv)
		case string:
			fmt.Println(k, "(string)", vv)
		case int:
			fmt.Println(k, "(int)", vv)
		case float64:
			fmt.Println(k, "(float)", vv)
		case []interface{}:
			fmt.Println(k, "(array):")
			for i, u := range vv {
				fmt.Println(i, u)
			}
		case map[string]interface {}:
			fmt.Println(k, "(object):")
			debugJSON(vv)
			fmt.Println("-")
		default:
			fmt.Println(k, "is of a type I don't know how to handle", vv)
		}
	}
}
