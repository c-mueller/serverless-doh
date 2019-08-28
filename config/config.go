package config

import (
	"fmt"
	"time"
)

var (
	ProgramName    = "sls-doh"
	Version        = "UNSET"
	Revision       = "UNSET"
	Branch         = "UNSET"
	BuildTimestamp = time.Now().Unix()
	BuildContext   = "UNSET"
	GoVersion      = "UNSET"

	Blacklist             map[string]bool
	Whitelist             map[string]bool
	ListCreationTimestamp int
	BlacklistItemCount    int
	WhitelistItemCount    int
)

func GetUserAgent() string {
	return fmt.Sprintf("%s %s", ProgramName, Version)
}
