FROM golang AS builder
ENV GO111MODULES on
ENV CGO_ENABLED 0
ENV STRICT_MODE 0
WORKDIR /go/src/github.com/c-mueller/serverless-doh
COPY . /go/src/github.com/c-mueller/serverless-doh
RUN go generate ./...
RUN cd server && go build -v -o sls-doh

FROM alpine AS runtime
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/src/github.com/c-mueller/serverless-doh/server/sls-doh sls-doh
EXPOSE 8053
ENV SERVER_PORT=8053
CMD ./sls-doh --env -e ":$SERVER_PORT"