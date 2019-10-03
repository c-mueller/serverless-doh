package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/c-mueller/serverless-doh/core"
)

func main() {
	hndlr, _ := core.NewHandler(core.GetConfigFromEnvironment())
	algnhsa.ListenAndServe(hndlr, &algnhsa.Options{
		BinaryContentTypes: []string{
			"application/dns-message",
			"application/dns-udpwireformat",
		},
	})
}
