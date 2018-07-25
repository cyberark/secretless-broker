FROM golang:1.11beta2-stretch as secretless-builder
MAINTAINER Conjur Inc.
LABEL builder="secretless-builder"

WORKDIR /secretless

# TODO: Expand this with build args when we support other arches
ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

COPY go.mod go.sum /secretless/

# https://github.com/golang/go/issues/26610
RUN go list -e $(go list -f '{{.Path}}' -m all 2>/dev/null)

COPY . /secretless
RUN go build -o dist/$GOOS/$GOARCH/secretless ./cmd/secretless && \
    go build -o dist/$GOOS/$GOARCH/summon2 ./cmd/summon2


# =================== MAIN CONTAINER ===================
FROM alpine:3.8 as secretless
MAINTAINER CyberArk Software, Inc.

RUN apk add -u shadow libc6-compat && \
    # Add Limited user
    groupadd -r secretless \
             -g 777 && \
    useradd -c "secretless runner account" \
            -g secretless \
            -u 777 \
            -m \
            -r \
            secretless && \
    # Ensure plugin dir is owned by secretless user
    mkdir -p /usr/local/lib/secretless && \
    # Make and setup a directory for sockets at /sock
    mkdir /sock && \
    chown secretless:secretless /usr/local/lib/secretless \
                                /sock

USER secretless

ENTRYPOINT [ "/usr/local/bin/secretless" ]

COPY --from=secretless-builder /secretless/dist/linux/amd64/secretless \
                               /secretless/dist/linux/amd64/summon2 /usr/local/bin/
