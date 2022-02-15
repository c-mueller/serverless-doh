package static

import (
	"github.com/c-mueller/serverless-doh/core/listprovider"
	"time"
)

var (
	StaticProvider = &provider{
		Blacklist:             make(map[string]bool),
		Whitelist:             make(map[string]bool),
		ListCreationTimestamp: 0,
		BlacklistItemCount:    0,
		WhitelistItemCount:    0,
	}
)

type provider struct {
	Blacklist             map[string]bool
	Whitelist             map[string]bool
	ListCreationTimestamp int
	BlacklistItemCount    int
	WhitelistItemCount    int
}

func (p *provider) MustBlock(qname string) bool {
	return p.Blacklist[qname]
}

func (p *provider) MustAllow(qname string) bool {
	return p.Whitelist[qname]
}

func (p *provider) GetListInfo() *listprovider.ListInfo {
	return &listprovider.ListInfo{
		QNamesBlacklisted: p.BlacklistItemCount,
		QNamesWhitelisted: p.WhitelistItemCount,
		LastUpdated:       time.Unix(int64(p.ListCreationTimestamp), 0),
	}
}
