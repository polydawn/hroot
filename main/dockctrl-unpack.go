package main

import (
	. "polydawn.net/dockctrl/fab"
	. "polydawn.net/gosh/psh"
	"os"
)

func doFabUnpack(dock *Dock, boxname string) {
	Memo("importing '"+boxname+"' images from git")

	GraphGit("checkout", boxname)()
	in, err := os.OpenFile(GraphDir+"image.tar", os.O_RDONLY, 0644)
	if err != nil { panic(err); }
	dock.Client()("import")("-", boxname)(Opts{In: in})()
}

func main() {
	cwd, _ := os.Getwd()
	dock := NewDock(cwd+"/dock")
	defer dock.Slay()

	for _, boxname := range os.Args[1:] {
		doFabUnpack(dock, boxname)
	}

	Memo("successfully unpacked!")
	dock.Client()("images")()
}
