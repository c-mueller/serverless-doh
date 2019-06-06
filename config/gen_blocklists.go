// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
)

var blocklists = []string{
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	"https://mirror1.malwaredomains.com/files/justdomains",
	"http://sysctl.org/cameleon/hosts",
	"https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
	"https://hosts-file.net/ad_servers.txt",
}

//const template = `
//package config
//
//import (
//	"encoding/base64"
//	"github.com/vmihailenco/msgpack"
//)
//
//var data = "%s"
//
//func init() {
//	var dmap map[string]bool
//	byteData, _ := base64.StdEncoding.DecodeString(data)
//	_ = msgpack.Unmarshal(byteData, &dmap)
//	Blocklists = dmap
//}`
//
//func main() {
//	blockmap, _ := generateBlockageMap(blocklists)
//
//	d , _ := msgpack.Marshal(blockmap)
//	b64data:= base64.StdEncoding.EncodeToString(d)
//
//	ioutil.WriteFile("generated_blocklists.go",[]byte(fmt.Sprintf(template,b64data)),0555)
//}

const template = `
package config

func init() {
	Blocklists = make(map[string]bool)
%s
}`

func main() {
	blockmap, _ := generateBlockageMap(blocklists)
	valuestr := ""
	total := len(blockmap)
	ctr := 0
	perc := -1.0
	for k, _ := range blockmap {
		valuestr += fmt.Sprintf("\tBlocklists[\"%s\"] = true\n", k)
		newPerc := math.Floor((float64(ctr) / float64(total)) * 100)
		if newPerc > perc {
			perc = newPerc
			fmt.Printf("%f%% (%d of %d)\n", perc, ctr, total)
		}
		ctr++
	}
	ioutil.WriteFile("generated_blocklists.go", []byte(fmt.Sprintf(template, valuestr)), 0555)
}

func generateBlockageMap(blocklistUrls []string) (map[string]bool, error) {
	blockageMap := make(map[string]bool, 0)
	for _, blocklistURL := range blocklistUrls {
		content, err := http.Get(blocklistURL)
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(content.Body)
		if err != nil {
			return nil, err
		}

		parseBlockFile(data, blockageMap)
	}

	return blockageMap, nil
}

func parseBlockFile(data []byte, blockageMap map[string]bool) {
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
		blockageMap[url] = true
		urlCount++
	}
}

func cleanHostsLine(line string) string {
	ln := strings.TrimSuffix(line, " ")
	ln = strings.Replace(line, " ", "\t", -1)
	ln = strings.Replace(ln, "\r", "", -1)
	return ln
}
