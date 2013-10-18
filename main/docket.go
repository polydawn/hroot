package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var parser = flags.NewNamedParser("docket", flags.Default | flags.HelpFlag)

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
		"Run a container",
		"Run a container based on configuration in the current directory.",
		&runCmdOpts{},
	)
	parser.AddCommand(
		"build",
		"Transform a container",
		"Transform a container based on configuration in the current directory.",
		&buildCmdOpts{},
	)
}
