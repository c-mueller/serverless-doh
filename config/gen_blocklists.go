// +build ignore

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
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

const blacklistTemplate = `
package %s

import "github.com/c-mueller/serverless-doh/config"

func init() {
	config.ListCreationTimestamp = %d
	
	if config.Blacklist == nil {
		config.Blacklist = make(map[string]bool)
	}
%s
	config.BlacklistItemCount = %d
}`

const pkglistTemplate = `
package registry
import (
%s
)`

const whitelistTemplate = `
package %s

import "github.com/c-mueller/serverless-doh/config"

func init() {
	config.ListCreationTimestamp = %d
	
	if config.Whitelist == nil {
		config.Whitelist = make(map[string]bool)
	}
%s
	config.WhitelistItemCount = %d
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
	pkgs := createFile("generated_blacklists_%03d.go", "\tconfig.Blacklist[\"%s\"] = true\n", blacklistTemplate, "blacklist%02d", bl, 100000)
	if *useWhitelists {
		fmt.Println("Creating Whitelist")
		p := createFile("generated_whitelists_%03d.go", "\tconfig.Whitelist[\"%s\"] = true\n", whitelistTemplate, "whitelist%02d", whitelists, 100000)
		pkgs = append(pkgs, p...)
	}

	pkgsb := strings.Builder{}
	for _, v := range pkgs {
		pkgsb.WriteString(fmt.Sprintf("\t_ %q\n", v))
	}
	ioutil.WriteFile("registry/imports.go", []byte(fmt.Sprintf(pkglistTemplate, pkgsb.String())), 0655)
}

func createFile(filename, lineTemplate, fileTemplate, pkgTemplate string, urls []string, threshold int) []string {
	blacklist, _ := generateMapFromUrls(urls)
	sb := strings.Builder{}
	total := len(blacklist)
	ctr := 0
	fileIdx := 0
	perc := -1.0
	startTime := time.Now()
	fmt.Println()
	packages := make([]string, 0)
	for k, _ := range blacklist {
		sb.WriteString(fmt.Sprintf(lineTemplate, k))
		newPerc := math.Floor((float64(ctr) / float64(total)) * 100)
		if newPerc > perc {
			perc = newPerc
			fmt.Printf("%f%% (%d of %d) Runtime: %s\n", perc, ctr, total, time.Now().Sub(startTime).String())
		}
		if ctr != 0 && (ctr%threshold) == 0 {
			fname := fmt.Sprintf(filename, fileIdx)
			fmt.Printf("Wrote %d Entries. Exceeding Threshold. Writing file %q\n", ctr, fname)
			pkgPath := fmt.Sprintf(pkgTemplate, fileIdx)
			packages = append(packages, fmt.Sprintf("github.com/c-mueller/serverless-doh/config/registry/%s", pkgPath))
			os.MkdirAll(fmt.Sprintf("registry/%s", pkgPath), os.ModePerm)
			ioutil.WriteFile(fmt.Sprintf("registry/%s/%s", pkgPath, fname), []byte(fmt.Sprintf(fileTemplate, pkgPath, time.Now().Unix(), sb.String(), len(blacklist))), 0555)
			sb = strings.Builder{}
			fileIdx++
		}
		ctr++
	}
	fname := fmt.Sprintf(filename, fileIdx)
	fmt.Printf("\nDone Building template in %s. Writing file %q\n", time.Now().Sub(startTime).String(), fname)
	pkgPath := fmt.Sprintf(pkgTemplate, fileIdx)
	packages = append(packages, fmt.Sprintf("github.com/c-mueller/serverless-doh/config/registry/%s", pkgPath))
	os.MkdirAll(fmt.Sprintf("registry/%s", pkgPath), os.ModePerm)
	ioutil.WriteFile(fmt.Sprintf("registry/%s/%s", pkgPath, fname), []byte(fmt.Sprintf(fileTemplate, pkgPath, time.Now().Unix(), sb.String(), len(blacklist))), 0555)

	return packages
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
