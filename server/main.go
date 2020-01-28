package main

import (
	"github.com/c-mueller/serverless-doh/config"
	"github.com/c-mueller/serverless-doh/core"
	"github.com/gin-gonic/gin"
	"gopkg.in/alecthomas/kingpin.v2"
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
	gin.SetMode(gin.ReleaseMode)
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

	cfghndlr, _ := core.NewHandler(cfg)
	bcfg := *cfg
	bcfg.EnableBlocking = false
	bcfghndlr, _ := core.NewHandler(&bcfg)

	engine := gin.Default()

	blockingFunc := func(ctx *gin.Context) {
		cfghndlr.ServeHTTP(ctx.Writer, ctx.Request)
	}
	nonblockingFunc := func(ctx *gin.Context) {
		bcfghndlr.ServeHTTP(ctx.Writer, ctx.Request)
	}

	engine.Any("dq", blockingFunc)
	engine.Any("dns-query", blockingFunc)
	engine.Any("/", blockingFunc)
	engine.Any("odq", nonblockingFunc)
	engine.Any("open-dns-query", nonblockingFunc)
	engine.GET("info", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"resolvers":                 cfg.Upstream,
			"version":                   config.Version,
			"build_timestamp":           config.BuildTimestamp,
			"list_generation_timestamp": config.ListCreationTimestamp,
			"build_context":             config.BuildContext,
			"blacklist_entry_count":     config.BlacklistItemCount,
			"whitelist_entry_count":     config.WhitelistItemCount,
		})
	})

	err := engine.Run(*endpoint)
	if err != nil {
		panic(err.Error())
	}
}
