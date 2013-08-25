package prime

import (
	. "polydawn.net/dockctrl/fab"
	. "fmt"
)

/**
 * @param projDir
 *      the directory to expect docker.conf and other materials,
 *      and also the base dir against which relative mounts are resolved.
 * @param dock
 *      a dock config object we'll use to get in contact with a daemon.
 *      the daemon should already be running.
 * @param override
 *      a final override layer to configuration.  nil is disregarded.
 */
func LauncherPrime(projDir string, dock *Dock, override map[string]interface{}) (string, int) {
	//Load config
	confCreate, confStart, _ := loadConfig(projDir, override)

	//Connect
	Memo("Connecting...")
	conn := dial("unix", dock.GetSockPath())
	defer conn.Close()

	//Print version
	data, _ := call(conn, "GET", "/version", nil)
	Memo("Connected to docker "+data["Version"].(string))

	//Create a docker
	data, _ = call(conn, "POST", "/containers/create", confCreate)

	//Print any warnings
	warnings, _ := data["Warnings"].([]interface{})
	for _, v := range warnings { Println(v) }

	//Get the container ID
	id := data["Id"].(string)
	Memo("Created docker "+id)

	//Start the container, warn if something's fishy
	_, code := callEmpty(conn, "POST", "/containers/" + id + "/start", confStart)
	if (code != 204) {
		Println("Warning: return code from start was ", code)
	}
	Memo("Docker is running.")

	//Wait for container to finish
	data, _ = call(conn, "POST", "/containers/" + id + "/wait", nil)
	status := int(data["StatusCode"].(float64))
	Memo("Container exited with status code "+string(status))
	return id, status
}
