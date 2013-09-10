package main

import (
	. "polydawn.net/dockctrl/trion"
	. "fmt"
	"os"
)

func main() {
	config := FindConfig(".")
	CID := Run(config)
	Println(CID)
	os.Exit(0) //GOTTA GO FAST. SIX TIMES AS MUCH FAST.
}
