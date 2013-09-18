package crocker

import (
	"fmt"
	. "polydawn.net/gosh/psh"
	"time"
)

/**
 * @param dir relative path to dock dir.
 */
func NewDock(dir string) *Dock {
	dock := &Dock{dir: dir}
	//TODO ping if exists, spawn process if not, etc.
	Sh("mkdir")("-p")(DefaultIO)(dock.Dir())()
	dock.daemon().Start()
	dock.isMine = true
	time.Sleep(200 * time.Millisecond)
	return dock
}

type Dock struct {
	/**
	 * Absolute path to the base dir for a docker daemon.
	 *
	 * 'docker.sock' and 'docker.pid' are expected to exist immediately inside this path.
	 * The daemon's working dir may also be here.
	 *
	 * The last segment of the path is quite probably a symlink, and should be respected
	 * even if dangling (unless that means making more than one directory on the far
	 * side; if things are that dangling, give up).
	 */
	dir string

	/**
	 * True iff the daemon at this dock location was spawned by us.
	 * Basically used to determine if Slay() should actually fire teh lazors or not.
	 */
	isMine bool
}

func (dock Dock) Dir() string {
	return dock.dir
}

func (dock *Dock) cmd() Shfn {
	return Sh("docker")(DefaultIO)(
		"-H="+fmt.Sprintf("unix://%s", dock.GetSockPath()),
	)
}

func (dock *Dock) Client() Shfn {
	return dock.cmd()
}

func (dock *Dock) GetPidfilePath() string {
	return fmt.Sprintf("/%s/%s", dock.Dir(), "docker.pid")
}

func (dock *Dock) GetSockPath() string {
	return fmt.Sprintf("/%s/%s", dock.Dir(), "docker.sock")
}

func (dock *Dock) daemon() Shfn {
	return dock.cmd()(
		"-d",
		"-g="+dock.Dir(),
		"-p="+dock.GetPidfilePath(),
	)(Opts{Cwd: dock.Dir()})
}

func (dock *Dock) Slay() {
	if !dock.isMine { return; }
	Sh("bash")("-c")(DefaultIO)("kill `cat \""+dock.GetPidfilePath()+"\"`")()
}
