package dex

import (
	"io/ioutil"
	"os"
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
