package main

import (
	. "fmt"
	"os"
	"github.com/jessevdk/go-flags"
	. "polydawn.net/hroot/commands"
	. "polydawn.net/hroot/util"
)

var parser = flags.NewNamedParser("hroot", flags.Default | flags.HelpFlag)

const EXIT_BADARGS = 1
const EXIT_PANIC = 2
const EXIT_BAD_USER = 10

// print only the error message (don't dump stacks).
// unless any debug mode is on; then don't recover, because we want to dump stacks.
func panicHandler() {
	if err := recover(); err != nil {

		//HrootError is used for user-friendly exits. Just print & exit.
		if dockErr, ok := err.(HrootError) ; ok {
			Print(dockErr.Error())
			os.Exit(EXIT_BAD_USER)
		}

		//Check for existence of debug environment variable
		if len(os.Getenv("DEBUG")) == 0 {
			//Debug not set, be friendlier about the problem
			Println(err)
			Println("\n" + "Hroot crashed! This could be a problem with docker or git, or hroot itself." + "\n" + "To see more about what went wrong, turn on stack traces by running:" + "\n\n" + "export DEBUG=1" + "\n\n" + "Feel free to contact the developers for help:" + "\n" + "https://github.com/polydawn/hroot" + "\n")
			os.Exit(EXIT_PANIC)
		} else {
			//Adds main to the top of the stack, but keeps original information.
			//Nothing we can do about it. Golaaaaannngggg....
			panic(err)
		}
	}
}

func main() {
	defer panicHandler()

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

		//Default settings
		&RunCmdOpts{ }, //cannot set a default source; default is determined intelligently at runtime
	)
	parser.AddCommand(
		"build",
		"Transform a container",
		"Transform a container based on configuration in the current directory.",

		//Default settings
		&BuildCmdOpts{
			Destination: "graph",
		},
	)
	parser.AddCommand(
		"version",
		"Print hroot version",
		"Print hroot version",
		&VersionCmdOpts{},
	)

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
