package fab

import (
	. "polydawn.net/gosh/psh"
)

var GraphDir = "/dockctrl.graph/"
var GraphGit = Sh("git")(DefaultIO)(Opts{Cwd: GraphDir})

var GraphGit_Cleanse = func() {
	GraphGit("reset")()
	GraphGit("checkout", ".")()
	GraphGit("clean", "-xf")()
}
