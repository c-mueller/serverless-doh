package main

import (
	"net/http"

	"github.com/c-mueller/serverless-doh/config"
	"github.com/c-mueller/serverless-doh/core"
)

func HandleDNS(w http.ResponseWriter, r *http.Request) {
	hndlr, _ := core.NewHandler(&core.Config{
		EnableBlocking: true,
		TCPOnly:        true,
		Upstream:       []string{"1.1.1.1:53", "1.0.0.1:53", "8.8.8.8:53", "8.8.4.4:53"},
		Verbose:        false,
		UserAgent:      config.GetUserAgent(),
		LogGuessedIP:   false,
		Timeout:        60,
		Tries:          10,
	})
	hndlr.ServeHTTP(w, r)
}
