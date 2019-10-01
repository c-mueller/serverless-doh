// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
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
	fmt.Println("Determining Current Revision")
	revcmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	revd, err := revcmd.Output()
	if err != nil {
		panic(err.Error())
	}
	revision := strings.Replace(string(revd), "\n", "", -1)
	fmt.Printf("Determined Revision %s\n", revision)

	//git rev-parse --abbrev-ref HEAD
	bcmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	bd, err := bcmd.Output()
	if err != nil {
		panic(err.Error())
	}
	branch := strings.Replace(string(bd), "\n", "", -1)
	fmt.Printf("Determined branch %s\n", revision)

	//branch := "master"
	//revision := "TBD"

	version := fmt.Sprintf("%s.%s", branch, revision)

	buildTimestamp := time.Now().Unix()
	usr, _ := user.Current()
	hname, _ := os.Hostname()
	buildContext := fmt.Sprintf("%s@%s", strings.ToLower(usr.Name), hname)
	goVersion := runtime.Version()

	file := fmt.Sprintf(template, version, revision, branch, buildContext, buildTimestamp, goVersion)
	fmt.Println(file)
	ioutil.WriteFile("generated_version_info.go", []byte(file), 0644)
}
