package dex

import (
	"fmt"
	"io"
	"os"
	. "polydawn.net/pogo/gosh"
	. "polydawn.net/docket/crocker"
	"polydawn.net/docket/util"
	"strings"
)

type Graph struct {
	/*
		Absolute path to the base/working dir for of the graph git repository.
	*/
	dir string

	/*
		Cached command template for exec'ing git with this graph's cwd.
	*/
	cmd Command
}

/*
	Loads a Graph if there is a git repo initialized at the given dir; returns nil if a graph repo not found.
	The dir must be the root of the working tree of the git dir.

	A graph git repo is distingushed by containing branches that start with "docket/" -- this is how docket outputs branches that contain its data.
*/
func LoadGraph(dir string) *Graph {
	// optimistically, set up the struct we're checking out
	g := newGraph(dir)

	// ask git what it thinks of all this.
	if g.isRepoRoot() {
		return g
	} else {
		return nil
	}
}

/*
	Attempts to load a Graph at the given dir, or creates a new one if no graph repo is found.
	If a new graph is fabricated, it will be initialized by:
	 - creating a new git repo,
	 - making a blank root commit,
	 - and tagging it with a branch name that declares it to be a graph repo.

	Note if your cwd is already in a git repo, the new graph will not be commited, nor will it be made a submodule.
	You're free to make it a submodule yourself, but git quite wants you to have a remote url before it accepts your submodule.
*/
func NewGraph(dir string) *Graph {
	// try for a load, and if that flies, return it.
	g := LoadGraph(dir)
	if g != nil {
		return g
	}
	g = newGraph(dir)

	// if the path to this doesn't exist yet, it will now.
	err := os.MkdirAll(g.dir, 0755)
	if err != nil { panic(err); }

	// git init
	g.cmd("init")()

	// set up basic repo to identify as graph repo
	g.cmd("commit", "--allow-empty", "-mdocket")()
	g.cmd("checkout", "-b", "docket/init")()

	// should be good to go
	return g
}

func newGraph(dir string) *Graph {
	dir, err := util.SanePath(dir)
	if err != nil { panic(err); }

	// optimistically, set up the struct.
	// we still need to either verify or initalize git here.
	return &Graph{
		dir: dir,
		cmd: Sh("git")(DefaultIO)(Opts{Cwd: dir}),
	}
}

func (g *Graph) isRepoRoot() (v bool) {
	defer func() {
		// if the path doesn't even exist, launching the command will panic, and that's fine.
		// if the path isn't within a git repo at all, it will exit with 128, gosh will panic, and that's fine.
		if recover() != nil {
			v = false
		}
	}()
	tld := g.cmd(NullIO)("rev-parse", "--show-toplevel").Output()
	v = (tld == g.dir+"\n")
	return
}

/*

*/
func (g *Graph) GetDir() string {
	return g.dir
}

/*
Wipes uncommitted changes in the git working tree.
*/
func (g *Graph) Cleanse() {
	g.cmd("reset")()
	g.cmd("reset", "--hard")()
	g.cmd("clean", "-xf")()
}

// Prepares for a publish by creating a branch
func (g *Graph) PreparePublish(lineage string, ancestor string) {
	//Handle tags - currently, we discard them when dealing with a graph repo.
	lineage, _  = SplitImageName(lineage)
	ancestor, _ = SplitImageName(ancestor)

	if strings.Count(g.cmd("branch", "--list", lineage).Output(), "\n") < 1 {
		//Memo("this is a new lineage!")
		g.cmd("checkout", "-b", lineage)()
		g.cmd("rm", "*")
	} else {
		g.cmd("checkout", lineage)()
	}
}

/*
Commits a new image.  The "lineage" branch name will be extended by this new commit (or
created, if it doesn't exist), and the "ancestor" branch will also be credited as a parent
of the new commit.
*/
func (g *Graph) Publish(lineage string, ancestor string) {
	//Handle tags - currently, we discard them when dealing with a graph repo.
	lineage, _  = SplitImageName(lineage)
	ancestor, _ = SplitImageName(ancestor)

	g.cmd("add", "--all")()
	g.forceMerge(ancestor, lineage)
	// g.cmd("show")(Opts{OkExit:[]int{0, 141}})()
}

func (g *Graph) forceMerge(source string, target string) {
	writeTree := g.cmd("write-tree").Output()
	writeTree = strings.Trim(writeTree, "\n")
	commitMsg := fmt.Sprintf("updated %s<<%s", target, source)
	mergeTree := g.cmd("commit-tree", writeTree, "-p", source, "-p", target, Opts{In: commitMsg}).Output()
	mergeTree = strings.Trim(mergeTree, "\n")
	g.cmd("merge", mergeTree)()
}

/*
Returns a read stream for the requested image.  Internally, the commit that the "lineage" branch ref
currently points to is opened and "image.tar" is read from.
*/
func (g *Graph) Load(lineage string) io.Reader {
	//FIXME: entirely possible to do this without doing a `git checkout`... do so
	lineage, _ = SplitImageName(lineage) //Handle tags
	g.cmd("checkout", lineage)()

	in, err := os.OpenFile(g.dir+"/image.tar", os.O_RDONLY, 0644)
	if err != nil { panic(err); }
	return in
}

//Checks if the graph has a branch.
func (g *Graph) HasBranch(branch string) bool {
	//Git magic is involved. Response will be of non-zero length if branch exists.
	result := g.cmd("ls-remote", ".", "refs/heads/" + branch).Output()
	return len(result) > 0
}
