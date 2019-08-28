package core


import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/m13253/dns-over-https/json-dns"
	"github.com/miekg/dns"
)

func (s *Handler) parseRequestIETF(ctx context.Context, w http.ResponseWriter, r *http.Request) *DNSRequest {
	requestBase64 := r.FormValue("dns")
	requestBinary, err := base64.RawURLEncoding.DecodeString(requestBase64)
	if err != nil {
		return &DNSRequest{
			ErrorCode: 400,
			ErrorText: fmt.Sprintf("Invalid argument value: \"dns\" = %q", requestBase64),
		}
	}
	if len(requestBinary) == 0 && (r.Header.Get("Content-Type") == "application/dns-message" || r.Header.Get("Content-Type") == "application/dns-udpwireformat") {
		requestBinary, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return &DNSRequest{
				ErrorCode: 400,
				ErrorText: fmt.Sprintf("Failed to read Request body (%s)", err.Error()),
			}
		}
	}
	if len(requestBinary) == 0 {
		return &DNSRequest{
			ErrorCode: 400,
			ErrorText: fmt.Sprintf("Invalid argument value: \"dns\""),
		}
	}

	if s.patchDNSCryptProxyReqID(w, r, requestBinary) {
		return &DNSRequest{
			ErrorCode: 444,
		}
	}

	msg := new(dns.Msg)
	err = msg.Unpack(requestBinary)
	if err != nil {
		return &DNSRequest{
			ErrorCode: 400,
			ErrorText: fmt.Sprintf("DNS packet parse failure (%s)", err.Error()),
		}
	}

	if s.Config.Verbose && len(msg.Question) > 0 {
		question := &msg.Question[0]
		questionName := question.Name
		questionClass := ""
		if qclass, ok := dns.ClassToString[question.Qclass]; ok {
			questionClass = qclass
		} else {
			questionClass = strconv.FormatUint(uint64(question.Qclass), 10)
		}
		questionType := ""
		if qtype, ok := dns.TypeToString[question.Qtype]; ok {
			questionType = qtype
		} else {
			questionType = strconv.FormatUint(uint64(question.Qtype), 10)
		}
		var clientip net.IP = nil
		if s.Config.LogGuessedIP {
			clientip = s.findClientIP(r)
		}
		if clientip != nil {
			fmt.Printf("%s - - [%s] \"%s %s %s\"\n", clientip, time.Now().Format("02/Jan/2006:15:04:05 -0700"), questionName, questionClass, questionType)
		} else {
			fmt.Printf("%s - - [%s] \"%s %s %s\"\n", r.RemoteAddr, time.Now().Format("02/Jan/2006:15:04:05 -0700"), questionName, questionClass, questionType)
		}
	}

	transactionID := msg.Id
	msg.Id = dns.Id()
	opt := msg.IsEdns0()
	if opt == nil {
		opt = new(dns.OPT)
		opt.Hdr.Name = "."
		opt.Hdr.Rrtype = dns.TypeOPT
		opt.SetUDPSize(dns.DefaultMsgSize)
		opt.SetDo(false)
		msg.Extra = append([]dns.RR{opt}, msg.Extra...)
	}
	var edns0Subnet *dns.EDNS0_SUBNET
	for _, option := range opt.Option {
		if option.Option() == dns.EDNS0SUBNET {
			edns0Subnet = option.(*dns.EDNS0_SUBNET)
			break
		}
	}
	isTailored := edns0Subnet == nil
	if edns0Subnet == nil {
		ednsClientFamily := uint16(0)
		ednsClientAddress := s.findClientIP(r)
		ednsClientNetmask := uint8(255)
		if ednsClientAddress != nil {
			if ipv4 := ednsClientAddress.To4(); ipv4 != nil {
				ednsClientFamily = 1
				ednsClientAddress = ipv4
				ednsClientNetmask = 24
			} else {
				ednsClientFamily = 2
				ednsClientNetmask = 56
			}
			edns0Subnet = new(dns.EDNS0_SUBNET)
			edns0Subnet.Code = dns.EDNS0SUBNET
			edns0Subnet.Family = ednsClientFamily
			edns0Subnet.SourceNetmask = ednsClientNetmask
			edns0Subnet.SourceScope = 0
			edns0Subnet.Address = ednsClientAddress
			opt.Option = append(opt.Option, edns0Subnet)
		}
	}

	return &DNSRequest{
		Request:       msg,
		TransactionID: transactionID,
		IsTailored:    isTailored,
	}
}

func (s *Handler) generateResponseIETF(ctx context.Context, w http.ResponseWriter, r *http.Request, req *DNSRequest) {
	respJSON := jsonDNS.Marshal(req.Response)
	req.Response.Id = req.TransactionID
	respBytes, err := req.Response.Pack()
	if err != nil {
		log.Println(err)
		jsonDNS.FormatError(w, fmt.Sprintf("DNS packet construct failure (%s)", err.Error()), 500)
		return
	}

	w.Header().Set("Content-Type", "application/dns-message")
	now := time.Now().UTC().Format(http.TimeFormat)
	w.Header().Set("Date", now)
	w.Header().Set("Last-Modified", now)
	w.Header().Set("Vary", "Accept")

	_ = s.patchFirefoxContentType(w, r, req)

	if respJSON.HaveTTL {
		if req.IsTailored {
			w.Header().Set("Cache-Control", "private, max-age="+strconv.FormatUint(uint64(respJSON.LeastTTL), 10))
		} else {
			w.Header().Set("Cache-Control", "public, max-age="+strconv.FormatUint(uint64(respJSON.LeastTTL), 10))
		}
		w.Header().Set("Expires", respJSON.EarliestExpires.Format(http.TimeFormat))
	}

	if respJSON.Status == dns.RcodeServerFailure {
		w.WriteHeader(503)
	}
	w.Write(respBytes)
}

// Workaround a bug causing DNSCrypt-Proxy to expect a Response with TransactionID = 0xcafe
func (s *Handler) patchDNSCryptProxyReqID(w http.ResponseWriter, r *http.Request, requestBinary []byte) bool {
	if strings.Contains(r.UserAgent(), "dnscrypt-proxy") && bytes.Equal(requestBinary, []byte("\xca\xfe\x01\x00\x00\x01\x00\x00\x00\x00\x00\x01\x00\x00\x02\x00\x01\x00\x00\x29\x10\x00\x00\x00\x80\x00\x00\x00")) {
		log.Println("DNSCrypt-Proxy detected. Patching Response.")
		w.Header().Set("Content-Type", "application/dns-message")
		w.Header().Set("Vary", "Accept, User-Agent")
		now := time.Now().UTC().Format(http.TimeFormat)
		w.Header().Set("Date", now)
		w.Write([]byte("\xca\xfe\x81\x05\x00\x01\x00\x01\x00\x00\x00\x00\x00\x00\x02\x00\x01\x00\x00\x10\x00\x01\x00\x00\x00\x00\x00\xa8\xa7\r\nWorkaround a bug causing DNSCrypt-Proxy to expect a Response with TransactionID = 0xcafe\r\nRefer to https://github.com/jedisct1/dnscrypt-proxy/issues/526 for details."))
		return true
	}
	return false
}

// Workaround a bug causing Firefox 61-62 to reject responses with Content-Type = application/dns-message
func (s *Handler) patchFirefoxContentType(w http.ResponseWriter, r *http.Request, req *DNSRequest) bool {
	if strings.Contains(r.UserAgent(), "Firefox") && strings.Contains(r.Header.Get("Accept"), "application/dns-udpwireformat") && !strings.Contains(r.Header.Get("Accept"), "application/dns-message") {
		log.Println("Firefox 61-62 detected. Patching Response.")
		w.Header().Set("Content-Type", "application/dns-udpwireformat")
		w.Header().Set("Vary", "Accept, User-Agent")
		req.IsTailored = true
		return true
	}
	return false
}