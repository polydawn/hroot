package trion

import (
	. "polydawn.net/gosh/psh"
	"strings"
	. "fmt"
	"os"
	"path/filepath"
)

//Default docker command template
var docker = Sh("docker").BakeOpts(DefaultIO)

//Where to place & call CIDfiles
const TempDir    = "/tmp"
const TempPrefix = "trion-"

//Prepare flags for 'docker run'
//	Returns a channel you can read to get the CIDfile. Sorry, this is needed due to docker being docker.
func PrepRun(config TrionConfig) (Shfn, chan string) {
	dockRun := docker("run")

	//Find the absolute path for each host mount
	for i, j := range config.Mount {
		cwd, err := filepath.Abs(j[0])
		if err != nil {
			Println("Fatal: Cannot determine absolute path:", j[0])
			os.Exit(1)
		}

		config.Mount[i][0] = cwd
	}

	//Where should docker write the new CID?
	CIDfilename := createCIDfile()
	dockRun = dockRun("-cidfile", CIDfilename)

	//Where should the container start?
	dockRun = dockRun("-w", config.StartIn)

	//Is the docker in privleged (pwn ur box) mode?
	if (config.Privileged) {
		dockRun = dockRun("-privileged")
	}

	//Custom DNS servers?
	if len(config.DNS) != 0 {
		dockRun = dockRun("-dns", strings.Join(config.DNS, ","))
	}

	//What folders get mounted?
	for i := range config.Mount {
		dockRun = dockRun("-v", config.Mount[i][0] + ":" + config.Mount[i][1] + ":" + config.Mount[i][2])
	}

	//Are we attaching?
	if config.Attach {
		dockRun = dockRun("-i", "-t")
	}

	//Add image name
	dockRun = dockRun(config.Image)

	//What command should it run?
	for i := range config.Command {
		dockRun = dockRun(config.Command[i])
	}

	return dockRun, pollCid(CIDfilename)
}

//Executes 'docker run' syncronously, and returns the container's CID.
func Run(config TrionConfig) string {
	run, getCID := PrepRun(config)
	run()
	return <- getCID
}

//Prepares 'docker rm'
func PrepPurge(CID string) Shfn {
	return docker("rm", CID)
}

//Executes 'docker rm' syncronously
func Purge(CID string) {
	PrepPurge(CID)()
}
