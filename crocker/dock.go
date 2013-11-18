package crocker

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http/httputil"
	"os"
	"path/filepath"
	. "polydawn.net/pogo/gosh"
	"strconv"
	"time"
)

const defaultDir = "/var/run" //Where a docker daemon will run by default
const localDir = "./dock"     //Where to start a local docker if desired

type Dock struct {
	/*
		Absolute path to the base dir for a docker daemon.

		'docker.sock' and 'docker.pid' are expected to exist immediately inside this path.
		The daemon's working dir may also be here.

		The last segment of the path is quite probably a symlink, and should be respected
		even if dangling (unless that means making more than one directory on the far
		side; if things are that dangling, give up).
	*/
	dir string

	/*
		True iff the daemon at this dock location was spawned by us.
		Basically used to determine if Slay() should actually fire teh lazors or not.
	*/
	isMine bool

	/*
		A socket connection to the docker for API calls
	*/
	sock *httputil.ClientConn
}

/*
	Produces a Dock struct referring to an active docker daemon.
	If an existing daemon can be found running, it is used; if not, one is started.
	@param dir path to dock dir.  May be relative.
*/
func NewDock(dir string) *Dock {
	dock := loadDock(dir)
	if dock == nil {
		dock = createDock(dir)
	}
	return dock
}

/*
	Launch a new docker daemon.
	You should try loadDock before this.  (Yes, there are inherently race conditions here.)
*/
func createDock(dir string) *Dock {
	dir, err := filepath.Abs(dir)
	if err != nil { panic(err); }

	dock := &Dock{
		dir:    dir,
		isMine: true,
	}
	Sh("mkdir")("-p")(DefaultIO)(dock.Dir())()
	dock.daemon().Start()
	dock.awaitSocket(250 * time.Millisecond)
	return dock
}

/*
	Check for what looks like an existing docker daemon setup, and return a Dock if one is found.
	We do a basic check if the pidfile and socket are present, and check if pid is stale, and that's it.
	No dialing or protocol negotiation is performed at this stage.
*/
func loadDock(dir string) *Dock {
	dir, err := filepath.Abs(dir)
	if err != nil { panic(err); }

	dock := &Dock{
		dir: dir,
		isMine: false,
	}

	// check pidfile presence.
	pidfileStat, err := os.Stat(dock.GetPidfilePath())
	if os.IsNotExist(err) { return nil; }
	if err != nil { panic(err); }
	if !pidfileStat.Mode().IsRegular() { return nil; }

	// check for process.
	pidfileBlob, err := ioutil.ReadFile(dock.GetPidfilePath())
	if os.IsNotExist(err) { return nil; }
	if err != nil { panic(err); }
	pid, err := strconv.Atoi(string(pidfileBlob))
	if err != nil { panic(err); }
	_, err = os.FindProcess(pid)
	if err != nil { panic(err); }

	// check for socket.
	if dock.awaitSocket(20 * time.Millisecond) != nil { return nil; }

	// alright, looks like a docker daemon.
	return dock
}

/*
	Check/wait for existence of docker.sock.
*/
func (dock *Dock) awaitSocket(patience time.Duration) error {
	timeout := time.Now().Add(patience)
	done := false
	for !done {
		done = time.Now().After(timeout)
		sockStat, err := os.Stat(dock.GetSockPath())
		if os.IsNotExist(err) {
			// continue
		} else if err != nil {
			panic(err)
		} else if (sockStat.Mode() & os.ModeSocket) != 0 {
			// still have to check if it's dialable; docker daemon doesn't even try to remove socket files when it's done.
			dial, err := net.Dial("unix", dock.GetSockPath())
			if err == nil {
				// success!
				dial.Close()
				return nil
			}
		} else {
			// file exists but isn't socket; do not want
			return fmt.Errorf("not a socket in place of docker socket")
		}
		if !done {
			time.Sleep(10 * time.Millisecond)
		}
	}
	return fmt.Errorf("timeout waiting for docker socket")
}

func (dock *Dock) Dir() string {
	return dock.dir
}

func (dock *Dock) IsChildProcess() bool {
	return dock.isMine
}

func (dock *Dock) cmd() Command {
	return Sh("docker")(DefaultIO)(
		"-H=" + fmt.Sprintf("unix://%s", dock.GetSockPath()),
	)
}

func (dock *Dock) Client() Command {
	return dock.cmd()
}

func (dock *Dock) GetPidfilePath() string {
	return filepath.Join(dock.Dir(), "docker.pid")
}

func (dock *Dock) GetSockPath() string {
	return filepath.Join(dock.Dir(), "docker.sock")
}

func (dock *Dock) daemon() Command {
	return dock.cmd()(
		"-d",
		"-g="+dock.Dir(),
		"-p="+dock.GetPidfilePath(),
	)(Opts{Cwd: dock.Dir()})
}

func (dock *Dock) Pull(image string) {
	dock.cmd()("pull", image)()
}

func (dock *Dock) Slay() {
	//Close the socket if it's open
	if dock.sock != nil {
		dock.sock.Close()
	}

	//Kill the daemon if docket started it
	if dock.isMine {
		Sh("bash")("-c")(DefaultIO)("kill `cat \""+dock.GetPidfilePath()+"\"`")()
	}
}

/*
	Import an image into repository, caching the expanded form so that it's
	ready to be used as a base filesystem for containers.
*/
func (dock *Dock) Import(reader io.Reader, name string, tag string) {
	fmt.Println("Importing", name + ":" + tag)
	dock.cmd()("import", "-", name, tag)(Opts{In: reader})()
}

func (dock *Dock) ImportFromFilename(path string, name string, tag string) {
	in, err := os.Open(path)
	if err != nil { panic(err) }
	dock.Import(in, name, tag)
}

/*
	Import an image from a docker-style image string, such as 'ubuntu:latest'
*/
func (dock *Dock) ImportFromFilenameTagstring(path, image string) {
	name, tag := SplitImageName(image)
	dock.ImportFromFilename(path, name, tag)
}
