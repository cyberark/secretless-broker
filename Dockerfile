FROM golang:1.8
MAINTAINER Conjur Inc.

COPY ./bin/linux/amd64/secretless-pg /
COPY ./config.docker.yaml /

WORKDIR /

ENTRYPOINT [ "./secretless-pg", "-config", "config.docker.yaml" ]
