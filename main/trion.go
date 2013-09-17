package main

import (
	. "polydawn.net/dockctrl/trion"
	. "fmt"
	"os"
)

func main() {
	config := FindConfig(".")

	if len(os.Args) > 1 && os.Args[1] == "build" {
		Println("Building")
		config.Command = config.BuildCommand
	}

	CID := Run(config)
	Println(CID)
	os.Exit(0) //GOTTA GO FAST. SIX TIMES AS MUCH FAST.
}
