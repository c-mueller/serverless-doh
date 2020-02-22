package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/c-mueller/serverless-doh/core"
	"github.com/sirupsen/logrus"
)

func main() {
	hndlr, _ := core.NewHandler(core.GetConfigFromEnvironment(), logrus.WithField("platform", "lambda"))
	algnhsa.ListenAndServe(hndlr, &algnhsa.Options{
		BinaryContentTypes: []string{
			"application/dns-message",
			"application/dns-udpwireformat",
		},
	})
}
