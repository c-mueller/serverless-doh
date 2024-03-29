package doh

import (
	"context"
	"fmt"
	"github.com/c-mueller/serverless-doh/core/listprovider"
	"github.com/c-mueller/serverless-doh/core/listprovider/providers/static"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/m13253/dns-over-https/json-dns"
	"github.com/miekg/dns"
)

func GetConfigFromEnvironment() *Config {
	var cfg Config
	envconfig.MustProcess("", &cfg)
	return &cfg
}

type Config struct {
	Upstream                 []string `json:"upstream" envconfig:"UPSTREAM_SERVERS" required:"true"`
	Timeout                  uint     `json:"timeout" envconfig:"UPSTREAM_CONNECTION_TIMEOUT" default:"10"`
	Tries                    uint     `json:"tries" envconfig:"UPSTREAM_RESOLUTION_RETRIES" default:"10"`
	TCPOnly                  bool     `json:"tcp_only" envconfig:"UPSTREAM_TCP_ONLY" default:"true"`
	UseTLS                   bool     `json:"use_tls" envconfig:"UPSTREAM_USE_TLS" default:"true"`
	Verbose                  bool     `json:"verbose" envconfig:"VERBOSE" default:"false"`
	LogGuessedIP             bool     `json:"log_guessed_ip" envconfig:"LOG_GUESSED_IP" default:"false"`
	UserAgent                string   `json:"user_agent" envconfig:"APPLICATION_USER_AGENT" default:"sls-doh/1.0.0"`
	EnableBlocking           bool     `json:"enable_blocking" envconfig:"ENABLE_BLOCKING" default:"true"`
	AppendListHeaders        bool     `json:"append_list_headers" envconfig:"RESPONSE_APPEND_LIST_HEADERS" default:"true"`
	AppendQueriedQNameHeader bool     `json:"append_queried_qname_header" envconfig:"RESPONSE_APPEND_QNAME_HEADER" default:"false"`
}

type Handler struct {
	Config     *Config
	UDPClient  *dns.Client
	TCPClient  *dns.Client
	TLSClient  *dns.Client
	IPv4Target net.IP
	IPv6Target net.IP
	Logger     *logrus.Entry
	Provider   listprovider.ListProvider
}

type DNSRequest struct {
	Request         *dns.Msg
	Response        *dns.Msg
	TransactionID   uint16
	CurrentUpstream string
	IsTailored      bool
	ErrorCode       int
	ErrorText       string
}

func NewStaticHandler(conf *Config, logger *logrus.Entry) (*Handler, error) {
	return NewHandler(conf, static.StaticProvider, logger)
}

func NewHandler(conf *Config, provider listprovider.ListProvider, logger *logrus.Entry) (*Handler, error) {
	timeout := time.Duration(conf.Timeout) * time.Second
	tcpmode := "tcp"
	if conf.UseTLS {
		tcpmode = "tcp-tls"
	}
	s := &Handler{
		Config: conf,
		UDPClient: &dns.Client{
			Net:     "udp",
			UDPSize: dns.DefaultMsgSize,
			Timeout: timeout,
		},
		TCPClient: &dns.Client{
			Net:     tcpmode,
			Timeout: timeout,
			//TLSConfig: &tls.Config{
			//	InsecureSkipVerify: true,
			//},
		},
		IPv4Target: net.ParseIP("127.0.0.1"),
		IPv6Target: net.ParseIP("::1"),
		Logger:     logger,
		Provider:   provider,
	}
	return s, nil
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := s.Logger.WithField("stage", "pre-post-processing").WithContext(ctx)
	log.Debug("Processing Request...")

	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS, POST")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Handler", s.Config.UserAgent)
	w.Header().Set("X-Powered-By", s.Config.UserAgent)
	if s.Config.AppendListHeaders {
		w.Header().Set("X-DoH-Blacklist-Item-Count", fmt.Sprintf("%d", s.Provider.GetListInfo().QNamesBlacklisted))
		w.Header().Set("X-DoH-Whitelist-Item-Count", fmt.Sprintf("%d", s.Provider.GetListInfo().QNamesBlacklisted))
		w.Header().Set("X-DoH-Lists-Last-Update", s.Provider.GetListInfo().LastUpdated.String())
	}

	if r.Method == "OPTIONS" {
		w.Header().Set("Content-Length", "0")
		return
	}

	if r.Form == nil {
		const maxMemory = 32 << 20 // 32 MB
		r.ParseMultipartForm(maxMemory)
	}

	contentType := r.Header.Get("Content-Type")
	if ct := r.FormValue("ct"); ct != "" {
		contentType = ct
	}
	if contentType == "" {
		// Guess Request Content-Type based on other parameters
		if r.FormValue("name") != "" {
			contentType = "application/dns-json"
		} else if r.FormValue("dns") != "" {
			contentType = "application/dns-message"
		}
	}
	var responseType string
	for _, responseCandidate := range strings.Split(r.Header.Get("Accept"), ",") {
		responseCandidate = strings.SplitN(responseCandidate, ";", 2)[0]
		if responseCandidate == "application/json" {
			responseType = "application/json"
			break
		} else if responseCandidate == "application/dns-udpwireformat" {
			responseType = "application/dns-message"
			break
		} else if responseCandidate == "application/dns-message" {
			responseType = "application/dns-message"
			break
		}
	}
	if responseType == "" {
		// Guess Response Content-Type based on Request Content-Type
		if contentType == "application/dns-json" {
			responseType = "application/json"
		} else if contentType == "application/dns-message" {
			responseType = "application/dns-message"
		} else if contentType == "application/dns-udpwireformat" {
			responseType = "application/dns-message"
		}
	}

	log.Debugf("Determined Response Type %s", responseType)

	var req *DNSRequest
	if contentType == "application/dns-json" {
		req = s.parseRequestGoogle(ctx, w, r)
	} else if contentType == "application/dns-message" {
		req = s.parseRequestIETF(ctx, w, r)
	} else if contentType == "application/dns-udpwireformat" {
		req = s.parseRequestIETF(ctx, w, r)
	} else {
		log.Errorf("Aborting due to a Format error")
		jsonDNS.FormatError(w, fmt.Sprintf("Invalid argument value: \"ct\" = %q", contentType), 415)
		return
	}
	if req.ErrorCode == 444 {
		return
	}
	if req.ErrorCode != 0 {
		log.Errorf("Aborting due to a invalid response error")
		jsonDNS.FormatError(w, req.ErrorText, req.ErrorCode)
		return
	}

	req = s.patchRootRD(req)

	var err error
	req, err = s.doDNSQuery(ctx, req, w)
	if err != nil {
		log.WithError(err).Error("Aborting due to a DNS Query Failure error")
		jsonDNS.FormatError(w, fmt.Sprintf("DNS query failure (%s)", err.Error()), 503)
		return
	}

	if responseType == "application/json" {
		s.generateResponseGoogle(ctx, w, r, req)
	} else if responseType == "application/dns-message" {
		s.generateResponseIETF(ctx, w, r, req)
	} else {
		panic("Unknown Response Content-Type")
	}
}

func (s *Handler) findClientIP(r *http.Request) net.IP {
	XForwardedFor := r.Header.Get("X-Forwarded-For")
	if XForwardedFor != "" {
		for _, addr := range strings.Split(XForwardedFor, ",") {
			addr = strings.TrimSpace(addr)
			ip := net.ParseIP(addr)
			if jsonDNS.IsGlobalIP(ip) {
				return ip
			}
		}
	}
	XRealIP := r.Header.Get("X-Real-IP")
	if XRealIP != "" {
		addr := strings.TrimSpace(XRealIP)
		ip := net.ParseIP(addr)
		if jsonDNS.IsGlobalIP(ip) {
			return ip
		}
	}
	remoteAddr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	if err != nil {
		return nil
	}
	if ip := remoteAddr.IP; jsonDNS.IsGlobalIP(ip) {
		return ip
	}
	return nil
}

// Workaround a bug causing Unbound to refuse returning anything about the root
func (s *Handler) patchRootRD(req *DNSRequest) *DNSRequest {
	for _, question := range req.Request.Question {
		if question.Name == "." {
			req.Request.RecursionDesired = true
		}
	}
	return req
}

// Return the position index for the question of qtype from a DNS msg, otherwise return -1
func (s *Handler) indexQuestionType(msg *dns.Msg, qtype uint16) int {
	for i, question := range msg.Question {
		if question.Qtype == qtype {
			return i
		}
	}
	return -1
}

func (s *Handler) doDNSQuery(ctx context.Context, req *DNSRequest, rw http.ResponseWriter) (resp *DNSRequest, err error) {
	log := s.Logger.WithField("stage", "do-query").WithContext(ctx)

	dnsQuestion := req.Request.Question[0]
	qname := dnsQuestion.Name
	qname = strings.TrimSuffix(qname, ".")

	if s.Config.AppendQueriedQNameHeader {
		rw.Header().Set("X-QName-Queried", qname)
	}

	if s.Config.EnableBlocking {
		if !s.Provider.MustAllow(qname) && s.Provider.MustBlock(qname) {
			var answers []dns.RR
			if dnsQuestion.Qtype == dns.TypeAAAA {
				answers = aaaa(dnsQuestion.Name, []net.IP{s.IPv6Target})
			} else {
				answers = a(dnsQuestion.Name, []net.IP{s.IPv4Target})
			}

			m := new(dns.Msg)
			m.SetReply(req.Request)
			m.Authoritative, m.RecursionAvailable = true, true
			m.Answer = answers

			// rw.Header().Set("X-QName-Blocked", "true")
			// rw.Header().Set("X-Used-Resolver", "127.0.0.1")

			req.Response = m
			return req, nil
		} else {
			rw.Header().Set("X-QName-Blocked", "false")
		}
	}

	// TODO(m13253): Make ctx work. Waiting for a patch for ExchangeContext from miekg/dns.
	numServers := len(s.Config.Upstream)
	for i := uint(0); i < s.Config.Tries; i++ {
		req.CurrentUpstream = s.Config.Upstream[rand.Intn(numServers)]
		rw.Header().Set("X-Used-Resolver", req.CurrentUpstream)

		// Use TCP if always configured to or if the Query type dictates it (AXFR)
		if s.Config.TCPOnly || (s.indexQuestionType(req.Request, dns.TypeAXFR) > -1) {
			req.Response, _, err = s.TCPClient.Exchange(req.Request, req.CurrentUpstream)
		} else {
			req.Response, _, err = s.UDPClient.Exchange(req.Request, req.CurrentUpstream)
			if err == nil && req.Response != nil && req.Response.Truncated {
				log.WithError(err).Error("DNS Lookup on server %s has failed with an error %s", req.CurrentUpstream, err.Error())
				req.Response, _, err = s.TCPClient.Exchange(req.Request, req.CurrentUpstream)
			}

			// Retry with TCP if this was an IXFR Request and we only received an SOA
			if (s.indexQuestionType(req.Request, dns.TypeIXFR) > -1) &&
				(len(req.Response.Answer) == 1) &&
				(req.Response.Answer[0].Header().Rrtype == dns.TypeSOA) {
				req.Response, _, err = s.TCPClient.Exchange(req.Request, req.CurrentUpstream)
			}
		}
		if err == nil {
			return req, nil
		}
		log.WithError(err).Errorf("DNS error from upstream %s: %s", req.CurrentUpstream, err.Error())
	}
	rw.Header().Set("X-Used-Resolver", "127.0.0.1")
	return req, err
}

func a(zone string, ips []net.IP) []dns.RR {
	var answers []dns.RR
	for _, ip := range ips {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: 3600}
		r.A = ip
		answers = append(answers, r)
	}
	return answers
}

func aaaa(zone string, ips []net.IP) []dns.RR {
	var answers []dns.RR
	for _, ip := range ips {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: 3600}
		r.AAAA = ip
		answers = append(answers, r)
	}
	return answers
}
