# Serverless - DNS over HTTPS

This small project is a proof of concept. Implementing a ad blocking DNS over HTTPS server on various Serverless cloud providers including:
- Google Cloud Platform
- AWS

## How does it work?

We use the server side code of m13253's DNS over HTTP implementation. For fulfilling the role of the DoH server. This was possible because all the code handling a DoH request is in a 'http.HandlerFunc'. This implementation just handles the configuration and the binding between the handler function and the DoH handler function.

For ad blocking we use the code from the CoreDNS ads plugin. Due to the stateless nature of functions we have to generate the Blocklists statically they will be compiled into the functions binary. While this prevents dynamic updating it keeps the initialization times very low. Making sure a cold start does not take too long. To update the Blocklists the function will have to be redeployed.

## Deployment guide

Before deploying we have to generate the Blocklist this is done by running the following command in the root directory of the repository.

```
go generate ./...
```

This command ensures the Blocklist used by the functions is generated and up to date.

### Google Cloud Functions

Assuming the `gcloud` CLI is set up run the following command in the repositories root directory

```
gcloud functions deploy HandleDNS --region=europe-west2 --runtime go111 --trigger-http --memory 128
```

Please adjust the region according to your needs. It is important to _not_ change the name and trigger type.
Memory can theoretically be increased however 128 MB seems to be enough to run the function.

### AWS

Assuming the `serverless` CLI is set up run the following commands in the `lambda/` subdirectory of the repository.

```
make
serverless deploy
```

Feel free to change the HTTP path and region in the 'serverless.yml' to fit your needs.

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
