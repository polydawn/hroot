package dex

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

	// we'll make exactly one new dir if the path doesn't exist yet.  more is probably argument error and we abort.
	// this is actually implemented via MkdirAll here (because Mkdir errors on existing, and I can't be arsed) and letting the SaneDir check earlier blow up if we're way out.
	err := os.MkdirAll(g.dir, 0755)
	if err != nil { panic(err); }

	// git init
	g.cmd("init")("--bare")()

	g.withTempTree(func (cmd Command) {
		// set up basic repo to identify as graph repo
		cmd("commit", "--allow-empty", "-mdocket")()
		cmd("checkout", "-b", "docket/init")()

		// discard master branch.  a docket graph has no real use for it.
		cmd("branch", "-D", "master")()
	})

	// should be good to go
	return g
}

func newGraph(dir string) *Graph {
	dir = util.SanePath(dir)

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
	revp := g.cmd(NullIO)("rev-parse", "--is-bare-repository").Output()
	v = (revp == "true\n")
	return
}

/*
	Creates a temporary working tree in a new directory.  Changes the cwd to that location.
	The directory will be empty.  The directory will be removed when your function returns.
*/
func (g *Graph) withTempTree(fn func(cmd Command)) {
	// ensure zone for temp trees is established
	tmpTreeBase := filepath.Join(g.dir, "worktrees")
	err := os.MkdirAll(tmpTreeBase, 0755)
	if err != nil { panic(err); }

	// make temp dir for tree
	tmpdir, err := ioutil.TempDir(tmpTreeBase, "tree.")
	if err != nil { panic(err); }
	defer os.RemoveAll(tmpdir)

	// set cwd
	retreat, err := os.Getwd()
	if err != nil { panic(err); }
	defer os.Chdir(retreat)
	err = os.Chdir(tmpdir)
	if err != nil { panic(err); }

	// construct git command template that knows what's up
	gt := g.cmd(
		Opts{
			Cwd:tmpdir,
		},
		Env{
			"GIT_WORK_TREE": tmpdir,
			"GIT_DIR": g.dir,
		},
	)

	// go time
	fn(gt)
}

func (g *Graph) Publish(lineage string, ancestor string, gr GraphStoreRequest) (hash string) {
	// Handle tags - currently, we discard them when dealing with a graph repo.
	lineage, _  = SplitImageName(lineage)
	ancestor, _ = SplitImageName(ancestor)

	g.withTempTree(func(cmd Command) {
		fmt.Println("starting publish", lineage, "<--", ancestor)

		// check if appropriate branches already exist, and make them if necesary
		if strings.Count(g.cmd("branch", "--list", lineage).Output(), "\n") >= 1 {
			fmt.Println("linage already existed")
			// this is an existing lineage
			//g.cmd("checkout", lineage)() //TODO: verify that we don't need to checkout here, don't think we should because of how we force merge, but add without a head might get startled
		} else {
			// this is a new lineage
			if ancestor == "" {
				fmt.Println("new linage!  making orphan branch for it")
				g.cmd("checkout", "--orphan", lineage)()	//TODO: docket/image/
			} else {
				fmt.Println("new linage!  forking it from ancestor branch.")
				g.cmd("branch", lineage, ancestor)()
			}
		}

		// apply the GraphStoreRequest to unpack the fs (read from fs.tarReader, essentially)
		gr.place(".")	//TODO: verify that a relative path here is safe, or just replace is os.Getwd again.

		// exec git add, tree write, merge, commit.
		g.cmd("add", "--all")()
		g.forceMerge(ancestor, lineage)

		hash = ""	//FIXME
	})
	return
}

func (g *Graph) Load(lineage string, gr GraphLoadRequest) (hash string) {
	lineage, _ = SplitImageName(lineage) //Handle tags

	g.withTempTree(func(cmd Command) {
		// checkout lineage.
		// "-f" because otherwise if git thinks we already had this branch checked out, this working tree is just chock full of deletes.
		g.cmd("checkout", "-f", lineage)()

		// the gr consumes this filesystem and shoves it at whoever it deals with; we're actually hands free after handing over a dir.
		gr.receive(".")	//TODO: verify that a relative path here is safe, or just replace is os.Getwd again.

		hash = ""	//FIXME
	})
	return
}

// having a load-by-hash:
//   - you can't combine it with lineage, because git doesn't really know what branches are, historically speaking.
//       - unless we decide we're committing lineage in some structured form of the commit messages, which is possible, but not sure if want.
//       - we could traverse up from the lineage branch ref and make sure the hash is reachable from it, but more than one ref is going to be able to reach most hashes (i.e. hashes that are pd-base will be reachable from pd-nginx).


func (g *Graph) forceMerge(source string, target string) {
	writeTree := g.cmd("write-tree").Output()
	writeTree = strings.Trim(writeTree, "\n")
	commitMsg := fmt.Sprintf("updated %s<<%s", target, source)
	commitTreeCmd := g.cmd("commit-tree", writeTree, Opts{In: commitMsg})
	if source != "" {
		commitTreeCmd = commitTreeCmd("-p", source, "-p", target)
	}
	mergeTree := strings.Trim(commitTreeCmd.Output(), "\n")
	g.cmd("merge", "-q", mergeTree)()
}

//Checks if the graph has a branch.
func (g *Graph) HasBranch(branch string) bool {
	//Git magic is involved. Response will be of non-zero length if branch exists.
	result := g.cmd("ls-remote", ".", "refs/heads/" + branch).Output()
	return len(result) > 0
}
