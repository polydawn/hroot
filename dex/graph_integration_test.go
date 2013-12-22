package dex

// Very nearly all testing for dex is integration testing, sadly; this is inevitable since we're relying on exec to use git.

import (
	"path/filepath"
	"io/ioutil"
	"bytes"
	"os"
	"archive/tar"
	"testing"
	"strings"
	"github.com/coocood/assrt"
)

func TestLoadGraphAbsentIsNil(t *testing.T) {
	do(func() {
		assert := assrt.NewAssert(t)

		assert.Nil(LoadGraph("."))

		assert.Nil(LoadGraph("notadir"))
	})
}

func assertLegitGraph(assert *assrt.Assert, g *Graph) {
	assert.NotNil(g)

	gstat, _ := os.Stat(filepath.Join(g.dir))
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
	// file 'a' is just ascii text with normal permissions
	if err := ioutil.WriteFile(
		filepath.Join(pth, "a"),
		[]byte{ 'a', 'b' },
		0644,
	); err != nil { panic(err); }

	// file 'b' is binary with unusual permissions
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
	// (you're just expected to have nuked the working tree before calling this)

	// add an executable file
	//TODO

	// all of this is horseshit, and what you're really going to do is make a tar stream programatically, because that's the input guitar understands.

	// file 'd/d/d' is renamed to 'd/e' and 'd/d' dropped
	//TODO
}

func fsSetA() *tar.Reader {
	var buf bytes.Buffer
	fs := tar.NewWriter(&buf)

	// file 'a' is just ascii text with normal permissions
	fs.WriteHeader(&tar.Header{
		Name:     "a",
		Mode:     0644,
		Size:     2,
		Typeflag: tar.TypeReg,
	})
	fs.Write([]byte{ 'a', 'b' })

	// file 'b' is binary with unusual permissions
	fs.WriteHeader(&tar.Header{
		Name:     "b",
		Mode:     0640,
		Size:     3,
		Typeflag: tar.TypeReg,
	})
	fs.Write([]byte{ 0x1, 0x2, 0x3 })

	fs.Close()
	return tar.NewReader(&buf)
}

func fsSetB() *tar.Reader {
	var buf bytes.Buffer
	fs := tar.NewWriter(&buf)

	// file 'a' is unchanged from SetA
	fs.WriteHeader(&tar.Header{
		Name:     "a",
		Mode:     0644,
		Size:     2,
		Typeflag: tar.TypeReg,
	})
	fs.Write([]byte{ 'a', 'b' })

	// file 'b' is removed

	// file 'e' is executable
	fs.WriteHeader(&tar.Header{
		Name:     "e",
		Mode:     0755,
		Size:     3,
		Typeflag: tar.TypeReg,
	})
	fs.Write([]byte{ 'e', 'x', 'e' })

	// file 'd/d/z' is deeper
	fs.WriteHeader(&tar.Header{
		Name:     "d/d/z",
		Mode:     0644,
		Size:     2,
		Typeflag: tar.TypeReg,
	})
	fs.Write([]byte{ 'z', '\n' })

	fs.Close()
	return tar.NewReader(&buf)
}

func TestNewOrphanLineage(t *testing.T) {
	do(func() {
		assert := assrt.NewAssert(t)

		g := NewGraph(".")
		lineage := "line"
		ancestor := ""

		g.Publish(
			lineage,
			ancestor,
			&GraphStoreRequest_Tar{
				Tarstream: fsSetA(),
			},
		)

		assert.Equal(
			3,
			strings.Count(
				g.cmd("ls-tree", "refs/heads/"+lineage).Output(),
				"\n",
			),
		)
	})
}

func TestLinearExtensionToLineage(t *testing.T) {
	do(func() {
		assert := assrt.NewAssert(t)

		g := NewGraph(".")
		lineage := "line"
		ancestor := "line"

		g.Publish(
			lineage,
			"",
			&GraphStoreRequest_Tar{
				Tarstream: fsSetA(),
			},
		)

		g.Publish(
			lineage,
			ancestor,
			&GraphStoreRequest_Tar{
				Tarstream: fsSetB(),
			},
		)

		assert.Equal(
			4,
			strings.Count(
				g.cmd("ls-tree", "refs/heads/"+lineage).Output(),
				"\n",
			),
		)
	})
}

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
