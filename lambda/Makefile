GO_FLAGS := -v -a

.PHONY: deploy

generate:
	cd .. && go generate ./...

build:
	env GO111MODULES=on GOOS=linux go build $(GO_FLAGS) -ldflags="-s -w" -o bin/dns-query lambda.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
