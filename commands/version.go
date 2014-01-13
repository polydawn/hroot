package commands

import (
	. "fmt"
)

const Version = "0.5.2"

type VersionCmdOpts struct { }

//Version command
func (opts *VersionCmdOpts) Execute(args []string) error {
	Println("hroot version", Version)
	return nil
}
