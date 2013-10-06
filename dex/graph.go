package dex

import (
	. "polydawn.net/gosh/psh"
	"io"
	"os"
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
		cmd: Sh("git")(DefaultIO)(Opts{Cwd: dir}),
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

/*
Commits a new image.  The "lineage" branch name will be extended by this new commit (or
created, if it doesn't exist), and the "ancestor" branch will also be credited as a parent
of the new commit.
*/
func (g *Graph) Publish(imageStream io.Writer, lineage string, ancestor string) {
	//TODO
}

/*
Returns a read stream for the requested image.  Internally, the commit that the "lineage" branch ref
currently points to is opened and "image.tar" is read from.
*/
func (g *Graph) Load(lineage string) io.Reader {
	//FIXME: entirely possible to do this without doing a `git checkout`... do so
	g.cmd("checkout", lineage)()
	in, err := os.OpenFile(g.dir+"/image.tar", os.O_RDONLY, 0644)
	if err != nil { panic(err); }
	return in
}
