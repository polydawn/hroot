package dex

// Very nearly all testing for dex is integration testing, sadly; this is inevitable since we're relying on exec to use git.

import (
	"path/filepath"
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

	assert.True(g.HasBranch("hroot/init"))

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

func fsSetA2() *tar.Reader {
	var buf bytes.Buffer
	fs := tar.NewWriter(&buf)

	// file 'a' diffs from SetA
	fs.WriteHeader(&tar.Header{
		Name:     "a",
		Mode:     0644,
		Size:     3,
		Typeflag: tar.TypeReg,
	})
	fs.Write([]byte{ 'a', '\n', 'b' })

	// file 'b' is unchanged from SetA
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

func fsSetC() *tar.Reader {
	var buf bytes.Buffer
	fs := tar.NewWriter(&buf)

	// file 'a' diffs from SetA (and from SetA2, differently)
	fs.WriteHeader(&tar.Header{
		Name:     "a",
		Mode:     0644,
		Size:     3,
		Typeflag: tar.TypeReg,
	})
	fs.Write([]byte{ 'a', '\n', 'c' })

	// file 'b' is still gone

	// file 'e' is deleted from SetB

	// file 'd/d/z' is renamed to 'd/z'
	fs.WriteHeader(&tar.Header{
		Name:     "d/z",
		Mode:     0644,
		Size:     2,
		Typeflag: tar.TypeReg,
	})
	fs.Write([]byte{ 'z', '\n' })

	// and and 'd/d' is *still around* as a (empty!) dir
	fs.WriteHeader(&tar.Header{
		Name:     "d/d",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	})
	fs.Write([]byte{ 'z', '\n' })	//FIXME: heh

	fs.Close()
	return tar.NewReader(&buf)
}

func TestPublishNewOrphanLineage(t *testing.T) {
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
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage).Output(),
				"\n",
			),
		)

		assert.Equal(
			`{"Name":"a","Type":"F","Mode":644,"ModTime":"1970-01-01T00:00:00Z"}` + "\n" +
			`{"Name":"b","Type":"F","Mode":640,"ModTime":"1970-01-01T00:00:00Z"}` + "\n",
			g.cmd("show", git_branch_ref_prefix+docket_image_ref_prefix+lineage+":"+".guitar").Output(),
		)
	})
}

func TestPublishLinearExtensionToLineage(t *testing.T) {
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
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage).Output(),
				"\n",
			),
		)

		assert.Equal(
			1,	// shows a tree
			strings.Count(
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage, "d/d").Output(),
				"\n",
			),
		)

		assert.Equal(
			1,	// shows the file
			strings.Count(
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage, "d/d/z").Output(),
				"\n",
			),
		)

		assert.Equal(
			`{"Name":"a","Type":"F","Mode":644,"ModTime":"1970-01-01T00:00:00Z"}` + "\n" +
			`{"Name":"d/d/z","Type":"F","Mode":644,"ModTime":"1970-01-01T00:00:00Z"}` + "\n" +
			`{"Name":"e","Type":"F","Mode":755,"ModTime":"1970-01-01T00:00:00Z"}` + "\n",
			g.cmd("show", git_branch_ref_prefix+docket_image_ref_prefix+lineage+":"+".guitar").Output(),
		)
	})
}

func TestPublishNewDerivedLineage(t *testing.T) {
	do(func() {
		assert := assrt.NewAssert(t)

		g := NewGraph(".")
		lineage := "ferk"
		ancestor := "line"

		g.Publish(
			ancestor,
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

		println(g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage).Output())
		assert.Equal(
			4,
			strings.Count(
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage).Output(),
				"\n",
			),
		)

		assert.Equal(
			1,	// shows a tree
			strings.Count(
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage, "d/d").Output(),
				"\n",
			),
		)

		assert.Equal(
			1,	// shows the file
			strings.Count(
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage, "d/d/z").Output(),
				"\n",
			),
		)

		assert.Equal(
			`{"Name":"a","Type":"F","Mode":644,"ModTime":"1970-01-01T00:00:00Z"}` + "\n" +
			`{"Name":"d/d/z","Type":"F","Mode":644,"ModTime":"1970-01-01T00:00:00Z"}` + "\n" +
			`{"Name":"e","Type":"F","Mode":755,"ModTime":"1970-01-01T00:00:00Z"}` + "\n",
			g.cmd("show", git_branch_ref_prefix+docket_image_ref_prefix+lineage+":"+".guitar").Output(),
		)
	})
}

func TestPublishDerivativeExtensionToLineage(t *testing.T) {
	do(func() {
		assert := assrt.NewAssert(t)

		g := NewGraph(".")
		lineage := "ferk"
		ancestor := "line"

		// original ancestor import
		g.Publish(
			ancestor,
			"",
			&GraphStoreRequest_Tar{
				Tarstream: fsSetA(),
			},
		)

		// derive 1
		g.Publish(
			lineage,
			ancestor,
			&GraphStoreRequest_Tar{
				Tarstream: fsSetB(),
			},
		)

		// advance the ancestor
		g.Publish(
			ancestor,
			ancestor,
			&GraphStoreRequest_Tar{
				Tarstream: fsSetA2(),
			},
		)

		// advance the derived from the updated ancestor
		g.Publish(
			lineage,
			ancestor,
			&GraphStoreRequest_Tar{
				Tarstream: fsSetC(),
			},
		)

		assert.Equal(
			3,
			strings.Count(
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage).Output(),
				"\n",
			),
		)

		assert.Equal(
			1,	// has the file.  git itself still doesn't see dirs; just guitar does that.
			strings.Count(
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage, "d/").Output(),
				"\n",
			),
		)

		assert.Equal(
			0,	// nothing here.  git itself still doesn't see dirs; just guitar does that.
			strings.Count(
				g.cmd("ls-tree", git_branch_ref_prefix+docket_image_ref_prefix+lineage, "d/d/").Output(),
				"\n",
			),
		)

		assert.Equal(
			`{"Name":"a","Type":"F","Mode":644,"ModTime":"1970-01-01T00:00:00Z"}` + "\n" +
			`{"Name":"d/d","Type":"D","Mode":755,"ModTime":"1970-01-01T00:00:00Z"}` + "\n" +
			`{"Name":"d/z","Type":"F","Mode":644,"ModTime":"1970-01-01T00:00:00Z"}` + "\n",
			g.cmd("show", git_branch_ref_prefix+docket_image_ref_prefix+lineage+":"+".guitar").Output(),
		)
	})
}
