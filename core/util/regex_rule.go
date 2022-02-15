package util

import "regexp"

type RegexRule struct {
	Regex       *regexp.Regexp
	Replacement string
	InLog       bool
}

func (r RegexRule) apply(data []byte, log bool) []byte {
	if !log || log && r.InLog {
		return r.Regex.ReplaceAll(data, []byte(r.Replacement))
	} else {
		return data
	}
}