
build:
	CGO_ENABLED=0 go build -v  -ldflags "-s -w"

build_docker:
	cd .. && docker build -t docker.pkg.github.com/c-mueller/serverless-doh/server:latest -f server/Dockerfile .
	cd .. && docker build -t docker.pkg.github.com/c-mueller/serverless-doh/server-strict:latest -f server/Dockerfile.strict .

push_docker: build_docker
	docker push docker.pkg.github.com/c-mueller/serverless-doh/server:latest
	docker push docker.pkg.github.com/c-mueller/serverless-doh/server-strict:latest