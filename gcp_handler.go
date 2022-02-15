package serverless_doh

import (
	"github.com/c-mueller/serverless-doh/core/doh"
	"github.com/sirupsen/logrus"
	"net/http"
)

func HandleDNS(w http.ResponseWriter, r *http.Request) {
	hndlr, _ := doh.NewStaticHandler(doh.GetConfigFromEnvironment(), logrus.WithField("platform", "gcp"))
	hndlr.ServeHTTP(w, r)
}
