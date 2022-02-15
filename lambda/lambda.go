package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/c-mueller/serverless-doh/core/doh"
	"github.com/sirupsen/logrus"
)

func main() {
	hndlr, _ := doh.NewStaticHandler(doh.GetConfigFromEnvironment(), logrus.WithField("platform", "lambda"))
	algnhsa.ListenAndServe(hndlr, &algnhsa.Options{
		BinaryContentTypes: []string{
			"application/dns-message",
			"application/dns-udpwireformat",
		},
	})
}
