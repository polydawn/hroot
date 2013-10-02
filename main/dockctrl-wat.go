package main

import (
	. "polydawn.net/dockctrl/fab"
	"os"
)

func main() {
	cwd, _ := os.Getwd()
	dock := NewDock(cwd+"/dock")
	defer dock.Slay()

	Memo("images")
	dock.Client()("images")()

	Memo("containers")
	dock.Client()("ps", "-a")()
}
