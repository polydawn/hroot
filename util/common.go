package util

import (
	. "fmt"
	"os"
	"path/filepath"
	"strings"
)

//Return the absolute path and evaluate for symlinks.
//Where we should call this (rather than just .Abs) is debatable.
func SanePath(loc string) (string, error) {
	//Get absolute representation and clean
	loc, error := filepath.Abs(loc)
	if error != nil { return loc, error }

	//Check that the directory exists, remove symlinks from path
	dir, base := filepath.Dir(loc), filepath.Base(loc)
	dir, error = filepath.EvalSymlinks(dir)
	if error != nil { return dir, error}

	return filepath.Join(dir, base), nil
}

//If the user specified a target, use that, else use the command's default target
func GetTarget(args []string, defaultTarget string) string {
	if len(args) >= 1 {
		return args[0]
	} else {
		return defaultTarget
	}
}

//Given a URI, return the scheme name separate from everything else
//	See: https://en.wikipedia.org/wiki/URI_scheme#Generic_syntax
func ParseURI(input string) (string, string) {
	//Parse input
	components := strings.SplitN(input, ":", 2)
	scheme := components[0]

	//Check if a path was provided
	path := ""
	if len(components) > 1 {
		path = components[1]
	}

	//Check that the scheme name is one we support
	switch scheme {
		case "docker", "index": //pass
		case "graph", "file": //sanitize paths
			path, err := SanePath(path)
			if err != nil {
				Println("Could not sanitize path: " + path)
				os.Exit(1)
			}
		case "":
			Println("Command source/destination is empty; must be one of (graph, file, docker, index)")
			os.Exit(1)
		default:
			Println("Unrecognized scheme '" + scheme + "': must be one of (graph, file, docker, index)")
			os.Exit(1)
	}

	return scheme, path
}
