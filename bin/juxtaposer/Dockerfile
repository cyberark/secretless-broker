FROM golang:1.12.5-alpine as perftool-builder

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
