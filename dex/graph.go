package dex

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	. "polydawn.net/pogo/gosh"
	. "polydawn.net/hroot/crocker"
	"polydawn.net/hroot/util"
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

// strap this in only sometimes -- some git commands need this prefix to be explicit about branches instead of tags; others refuse it because they're already forcibly about branches.
const git_branch_ref_prefix = "refs/heads/"
const hroot_ref_prefix = "hroot/"
const hroot_image_ref_prefix = hroot_ref_prefix+"image/"

/*
	Loads a Graph if there is a git repo initialized at the given dir; returns nil if a graph repo not found.
	The dir must be the root of the working tree of the git dir.

	A graph git repo is distingushed by containing branches that start with "hroot/" -- this is how hroot outputs branches that contain its data.
*/
func LoadGraph(dir string) *Graph {
	// optimistically, set up the struct we're checking out
	g := newGraph(dir)

	// ask git what it thinks of all this.
	if g.isHrootGraphRepo() {
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
	g := newGraph(dir)
	if g.isHrootGraphRepo() {
		// if we can just be a load, do it
		return g
	} else if g.isRepoRoot() {
		// if this is a repo root, but didn't look like a real graph...
		util.ExitGently("Attempted to make a hroot graph at ", g.dir, ", but there is already a git repo there and it does not appear to belong to hroot.")
	} // else carry on, make it!

	// we'll make exactly one new dir if the path doesn't exist yet.  more is probably argument error and we abort.
	// this is actually implemented via MkdirAll here (because Mkdir errors on existing, and I can't be arsed) and letting the SaneDir check earlier blow up if we're way out.
	err := os.MkdirAll(g.dir, 0755)
	if err != nil { panic(err); }

	// git init
	g.cmd("init")("--bare")()

	g.withTempTree(func (cmd Command) {
		// set up basic repo to identify as graph repo
		cmd("commit", "--allow-empty", "-mhroot")()
		cmd("checkout", "-b", hroot_ref_prefix+"init")()

		// discard master branch.  a hroot graph has no real use for it.
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

//Is git ready and configured to make commits?
func (g *Graph) IsConfigReady() bool {
	//Get the current git configuration
	config := g.cmd(NullIO)("config", "--list").Output()

	//Check that a user name and email is defined
	return strings.Contains(config, "user.name=") && strings.Contains(config, "user.email=")
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
		fmt.Println("Starting publish of ", lineage, " <-- ", ancestor)

		// check if appropriate branches already exist, and make them if necesary
		if strings.Count(g.cmd("branch", "--list", hroot_image_ref_prefix+lineage).Output(), "\n") >= 1 {
			fmt.Println("Lineage already existed.")
			// this is an existing lineage
			g.cmd("symbolic-ref", "HEAD", git_branch_ref_prefix+hroot_image_ref_prefix+lineage)()
		} else {
			// this is a new lineage
			if ancestor == "" {
				fmt.Println("New lineage!  Making orphan branch for it.")
				g.cmd("checkout", "--orphan", hroot_image_ref_prefix+lineage)()
			} else {
				fmt.Println("New lineage!  Forking it from ancestor branch.")
				g.cmd("branch", hroot_image_ref_prefix+lineage, hroot_image_ref_prefix+ancestor)()
				g.cmd("symbolic-ref", "HEAD", git_branch_ref_prefix+hroot_image_ref_prefix+lineage)()
			}
		}
		g.cmd("reset")

		// apply the GraphStoreRequest to unpack the fs (read from fs.tarReader, essentially)
		gr.place(".")

		// exec git add, tree write, merge, commit.
		g.cmd("add", "--all")()
		g.forceMerge(ancestor, lineage)

		hash = ""	//FIXME
	})
	return
}

func (g *Graph) Load(lineage string, gr GraphLoadRequest) (hash string) {
	lineage, _ = SplitImageName(lineage) //Handle tags

	//Check if the image is in the graph so we can generate a relatively friendly error message
	if !g.HasBranch(hroot_image_ref_prefix+lineage) {	//HALP
		util.ExitGently("Image branch name", lineage, "not found in graph.")
	}

	g.withTempTree(func(cmd Command) {
		// checkout lineage.
		// "-f" because otherwise if git thinks we already had this branch checked out, this working tree is just chock full of deletes.
		g.cmd("checkout", "-f", git_branch_ref_prefix+hroot_image_ref_prefix+lineage)()

		// the gr consumes this filesystem and shoves it at whoever it deals with; we're actually hands free after handing over a dir.
		gr.receive(".")

		hash = ""	//FIXME
	})
	return
}

// having a load-by-hash:
//   - you can't combine it with lineage, because git doesn't really know what branches are, historically speaking.
//       - we could traverse up from the lineage branch ref and make sure the hash is reachable from it, but more than one ref is going to be able to reach most hashes (i.e. hashes that are pd-base will be reachable from pd-nginx).
//       - unless we decide we're committing lineage in some structured form of the commit messages.
//            - which... yes, yes we are.  the first word of any commit is going to be the image lineage name.
//            - after the space can be anything, but we're going to default to a short description of where it came from.
//            - additional lines can be anything you want.
//            - if we need more attributes in the future, we'll start doing them with the git psuedo-standard of trailing "Signed-Off-By: %{name}\n" key-value pairs.
//            - we won't validate any of this if you're not using load-by-hash.


func (g *Graph) forceMerge(source string, target string) {
	writeTree := g.cmd("write-tree").Output()
	writeTree = strings.Trim(writeTree, "\n")
	commitMsg := ""
	if source == "" {
		commitMsg = fmt.Sprintf("%s imported from an external source", target)
	} else {
		commitMsg = fmt.Sprintf("%s updated from %s", target, source)
	}
	commitTreeCmd := g.cmd("commit-tree", writeTree, Opts{In: commitMsg})
	if source != "" {
		commitTreeCmd = commitTreeCmd(
			"-p", git_branch_ref_prefix+hroot_image_ref_prefix+source,
			"-p", git_branch_ref_prefix+hroot_image_ref_prefix+target,
		)
	}
	mergeTree := strings.Trim(commitTreeCmd.Output(), "\n")
	g.cmd("merge", "-q", mergeTree)()
}

//Checks if the graph has a branch.
func (g *Graph) HasBranch(branch string) bool {
	//Git magic is involved. Response will be of non-zero length if branch exists.
	result := g.cmd("ls-remote", ".", git_branch_ref_prefix + branch).Output()
	return len(result) > 0
}

/*
	Check if a git repo exists and if it has the branches that declare it a hroot graph.
*/
func (g *Graph) isHrootGraphRepo() bool {
	if !g.isRepoRoot() { return false; }
	if !g.HasBranch("hroot/init") { return false; }
	// We could say a hroot graph shouldn't have a master branch, but we won't.
	// We don't create one by default, but you're perfectly welcome to do so and put a readme for your coworkers in it or whatever.
	return true
}
