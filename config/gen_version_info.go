// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"time"
)

const template = `
package config

func init() {
	Version = "%s"
	Revision = "%s"
	Branch = "%s"
	BuildContext = "%s"
	BuildTimestamp = %d
	GoVersion = "%s"
}`

func main() {
	latestTagCmd := exec.Command("git describe --abbrev=0 --tags")
	err := latestTagCmd.Run()
	version := "dev"
	if err == nil {
		data, _ := latestTagCmd.Output()
		version = string(data)
	}

	//git rev-parse --abbrev-ref HEAD

	revisionCmd := exec.Command("git log -n1 --pretty=%H")
	err = revisionCmd.Run()
	d, _ := revisionCmd.Output()
	revision := string(d)

	branchCmd := exec.Command("git rev-parse --abbrev-ref HEAD")
	err = branchCmd.Run()
	d, _ = branchCmd.Output()
	branch := string(d)

	buildTimestamp := time.Now().Unix()
	usr, _ := user.Current()
	hname, _ := os.Hostname()
	buildContext := fmt.Sprintf("%s@%s", usr.Name, hname)
	goVersion := runtime.Version()

	file := fmt.Sprintf(template, version, revision, branch, buildContext, buildTimestamp, goVersion)
	fmt.Println(file)
}
