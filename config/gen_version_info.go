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
	version := "0.0.1"
	revision := "NYI"

	branch := "master"

	buildTimestamp := time.Now().Unix()
	usr, _ := user.Current()
	hname, _ := os.Hostname()
	buildContext := fmt.Sprintf("%s@%s", strings.ToLower(usr.Name), hname)
	goVersion := runtime.Version()

	file := fmt.Sprintf(template, version, revision, branch, buildContext, buildTimestamp, goVersion)
	fmt.Println(file)
	ioutil.WriteFile("generated_version_info.go", []byte(file),0555)
}
