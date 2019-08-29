package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/c-mueller/serverless-doh/config"
	"github.com/c-mueller/serverless-doh/core"
)

func main() {
	hndlr, _ := core.NewHandler(&core.Config{
		EnableBlocking: true,
		TCPOnly:        true,
		Upstream:       []string{"1.1.1.1:53", "8.8.8.8:53"},
		Verbose:        true,
		UserAgent:      config.GetUserAgent(),
		LogGuessedIP:   false,
		Timeout:        60,
		Tries:          10,
	})
	algnhsa.ListenAndServe(hndlr, &algnhsa.Options{
		BinaryContentTypes: []string{
			"application/dns-message",
			"application/dns-udpwireformat",
		},
	})
}
