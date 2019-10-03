package serverless_doh

import (
	"net/http"

	"github.com/c-mueller/serverless-doh/core"
)

func HandleDNS(w http.ResponseWriter, r *http.Request) {
	hndlr, _ := core.NewHandler(core.GetConfigFromEnvironment())
	hndlr.ServeHTTP(w, r)
}
