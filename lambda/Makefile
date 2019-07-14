.PHONY: build clean deploy

build:
	env GOOS=linux go build -v -ldflags="-s -w" -o bin/dns-query lambda.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
