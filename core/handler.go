package core

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/m13253/dns-over-https/json-dns"
	"github.com/miekg/dns"
)

type config struct {
	Upstream         []string `toml:"upstream"`
	Timeout          uint     `toml:"timeout"`
	Tries            uint     `toml:"tries"`
	TCPOnly          bool     `toml:"tcp_only"`
	Verbose          bool     `toml:"verbose"`
	LogGuessedIP     bool     `toml:"log_guessed_client_ip"`
	UserAgent string
}

type Handler struct {
	Config    *config
	UDPClient *dns.Client
	TCPClient *dns.Client
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

func NewHandler(conf *config) (*Handler, error) {
	timeout := time.Duration(conf.Timeout) * time.Second
	s := &Handler{
		Config: conf,
		UDPClient: &dns.Client{
			Net:     "udp",
			UDPSize: dns.DefaultMsgSize,
			Timeout: timeout,
		},
		TCPClient: &dns.Client{
			Net:     "tcp",
			Timeout: timeout,
		},
	}
	return s, nil
}


func (s *Handler) HandleDNSLookup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS, POST")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Handler", s.Config.UserAgent)
	w.Header().Set("X-Powered-By", s.Config.UserAgent)

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

	var req *DNSRequest
	if contentType == "application/dns-json" {
		req = s.parseRequestGoogle(ctx, w, r)
	} else if contentType == "application/dns-message" {
		req = s.parseRequestIETF(ctx, w, r)
	} else if contentType == "application/dns-udpwireformat" {
		req = s.parseRequestIETF(ctx, w, r)
	} else {
		jsonDNS.FormatError(w, fmt.Sprintf("Invalid argument value: \"ct\" = %q", contentType), 415)
		return
	}
	if req.ErrorCode == 444 {
		return
	}
	if req.ErrorCode != 0 {
		jsonDNS.FormatError(w, req.ErrorText, req.ErrorCode)
		return
	}

	req = s.patchRootRD(req)

	var err error
	req, err = s.doDNSQuery(ctx, req)
	if err != nil {
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

func (s *Handler) doDNSQuery(ctx context.Context, req *DNSRequest) (resp *DNSRequest, err error) {
	// TODO(m13253): Make ctx work. Waiting for a patch for ExchangeContext from miekg/dns.
	numServers := len(s.Config.Upstream)
	for i := uint(0); i < s.Config.Tries; i++ {
		req.CurrentUpstream = s.Config.Upstream[rand.Intn(numServers)]

		// Use TCP if always configured to or if the Query type dictates it (AXFR)
		if s.Config.TCPOnly || (s.indexQuestionType(req.Request, dns.TypeAXFR) > -1) {
			req.Response, _, err = s.TCPClient.Exchange(req.Request, req.CurrentUpstream)
		} else {
			req.Response, _, err = s.UDPClient.Exchange(req.Request, req.CurrentUpstream)
			if err == nil && req.Response != nil && req.Response.Truncated {
				log.Println(err)
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
		log.Printf("DNS error from upstream %s: %s\n", req.CurrentUpstream, err.Error())
	}
	return req, err
}
