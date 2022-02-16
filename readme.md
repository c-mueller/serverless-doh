# Serverless - DNS over HTTPS

This small project is a proof of concept. Implementing a ad blocking DNS over HTTPS server on various Serverless cloud providers including:
- Google Cloud Platform
- AWS

It can also be deployed as a standalone application to serve as a "monolithic" DoH resolver with ad blocking functionality.
Since the `server` implementation currently only servers HTTP and not HTTPS it is still mandatory to use a reverse proxy like
Caddy or nginx in front to provide encryption of the HTTP traffic. 

## How does it work?

We use the server side code of m13253's DNS over HTTP implementation. For fulfilling the role of the DoH server. This was possible because all the code handling a DoH request is in a 'http.HandlerFunc'. This implementation just handles the configuration and the binding between the handler function and the DoH handler function.

For ad blocking we use the code from the CoreDNS ads plugin. Due to the stateless nature of functions we have to generate the Blocklists statically they will be compiled into the functions binary. While this prevents dynamic updating it keeps the initialization times very low. Making sure a cold start does not take too long. To update the Blocklists the function will have to be redeployed.

## Deployment guide

Before deploying we have to generate the Blocklist this is done by running the following command in the root directory of the repository.

```
go generate ./...
```

This command ensures the Blacklist used by the functions is generated and up to date.
By default this will download the contents of the following blacklists are downloaded:
```
https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts
https://mirror1.malwaredomains.com/files/justdomains
http://sysctl.org/cameleon/hosts
https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist
https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt
https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt
https://hosts-file.net/ad_servers.txt
```

There is also a `strict` option that can be enabled by setting the `STRICT_MODE` environment variable to `true` or `1`. This will download the following blacklists:

```
https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts
https://mirror1.malwaredomains.com/files/justdomains
http://sysctl.org/cameleon/hosts
https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist
https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt
https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt
https://hosts-file.net/ad_servers.txt
https://hosts-file.net/grm.txt
https://reddestdream.github.io/Projects/MinimalHosts/etc/MinimalHostsBlocker/minimalhosts
https://raw.githubusercontent.com/StevenBlack/hosts/master/data/KADhosts/hosts
https://raw.githubusercontent.com/StevenBlack/hosts/master/data/add.Spam/hosts
https://v.firebog.net/hosts/static/w3kbl.txt
https://v.firebog.net/hosts/BillStearns.txt
https://www.dshield.org/feeds/suspiciousdomains_Low.txt
https://www.dshield.org/feeds/suspiciousdomains_Medium.txt
https://www.dshield.org/feeds/suspiciousdomains_High.txt
https://www.joewein.net/dl/bl/dom-bl-base.txt
https://raw.githubusercontent.com/matomo-org/referrer-spam-blacklist/master/spammers.txt
https://hostsfile.org/Downloads/hosts.txt
https://someonewhocares.org/hosts/zero/hosts
https://raw.githubusercontent.com/Dawsey21/Lists/master/main-blacklist.txt
https://raw.githubusercontent.com/vokins/yhosts/master/hosts
https://hostsfile.mine.nu/hosts0.txt
https://v.firebog.net/hosts/Kowabit.txt
https://adaway.org/hosts.txt
https://v.firebog.net/hosts/AdguardDNS.txt
https://raw.githubusercontent.com/anudeepND/blacklist/master/adservers.txt
https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt
https://hosts-file.net/ad_servers.txt
https://v.firebog.net/hosts/Easylist.txt
https://pgl.yoyo.org/adservers/serverlist.php?hostformat=hosts;showintro=0
https://raw.githubusercontent.com/StevenBlack/hosts/master/data/UncheckyAds/hosts
https://www.squidblacklist.org/downloads/dg-ads.acl
https://v.firebog.net/hosts/Easyprivacy.txt
https://v.firebog.net/hosts/Prigent-Ads.txt
https://gitlab.com/quidsup/notrack-blocklists/raw/master/notrack-blocklist.txt
https://raw.githubusercontent.com/StevenBlack/hosts/master/data/add.2o7Net/hosts
https://raw.githubusercontent.com/crazy-max/WindowsSpyBlocker/master/data/hosts/spy.txt
https://zerodot1.gitlab.io/CoinBlockerLists/hosts
```

Both options will also download this whitelist `https://files.krnl.eu/whitelist.txt`

To use custom settings, just modify the `config/gen_blocklists.go` files `whitelists`,`blacklists` or `strictBlacklists` arrays
with the URLs you want to use.

### Google Cloud Functions

Assuming the `gcloud` CLI is set up run the following command in the repositories root directory

```
gcloud functions deploy HandleDNS --region=europe-west2 --runtime go116 --trigger-http --memory 128 --env-vars-file=gcp_config.yml --security-level=secure-always --timeout=2s
```

Please adjust the region according to your needs. It is important to _not_ change the name and trigger type.
Memory can theoretically be increased however 128 MB seems to be enough to run the function.

The configuration of the resolver, for example to define custom upstream servers, can be done by updating the `gcp_config.yml`.

**IMPORTANT:** Because of memory constraints in the functions compilation container this platform only supports the `default`
set of blacklists. The `strict` set will fail due to lack of memory (out of memory error). 

### AWS

Assuming the `serverless` CLI is set up run the following commands in the `lambda/` subdirectory of the repository.

```
make generate # or "make generate_strict" for strict mode
make deploy
```

The configuration for the DoH resolver is done through environment variables. To change it modify the environment section in the `serverless.yml`.

## Running Standalone

The application can either be built manually by running `go generate ./...` in the repository root and `go build -v` in the `server` subdirectory or you can use the Dockerfiles to build Docker images for the standalone application. For this run one of the following commands in the repository root:

To build with default lists:
```
docker build -t <IMAGE_NAME> -f server/Dockerfile .
```

To build with strict lists:
```
docker build -t <IMAGE_NAME> -f server/Dockerfile.strict .
```

the image is configured using environment variables, just like GCP or AWS lambda. The names are also identical. Please take a look at the
`gcp_config.yml` to find out the meanings for every environment variable. 

By default the application serves HTTP on port 8053 in Docker. This can be changed by changing the `SERVER_PORT` environment variable accordingly.


There are also prebuilt images, however these images are updated irregularly and might contain outdated lists:
```
For default lists:
docker.pkg.github.com/c-mueller/serverless-doh/server:latest

For strict lists:
docker.pkg.github.com/c-mueller/serverless-doh/server-strict:latest
```

## Credits
https://github.com/m13253/dns-over-https

https://github.com/c-mueller/ads

## License

```
MIT License

Copyright (c) 2019 Christian Müller

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
