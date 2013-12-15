package dex

import (
	. "polydawn.net/pogo/gosh"
	"io/ioutil"
	"os"
	"testing"
	"strings"
	"github.com/coocood/assrt"
)

func TestGitSeparateWorkTree(t *testing.T) {
	do(func() {
		assert := assrt.NewAssert(t)
		cwd, _ := os.Getwd()

		// note that currently NewGraph doesn't make a bare repo.
		// if this goes well, might change it to do that.
		g := NewGraph("graph.git")

		// override the Cwd that Graph initialized.
		// if this goes well, we'll probably add a Graph.treecmd(tree string) method to do this.
		gt1 := g.cmd(
			Opts{
				Cwd:cwd,
			},
			Env{
				"GIT_WORK_TREE":"tree",
				"GIT_DIR":"graph.git/.git/",
			},
		)

		assert.Equal(
			"",
			gt1("ls-tree")("HEAD").Output(),
		)

		os.MkdirAll("tree", 0755)
		//gt1("checkout")()

		ioutil.WriteFile(
			"tree/a",
			[]byte{ 'a', 'b' },
			0644,
		)
		gt1("add")("--all")()
		gt1("commit")("-m=commit 1")()



		os.MkdirAll("tree2", 0755)
		gt2 := g.cmd(
			Opts{
				Cwd:cwd,
			},
			Env{
				"GIT_WORK_TREE":"tree2",
				"GIT_DIR":"graph.git/.git/",
			},
		)

		assert.Equal(
			1, // because ls-tree asks what's in history, not what's in this working tree
			strings.Count(
				gt2("ls-tree")("HEAD").Output(),
				"\n",
			),
		)

		ioutil.WriteFile(
			"tree2/y",
			[]byte{ 'y' },
			0644,
		)
		ioutil.WriteFile(
			"tree2/z",
			[]byte{ 'z' },
			0644,
		)
		gt1("add")("--all")()
		gt1("commit")("-m=commit 2")()

		assert.Equal(
			2, // file 'a' should have been removed, and 'y' and 'z' present
			strings.Count(
				gt2("ls-tree")("HEAD").Output(),
				"\n",
			),
		)
	})
}
