package dex

// Very nearly all testing for dex is integration testing, sadly; this is inevitable since we're relying on exec to use git.

import (
	"path/filepath"
	"io/ioutil"
	"os"
	"testing"
	"strings"
	"github.com/coocood/assrt"
)

func do(fn func()) {
	retreat, err := os.Getwd()
	if err != nil { panic(err); }

	defer os.Chdir(retreat)

	basedir := os.Getenv("BASEDIR")
	if len(basedir) != 0 {
		err = os.Chdir(basedir)
		if err != nil { panic(err); }
	}

	err = os.MkdirAll("target/test", 0755)
	if err != nil { panic(err); }
	tmpdir, err := ioutil.TempDir("target/test","")
	if err != nil { panic(err); }
	err = os.Chdir(tmpdir)
	if err != nil { panic(err); }

	fn()
}

func TestLoadGraphAbsentIsNil(t *testing.T) {
	do(func() {
		assert := assrt.NewAssert(t)

		assert.Nil(LoadGraph("."))

		assert.Nil(LoadGraph("notadir"))
	})
}

func assertLegitGraph(assert *assrt.Assert, g *Graph) {
	assert.NotNil(g)

	gstat, _ := os.Stat(filepath.Join(g.dir,".git"))
	assert.True(gstat.IsDir())

	assert.True(g.HasBranch("docket/init"))

	assert.Equal(
		"",
		g.cmd("ls-tree")("HEAD").Output(),
	)
}

func TestNewGraphInit(t *testing.T) {
	do(func() {
		assertLegitGraph(
			assrt.NewAssert(t),
			NewGraph("."),
		)
	})
}

func TestLoadGraphEmpty(t *testing.T) {
	do(func() {
		assert := assrt.NewAssert(t)

		NewGraph(".")

		assertLegitGraph(assert, LoadGraph("."))
	})
}

func TestNewGraphInitNewDir(t *testing.T) {
	do(func() {
		assertLegitGraph(
			assrt.NewAssert(t),
			NewGraph("deep"),
		)
	})
}

func TestNewGraphInitRejectedOnDeeper(t *testing.T) {
	do(func() {
		defer func() {
			err := recover()
			if err == nil { t.Fail(); }
		}()
		NewGraph("deep/deeper")
	})
}

func fwriteSetA(pth string) {
	// file 'a' is just ascii text
	if err := ioutil.WriteFile(
		filepath.Join(pth, "a"),
		[]byte{ 'a', 'b' },
		0644,
	); err != nil { panic(err); }

	// file 'b' is a secret
	if err := ioutil.WriteFile(
		filepath.Join(pth, "b"),
		[]byte{ 0x1, 0x2, 0x3 },
		0640,
	); err != nil { panic(err); }

	// file 'd/d/d' is so dddeep
	//TODO
}

func fwriteSetB(pth string) {
	// file 'a' is unchanged
	if err := ioutil.WriteFile(
		filepath.Join(pth, "a"),
		[]byte{ 'a', 'b' },
		0644,
	); err != nil { panic(err); }

	// file 'b' is removed
	//TODO

	// add an executable file
	//TODO

	// file 'd/d/d' is renamed to 'd/e' and 'd/d' dropped
	//TODO
}

func TestNewOrphanLineage(t *testing.T) {
	do(func() {
		assert := assrt.NewAssert(t)

		g := NewGraph(".")
		lineage := "line"
		ancestor := ""

		g.PreparePublish(lineage, ancestor)

		fwriteSetA(g.GetDir())

		g.Publish(lineage, ancestor)

		assert.Equal(
			2,
			strings.Count(
				g.cmd("ls-tree", "refs/heads/"+lineage).Output(),
				"\n",
			),
		)
	})
}

// func TestCleanBeforeNewLineage(t *testing.T) {

// func TestLinearExtensionToLineage(t *testing.T) {
// 	do(func() {
// 		assert := assrt.NewAssert(t)

// 		//TODO
// 	})
// }

// func TestNewDerivedLineage(t *testing.T) {
// 	do(func() {
// 		assert := assrt.NewAssert(t)

// 		//TODO
// 	})
// }

// func TestDerivativeExtensionToLineage(t *testing.T) {
// 	do(func() {
// 		assert := assrt.NewAssert(t)

// 		//TODO
// 	})
// }
