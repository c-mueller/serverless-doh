package main

import (
	"github.com/c-mueller/serverless-doh/core/config"
	"github.com/c-mueller/serverless-doh/core/doh"
	"github.com/c-mueller/serverless-doh/core/listprovider/providers/static"
	"github.com/c-mueller/serverless-doh/core/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
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

var logger *logrus.Entry

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logger = logrus.WithField("version", config.Version)
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	log := logger.WithField("stage", "init")

	log.Info("Parsing Configuration...")
	gin.SetMode(gin.ReleaseMode)
	kingpin.Parse()
	cfg := &doh.Config{}
	if *useEnvironment {
		cfg = doh.GetConfigFromEnvironment()
	} else {
		cfg = &doh.Config{
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
	log.WithField("config", cfg).Info("Parsed configuration")

	log.Info("Creating Handlers...")
	cfghndlr, _ := doh.NewStaticHandler(cfg, logger)
	bcfg := *cfg
	bcfg.EnableBlocking = false
	bcfghndlr, _ := doh.NewStaticHandler(&bcfg, logger)

	log.Info("Initializing Gin....")

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(util.LogMiddleware("sls-doh.srv", []util.RegexRule{}, logger.WithField("stage", "http-middleware"), cfg.Verbose))

	blockingFunc := func(ctx *gin.Context) {
		cfghndlr.ServeHTTP(ctx.Writer, ctx.Request)
	}
	nonblockingFunc := func(ctx *gin.Context) {
		bcfghndlr.ServeHTTP(ctx.Writer, ctx.Request)
	}

	engine.GET("dns-query", blockingFunc)
	engine.POST("dns-query", blockingFunc)
	engine.GET("open-dns-query", nonblockingFunc)
	engine.POST("open-dns-query", nonblockingFunc)
	engine.GET("info", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"resolvers":       cfg.Upstream,
			"version":         config.Version,
			"build_timestamp": config.BuildTimestamp,
			"build_context":   config.BuildContext,
			"list_info":       static.StaticProvider.GetListInfo(),
		})
	})

	logger.Infof("Listening for Requests on %s", *endpoint)
	err := engine.Run(*endpoint)
	if err != nil {
		log.WithError(err).Errorf("Server execution stopped due to an error. %q", err.Error())
		os.Exit(1)
	}
}
