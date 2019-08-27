GO_FLAGS := -a -v

.PHONY: deploy

build:
	env GO111MODULES=on CGO_ENABLED=0 GOOS=linux go build $(GO_FLAGS) -ldflags="-s -w" -o bin/dns-query lambda.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
