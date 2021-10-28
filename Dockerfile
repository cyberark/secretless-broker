FROM golang:1.16-buster as secretless-builder
MAINTAINER CyberArk Software Ltd.
LABEL builder="secretless-builder"

# On CyberArk dev laptops, golang module dependencies are downloaded with a
# corporate proxy in the middle. For these connections to succeed we need to
# configure the proxy CA certificate in build containers.
#
# To allow this script to also work on non-CyberArk laptops where the CA
# certificate is not available, we copy the (potentially empty) directory
# and update container certificates based on that, rather than rely on the
# CA file itself.
ADD build_ca_certificate /usr/local/share/ca-certificates/
RUN update-ca-certificates

WORKDIR /secretless

# TODO: Expand this with build args when we support other arches
ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

COPY go.mod go.sum /secretless/
COPY third_party/ /secretless/third_party

RUN go env -w GOFLAGS=-mod=mod
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
FROM alpine:3.14 as secretless-broker
MAINTAINER CyberArk Software Ltd.

RUN apk add -u shadow libc6-compat openssl && \
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
FROM registry.access.redhat.com/ubi8/ubi as secretless-broker-redhat
MAINTAINER CyberArk Software Ltd.

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

ENTRYPOINT [ "/usr/local/bin/secretless-broker" ]

COPY --from=secretless-builder /secretless/dist/linux/amd64/secretless-broker \
                               /secretless/dist/linux/amd64/summon2 /usr/local/bin/
