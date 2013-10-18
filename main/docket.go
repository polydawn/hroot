package main

import (
	"github.com/jessevdk/go-flags"
	. "polydawn.net/docket/commands"
	"os"
)

var parser = flags.NewNamedParser("docket", flags.Default | flags.HelpFlag)

var EXIT_BADARGS = 1
var EXIT_PANIC = 2

func main() {
	//Go-flags is a little too clever with sub-commands.
	//To keep the help-command parity with git & docker / etc, check for 'help' manually before args parse
	if len(os.Args) < 2 || os.Args[1] == "help" {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	//Parse for command & flags, and exit with a relevant return code.
	_, err := parser.Parse()
	if err != nil {
		os.Exit(EXIT_BADARGS)
	} else {
		os.Exit(0)
	}
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
		&RunCmdOpts{},
	)
	parser.AddCommand(
		"build",
		"Transform a container",
		"Transform a container based on configuration in the current directory.",
		&BuildCmdOpts{},
	)
}
