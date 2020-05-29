FROM golang:1.13-stretch as secretless-builder
MAINTAINER Conjur Inc.
LABEL builder="secretless-builder"

WORKDIR /secretless

# TODO: Expand this with build args when we support other arches
ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

COPY go.mod go.sum /secretless/
COPY third_party/ /secretless/third_party

RUN go mod download

# secretless source files
COPY ./cmd /secretless/cmd
COPY ./internal /secretless/internal
COPY ./pkg /secretless/pkg
COPY ./resource-definitions /secretless/resource-definitions

ARG TAG="dev"

# The `Tag` override is there to provide the git commit information in the
# final binary. See `Static long version tags` in the `Building` section
# of `CONTRIBUTING.md` for more information.
RUN go build -ldflags="-X github.com/cyberark/secretless-broker/pkg/secretless.Tag=$TAG" \
             -o dist/$GOOS/$GOARCH/secretless-broker ./cmd/secretless-broker && \
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

# =================== MAIN CONTAINER (REDHAT) ===================
FROM registry.access.redhat.com/rhel as secretless-broker-redhat
MAINTAINER CyberArk Software, Inc.

ARG VERSION

LABEL name="Secretless-broker"
LABEL vendor="CyberArk"
LABEL version="$VERSION"
LABEL release="$VERSION"
LABEL summary="Secure your apps by making them Secretless"
LABEL description="Secretless Broker is a connection broker which relieves client \
applications of the need to directly handle secrets to target services"

    # Add Limited user
RUN groupadd -r secretless \
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
    mkdir -p /licenses && \
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

COPY LICENSE /licenses

USER secretless

ENTRYPOINT [ "/usr/local/bin/secretless-broker -debug=true" ]

COPY --from=secretless-builder /secretless/dist/linux/amd64/secretless-broker \
                               /secretless/dist/linux/amd64/summon2 /usr/local/bin/
