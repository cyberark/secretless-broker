FROM golang:1.12.5-alpine as perftool-builder

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

WORKDIR /perftool
ENV CGO_ENABLED=0

RUN apk add --no-cache gcc \
                       git \
                       libc-dev

COPY go.mod go.sum /perftool/
RUN go mod download

# secretless source files
COPY . /perftool/
RUN go build -a -ldflags '-extldflags "-static"' -o juxtaposer ./main.go

# =================== MAIN CONTAINER ===================
FROM alpine:3.9

ENTRYPOINT [ "/bin/juxtaposer" ]

COPY --from=perftool-builder /perftool/juxtaposer /bin/juxtaposer
