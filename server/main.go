package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"

	"github.com/c-mueller/serverless-doh/config"
	"github.com/c-mueller/serverless-doh/core"
)

var (
	endpoint                   = kingpin.Flag("endpoint", "HTTP Server Endpoint").Short('e').Default(":8888").String()
	upstreamServers            = kingpin.Flag("upstream", "Add an Upstream server").Short('u').Default("1.1.1.1:53").Strings()
	verbose                    = kingpin.Flag("verbose", "Enable verbose output").Short('v').Bool()
	disableBlocking            = kingpin.Flag("disable-blocking", "").Bool()
	appendQueriedQnameToHeader = kingpin.Flag("qname-header", "Append the queried QName in the response header").Bool()
	useTLS                     = kingpin.Flag("tls", "Use DNS over TLS servers as Upstream").Short('T').Bool()
	useEnvironment             = kingpin.Flag("env", "Load Configuration from environment variables").Bool()
)

func main() {
	kingpin.Parse()
	cfg := &core.Config{}
	if *useEnvironment {
		cfg = core.GetConfigFromEnvironment()
	} else {
		cfg = &core.Config{
			EnableBlocking:           !*disableBlocking,
			TCPOnly:                  true,
			UseTLS:                   *useTLS,
			Upstream:                 *upstreamServers,
			Verbose:                  *verbose,
			AppendListHeaders:        *verbose,
			UserAgent:                config.GetUserAgent(),
			LogGuessedIP:             false,
			Timeout:                  60,
			Tries:                    10,
			AppendQueriedQNameHeader: *appendQueriedQnameToHeader,
		}
	}

	hndlr, _ := core.NewHandler(cfg)
	err := http.ListenAndServe(*endpoint, hndlr)
	if err != nil {
		panic(err.Error())
	}
}
