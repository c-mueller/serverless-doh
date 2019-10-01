package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/c-mueller/serverless-doh/config"
	"github.com/c-mueller/serverless-doh/core"
	"os"
	"strconv"
)

func MustGetEnvBool(name string) bool {
	val := os.Getenv(name)
	r, _ := strconv.ParseBool(val)
	return r
}

func main() {
	hndlr, _ := core.NewHandler(&core.Config{
		EnableBlocking:    MustGetEnvBool("ENABLE_BLOCKING"),
		TCPOnly:           true,
		Upstream:          []string{"1.1.1.1:53", "8.8.8.8:53"},
		Verbose:           MustGetEnvBool("VERBOSE"),
		UserAgent:         config.GetUserAgent(),
		LogGuessedIP:      MustGetEnvBool("VERBOSE"),
		AppendListHeaders: MustGetEnvBool("APPEND_LIST_INFO"),
		Timeout:           60,
		Tries:             10,
	})
	algnhsa.ListenAndServe(hndlr, &algnhsa.Options{
		BinaryContentTypes: []string{
			"application/dns-message",
			"application/dns-udpwireformat",
		},
	})
}
