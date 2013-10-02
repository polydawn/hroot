package main

import (
	. "polydawn.net/dockctrl/fab"
	. "polydawn.net/dockctrl/prime"
	"os"
)

func main() {
	cwd, _ := os.Getwd()
	dock := NewDock(cwd+"/dock")
	defer dock.Slay()

	LauncherPrime(cwd, dock, nil)
}
