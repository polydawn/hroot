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
}


func init() {
	// parser.AddCommand(
	// 	"command",
	// 	"description",
	// 	"long description",
	// 	&whateverCmd{}
	// )
	parser.AddCommand(
		"launch",
		"launch a container",
		"launch a container based on configuration in the current directory",
		&launchCmd{},
	)
}

type launchCmd struct {}
func (opts *launchCmd) Execute(args []string) error {
	return nil
}
