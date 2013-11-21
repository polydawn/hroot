package dex

// Very nearly all testing for dex is integration testing, sadly; this is inevitable since we're relying on exec to use git.

import (
	"path/filepath"
	"io/ioutil"
	"os"
	"testing"
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
