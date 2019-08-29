// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"os"
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
	//fmt.Println("Determining Current Revision")
	//revcmd := exec.Command("git", "show", "--format=%h")
	//err := revcmd.Run()
	//if err != nil {
	//	data, _ := revcmd.StderrPipe()
	//	dd, _ := ioutil.ReadAll(data)
	//	fmt.Println(string(dd))
	//	panic(err.Error())
	//}
	//revd, _ := revcmd.Output()
	//revision := string(revd)
	//fmt.Printf("Determined Revision %s\n", revision)
	//
	////git rev-parse --abbrev-ref HEAD
	//bcmd := exec.Command("git", "rev-parse", "--abbrev-ref=HEAD")
	//err = bcmd.Run()
	//if err != nil {
	//	data, _ := bcmd.StderrPipe()
	//	dd, _ := ioutil.ReadAll(data)
	//	fmt.Println(string(dd))
	//	panic(err.Error())
	//}
	//bd, _ := bcmd.Output()
	//branch := string(bd)
	//fmt.Printf("Determined branch %s\n", revision)

	branch := "master"
	revision := "TBD"

	version := fmt.Sprintf("%s.%s", branch, revision)

	buildTimestamp := time.Now().Unix()
	usr, _ := user.Current()
	hname, _ := os.Hostname()
	buildContext := fmt.Sprintf("%s@%s", strings.ToLower(usr.Name), hname)
	goVersion := runtime.Version()

	file := fmt.Sprintf(template, version, revision, branch, buildContext, buildTimestamp, goVersion)
	fmt.Println(file)
	ioutil.WriteFile("generated_version_info.go", []byte(file), 0555)
}
