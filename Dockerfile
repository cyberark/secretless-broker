FROM golang:1.10
MAINTAINER Conjur Inc.

COPY ./bin/linux/amd64/secretless /
COPY ./bin/linux/amd64/summon2 /

WORKDIR /

ENTRYPOINT [ "./secretless" ]
