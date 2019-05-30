FROM golang:1.12.5-stretch as secretless-builder
MAINTAINER Conjur Inc.
LABEL builder="secretless-builder"

WORKDIR /secretless

# TODO: Expand this with build args when we support other arches
ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

COPY go.mod go.sum /secretless/

RUN go mod download

# secretless source files
COPY ./cmd /secretless/cmd
COPY ./internal /secretless/internal
COPY ./pkg /secretless/pkg
COPY ./resource-definitions /secretless/resource-definitions

RUN go build -o dist/$GOOS/$GOARCH/secretless-broker ./cmd/secretless-broker && \
    go build -o dist/$GOOS/$GOARCH/summon2 ./cmd/summon2


# =================== MAIN CONTAINER ===================
FROM alpine:3.8 as secretless-broker
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
    # Make and setup a directory for the Conjur client certificate/access token
    mkdir -p /etc/conjur/ssl && \
    mkdir -p /run/conjur && \
    # Use GID of 0 since that is what OpenShift will want to be able to read things
    chown secretless:0 /usr/local/lib/secretless \
                       /sock \
                       /etc/conjur/ssl \
                       /run/conjur && \
    # We need open group permissions in these directories since OpenShift won't
    # match our UID when we try to write files to them
    chmod 770 /sock \
              /etc/conjur/ssl \
              /run/conjur

USER secretless

ENTRYPOINT [ "/usr/local/bin/secretless-broker" ]

COPY --from=secretless-builder /secretless/dist/linux/amd64/secretless-broker \
                               /secretless/dist/linux/amd64/summon2 /usr/local/bin/
