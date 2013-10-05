package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var parser = flags.NewNamedParser("dockctrl", flags.Default)

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
		"export",
		"export a container",
		"export a container based on configuration in the current directory.",
		&exportCmdOpts{},
	)
	parser.AddCommand(
		"publish",
		"build and publish a versioned-controlled image",
		"build a container, and place the exported tar into a git repository.",
		&struct{}{},
	)
	parser.AddCommand(
		"unpack",
		"unpack a base image from versioned-controlled storage",
		"unpack a base image from versioned-controlled storage so that it's ready to be used to run a container.",
		&struct{}{},
	)
}
