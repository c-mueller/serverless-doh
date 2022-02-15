package serverless_doh

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/c-mueller/serverless-doh/doh"
)

func HandleDNS(w http.ResponseWriter, r *http.Request) {
	hndlr, _ := doh.NewStaticHandler(doh.GetConfigFromEnvironment(), logrus.WithField("platform", "gcp"))
	hndlr.ServeHTTP(w, r)
}
