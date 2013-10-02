package main

import (
	. "polydawn.net/dockctrl/fab"
	  "polydawn.net/dockctrl/prime"
	. "polydawn.net/gosh/psh"
	. "fmt"
	"os"
	"strings"
)

func doFab(dock *Dock, source string, target string, servicePath string) {
	GraphGit_Cleanse()

	Memo("importing source box '"+source+"'")
	GraphGit("checkout", source)()
	in, err := os.OpenFile(GraphDir+"image.tar", os.O_RDONLY, 0644)
	if err != nil { panic(err); }
	dock.Client()("import")("-", source)(Opts{In: in})()
	dock.Client()("images")()

	Memo("running transition "+source+":"+target)
	confCreateOverride := map[string]interface{}{
		"Image": source,
	}
	cid, exitCode := prime.LauncherPrime(servicePath, dock, confCreateOverride)
	if exitCode != 0 {
		panic(Errorf("container died in transition (%d)", exitCode))
	}

	Memo("exporting box '"+target+"'")
	if strings.Count(GraphGit("branch", "--list", target).Output(), "\n") < 1 {
		Memo("this is a new lineage!")
		GraphGit("checkout", "-b", target)()
		GraphGit("rm", "*")
	} else {
		GraphGit("checkout", target)()
	}
	out, err := os.OpenFile(GraphDir+"image.tar", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil { panic(err); }
	dock.Client()("export")(cid)(Opts{Out: out})()

	Memo("committing '"+target+"'")
	GraphGit("add", "image.tar")()
	forceMerge(source, target)
	GraphGit("show")()
}

func forceMerge(source string, target string) {
	writeTree := GraphGit("write-tree").Output()
	writeTree = strings.Trim(writeTree, "\n")
	commitMsg := Sprintf("updated %s<<%s", target, source)
	mergeTree := GraphGit("commit-tree", writeTree, "-p", source, "-p", target, Opts{In: commitMsg}).Output()
	mergeTree = strings.Trim(mergeTree, "\n")
	GraphGit("merge", mergeTree)()
}

func main() {
	cwd, _ := os.Getwd()
	if len(os.Args) != 2 {
		Fprintf(os.Stderr, "failed: RTFM: expect one argument.\n")
		os.Exit(3)
	}
	servicePath := cwd+"/fab/boxen/"+os.Args[1]
	boxenStat, err := os.Stat(servicePath)
	if err != nil || !boxenStat.IsDir() {
		Fprintf(os.Stderr, "failed: RTFM: expect argument to be a dir in fab/boxen .\n")
		os.Exit(3)
	}
	watDo := strings.Split(os.Args[1], "<<")
	if len(watDo) != 2 {
		Fprintf(os.Stderr, "failed: RTFM: argument to be of form 'A<<B'.\n")
		os.Exit(3)
	}

	dock := NewDock(cwd+"/dock")
	defer dock.Slay()

	doFab(dock, watDo[1], watDo[0], servicePath)
}
