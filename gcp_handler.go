package serverless_doh

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/c-mueller/serverless-doh/core"
)

func HandleDNS(w http.ResponseWriter, r *http.Request) {
	hndlr, _ := core.NewHandler(core.GetConfigFromEnvironment(), logrus.WithField("platform", "gcp"))
	hndlr.ServeHTTP(w, r)
}
