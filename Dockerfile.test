FROM golang:1.12.5-alpine
MAINTAINER Conjur Inc.
LABEL id="secretless-test-runner"

ENTRYPOINT [ "go", "test", "-v", "-timeout", "3m" ]
WORKDIR /secretless

RUN apk add -u curl \
               gcc \
               git \
               mercurial \
               musl-dev

COPY go.mod go.sum /secretless/

RUN go mod download

COPY . .
