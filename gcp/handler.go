package p

import (
	"github.com/c-mueller/serverless-doh/config"
	"github.com/c-mueller/serverless-doh/core"
	"net/http"
)

func HandleDNS(w http.ResponseWriter, r *http.Request) {
	hndlr, _ := core.NewHandler(&core.Config{
		EnableBlocking: true,
		TCPOnly:        true,
		Upstream:       []string{"1.1.1.1", "8.8.8.8"},
		Verbose:        false,
		UserAgent:      config.GetUserAgent(),
		LogGuessedIP:   false,
		Timeout:        60,
		Tries:          10,
	})
	hndlr.ServeHTTP(w, r)
}
