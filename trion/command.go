package trion

import (
	. "polydawn.net/gosh/psh"
	"strings"
)

//Default docker command template
var docker = Sh("docker").BakeOpts(DefaultIO)

//Prepare flags for 'docker run'
func PrepRun(config TrionConfig) Shfn {
	dockRun := docker("run")

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

	return dockRun
}
