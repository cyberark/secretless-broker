FROM golang:1.8
MAINTAINER Conjur Inc.

COPY ./bin/linux/amd64/secretless /

WORKDIR /

ENTRYPOINT [ "./secretless" ]
