package dex

import (
	. "polydawn.net/gosh/psh"
	"path/filepath"
)

type Graph struct {
	/*
		Absolute path to the base/working dir for of the graph git repository.
	*/
	dir string

	/*
		Cached command template for exec'ing git with this graph's cwd.
	*/
	cmd Shfn
}

func NewGraph(dir string) *Graph {
	dir, err := filepath.Abs(dir)
	if err != nil { panic(err); }

	return &Graph{
		dir: dir,
		cwd: Sh("git")(DefaultIO)(Opts{Cwd: dir}),
	}
}

/*
Wipes uncommitted changes in the git working tree.
*/
func (g *Graph) Cleanse() {
	g.cmd("reset")()
	g.cmd("checkout", ".")()
	g.cmd("clean", "-xf")()
}
