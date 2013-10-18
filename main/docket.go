package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var parser = flags.NewNamedParser("docket", flags.Default)

var EXIT_BADARGS = 1
var EXIT_PANIC = 2

func main() {
	_, err := parser.Parse()
	if err != nil {
		os.Exit(EXIT_BADARGS)
	}
	os.Exit(0)
}

func init() {
	// parser.AddCommand(
	// 	"command",
	// 	"description",
	// 	"long description",
	// 	&whateverCmd{}
	// )
	parser.AddCommand(
		"run",
		"run a container",
		"run a container based on configuration in the current directory.",
		&runCmdOpts{},
	)
	parser.AddCommand(
		"build",
		"transform a container",
		"transform a container based on configuration in the current directory.",
		&buildCmdOpts{},
	)
}
