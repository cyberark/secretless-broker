FROM golang:1.10.3-stretch as secretless-builder
MAINTAINER Conjur Inc.
LABEL builder="secretless-builder"

WORKDIR /go/src/github.com/conjurinc/secretless

RUN curl -fsSL -o /usr/local/bin/dep \
    https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && \
    chmod +x /usr/local/bin/dep

COPY Gopkg.toml Gopkg.lock ./

# TODO: Expand this with build args when we support other arches
ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

RUN dep ensure --vendor-only

COPY . /go/src/github.com/conjurinc/secretless

RUN go build -o dist/$GOOS/$GOARCH/secretless ./cmd/secretless && \
    go build -o dist/$GOOS/$GOARCH/summon2 ./cmd/summon2


# =================== MAIN CONTAINER ===================
FROM alpine:3.7 as secretless
MAINTAINER CyberArk Software, Inc.

WORKDIR /

RUN mkdir -p /lib64 \
    && ln -fs /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ENTRYPOINT [ "/usr/local/bin/secretless" ]

COPY --from=secretless-builder /go/src/github.com/conjurinc/secretless/dist/linux/amd64/secretless \
                               /go/src/github.com/conjurinc/secretless/dist/linux/amd64/summon2 /usr/local/bin/
