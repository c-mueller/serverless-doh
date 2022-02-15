package listprovider

import "time"

type ListProvider interface {
	MustBlock(qname string) bool
	MustAllow(qname string) bool
	GetListInfo() *ListInfo
}

type ListInfo struct {
	QNamesBlacklisted int       `json:"qnames_blacklisted"`
	QNamesWhitelisted int       `json:"qnames_whitelisted"`
	LastUpdated       time.Time `json:"last_updated"`
}
