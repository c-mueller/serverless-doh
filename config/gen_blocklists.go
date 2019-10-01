// +build ignore

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	useStrictMode = flag.Bool("strict", false, "Generate file using strict Blacklist")
	useWhitelists = flag.Bool("whitelist", true, "Generate Whitelist")
)

var whitelists = []string{
	"https://files.krnl.eu/whitelist.txt",
}

var blocklists = []string{
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	"https://mirror1.malwaredomains.com/files/justdomains",
	"http://sysctl.org/cameleon/hosts",
	"https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
	"https://hosts-file.net/ad_servers.txt",
}

var strictBlocklists = []string{
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	"https://mirror1.malwaredomains.com/files/justdomains",
	"http://sysctl.org/cameleon/hosts",
	"https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
	"https://hosts-file.net/ad_servers.txt",
	"https://hosts-file.net/grm.txt",
	"https://reddestdream.github.io/Projects/MinimalHosts/etc/MinimalHostsBlocker/minimalhosts",
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/data/KADhosts/hosts",
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/data/add.Spam/hosts",
	"https://v.firebog.net/hosts/static/w3kbl.txt",
	"https://v.firebog.net/hosts/BillStearns.txt",
	"https://www.dshield.org/feeds/suspiciousdomains_Low.txt",
	"https://www.dshield.org/feeds/suspiciousdomains_Medium.txt",
	"https://www.dshield.org/feeds/suspiciousdomains_High.txt",
	"https://www.joewein.net/dl/bl/dom-bl-base.txt",
	"https://raw.githubusercontent.com/matomo-org/referrer-spam-blacklist/master/spammers.txt",
	"https://hostsfile.org/Downloads/hosts.txt",
	"https://someonewhocares.org/hosts/zero/hosts",
	"https://raw.githubusercontent.com/Dawsey21/Lists/master/main-blacklist.txt",
	"https://raw.githubusercontent.com/vokins/yhosts/master/hosts",
	"https://hostsfile.mine.nu/hosts0.txt",
	"https://v.firebog.net/hosts/Kowabit.txt",
	"https://adaway.org/hosts.txt",
	"https://v.firebog.net/hosts/AdguardDNS.txt",
	"https://raw.githubusercontent.com/anudeepND/blacklist/master/adservers.txt",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
	"https://hosts-file.net/ad_servers.txt",
	"https://v.firebog.net/hosts/Easylist.txt",
	"https://pgl.yoyo.org/adservers/serverlist.php?hostformat=hosts;showintro=0",
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/data/UncheckyAds/hosts",
	"https://www.squidblacklist.org/downloads/dg-ads.acl",
	"https://v.firebog.net/hosts/Easyprivacy.txt",
	"https://v.firebog.net/hosts/Prigent-Ads.txt",
	"https://gitlab.com/quidsup/notrack-blocklists/raw/master/notrack-blocklist.txt",
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/data/add.2o7Net/hosts",
	"https://raw.githubusercontent.com/crazy-max/WindowsSpyBlocker/master/data/hosts/spy.txt",
	"https://zerodot1.gitlab.io/CoinBlockerLists/hosts",
}

const blacklistTemplate = `package config
func init() {
	ListCreationTimestamp = %d
	
	if Blacklist == nil {
		Blacklist = make(map[string]bool)
	}
%s
	BlacklistItemCount = %d
}`

const whitelistTemplate = `package config
func init() {
	ListCreationTimestamp = %d
	
	if Whitelist == nil {
		Whitelist = make(map[string]bool)
	}
%s
	WhitelistItemCount = %d
}`

var ValidateQName = regexp.MustCompile("([a-zA-Z0-9]|\\.|-)*").MatchString

func init() {
	flag.Parse()
}

func main() {
	bl := blocklists
	if *useStrictMode {
		bl = strictBlocklists
	}
	fmt.Println("Creating Blacklist")
	createFile("generated_blacklists.go", "\tBlacklist[\"%s\"] = true\n", blacklistTemplate, bl)
	if *useWhitelists {
		fmt.Println("Creating Whitelist")
		createFile("generated_whitelists.go", "\tWhitelist[\"%s\"] = true\n", whitelistTemplate, whitelists)
	}
}

func createFile(filename, lineTemplate, template string, urls []string) {
	blacklist, _ := generateMapFromUrls(urls)
	sb := strings.Builder{}
	total := len(blacklist)
	ctr := 0
	fileIdx := 0
	perc := -1.0
	startTime := time.Now()
	fmt.Println()
	for k, _ := range blacklist {
		sb.WriteString(fmt.Sprintf(lineTemplate, k))
		newPerc := math.Floor((float64(ctr) / float64(total)) * 100)
		if newPerc > perc {
			perc = newPerc
			fmt.Printf("%f%% (%d of %d) Runtime: %s\n", perc, ctr, total, time.Now().Sub(startTime).String())
		}
		ctr++
	}
	fname := fmt.Sprintf(filename, fileIdx)
	fmt.Printf("\nDone Building template in %s. Writing file %q\n", time.Now().Sub(startTime).String(), fname)
	ioutil.WriteFile(filename, []byte(fmt.Sprintf(template, time.Now().Unix(), sb.String(), len(blacklist))), 0644)
}

func generateMapFromUrls(blocklistUrls []string) (map[string]bool, error) {
	blockageMap := make(map[string]bool, 0)
	cntt := len(blocklistUrls)
	for i, blocklistURL := range blocklistUrls {
		fmt.Printf("[%d/%d]: Downloading contents of list %q\n", i+1, cntt, blocklistURL)
		content, err := http.Get(blocklistURL)
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(content.Body)
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%d/%d]: Parsing qnames from list %q\n", i+1, cntt, blocklistURL)
		cnt := parseBlockFile(data, blockageMap)
		fmt.Printf("[%d/%d]: Loaded %d qnames from list %q\n", i+1, cntt, cnt, blocklistURL)
	}
	return blockageMap, nil
}

func parseBlockFile(data []byte, blockageMap map[string]bool) int {
	urlCount := 0
	for _, line := range strings.Split(string(data), "\n") {
		// Skip lines containing comments
		if strings.Contains(line, "#") {
			continue
		}

		ln := cleanHostsLine(line)

		substrings := strings.Split(ln, "\t")

		url := ""

		if len(substrings) == 0 {
			continue
		} else if len(substrings) == 1 {
			url = substrings[0]
		} else {
			i := 1
			for ; len(substrings[i]) == 0 && i < len(substrings)-1; i++ {
				// Count up to determine last index
			}

			if len(substrings) == i {
				continue
			}

			url = substrings[i]
		}

		if url == "" {
			continue
		}

		// Enable blocking for url
		if ValidateQName(url) && utf8.Valid([]byte(url)) {
			blockageMap[url] = true
			urlCount++
		}
	}
	return urlCount
}

func cleanHostsLine(line string) string {
	ln := strings.TrimSuffix(line, " ")
	ln = strings.Replace(line, " ", "\t", -1)
	ln = strings.Replace(ln, "\r", "", -1)
	// Escape quotes to prevent compialtion issues
	// Of course entries containing such data are useless
	ln = strings.Replace(ln, "\"", "\\\"", -1)
	return ln
}
