package config

import (
	"fmt"
	"time"
)

//go:generate go run gen_blocklists.go
//go:generate go run gen_version_info.go

var (
	ProgramName    = "sls-doh"
	Version        = "UNSET"
	Revision       = "UNSET"
	Branch         = "UNSET"
	BuildTimestamp = time.Now().Unix()
	BuildContext   = "UNSET"
	GoVersion      = "UNSET"

	Blocklists map[string]bool
)

func GetUserAgent() string {
	return fmt.Sprintf("%s %s.%s", ProgramName, Version, Revision)
}
