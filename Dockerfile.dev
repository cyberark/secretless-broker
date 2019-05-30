FROM golang:1.12.5-stretch
MAINTAINER Conjur Inc.

RUN apt-get update && \
    apt-get install -y curl \
                       jq \
                       less \
                       vim

ENV ROOT_DIR=/secretless

WORKDIR $ROOT_DIR

RUN groupadd -r secretless \
             -g 777 && \
    useradd -c "secretless runner account" \
            -g secretless \
            -u 777 \
            -m \
            -r \
            secretless && \
    mkdir -p /usr/local/lib/secretless \
             /sock && \
    chown secretless:secretless /usr/local/lib/secretless \
                                /sock

# these are binaries necessary for development
# the happen to be written in Go:
#
# go-junit-report => Convert go test output to junit xml
# goconvey => Go testing tool
# reflex => Utility for watching files and executing process in response to changes
RUN go get -u github.com/jstemmer/go-junit-report && \
    go get github.com/smartystreets/goconvey && \
    go get github.com/cespare/reflex

# go mod dependency management for the secretless project
COPY go.mod go.sum /secretless/

RUN go mod download

# TODO: all the stuff below this line is not needed
# this image should just be a development environment for Secretless
# and not be a snapshot of the repository
# the repo should be volume mounted
# NOTE: don't forget all the parts of the repo that are dependent on the definition
# of secretless-dev as dev environment + secretless repo snapshot +
# binaries, you'll need to fix them all

# TODO: Expand this with build args when we support other arches
ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

# secretless source files
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./resource-definitions ./resource-definitions

# Not strictly needed but we might as well do this step too since
# the dev may want to run the binary
RUN go build -o dist/$GOOS/$GOARCH/secretless-broker ./cmd/secretless-broker && \
    go build -o dist/$GOOS/$GOARCH/summon2 ./cmd/summon2 && \
    ln -s $ROOT_DIR/dist/$GOOS/$GOARCH/secretless-broker /usr/local/bin/ && \
    ln -s $ROOT_DIR/dist/$GOOS/$GOARCH/summon2 /usr/local/bin/

COPY . .
