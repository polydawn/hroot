package main

import (
	. "polydawn.net/dockctrl/trion"
	. "fmt"
)

func main() {
	config := FindConfig(".")
	CID := Run(config)
	Println(CID)
}
