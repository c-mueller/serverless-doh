GO_FLAGS := -v

.PHONY: deploy

generate:
	cd .. && go generate ./...

build:
	GO111MODULES=on GOOS=linux go build $(GO_FLAGS) -ldflags="-s -w" -o bin/dns-query lambda.go

update: build
	sls deploy -v -f dns-query

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
