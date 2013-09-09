package main

import (
	. "polydawn.net/dockctrl/trion"
)

func main() {
	config := FindConfig(".")
	PrepRun(config)()
}
