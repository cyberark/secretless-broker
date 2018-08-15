FROM golang:1.11beta2-stretch as mutating-webhook-service-builder

ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/cyberark/secretless-broker/sidecar-injector
WORKDIR /go/src/github.com/cyberark/secretless-broker/sidecar-injector

ADD go.* ./
ADD pkg ./pkg
ADD cmd ./cmd

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o secretless-sidecar-injector ./cmd/sidecar-injector/main.go

FROM alpine:latest

COPY --from=mutating-webhook-service-builder /go/src/github.com/cyberark/secretless-broker/sidecar-injector/secretless-sidecar-injector /secretless-sidecar-injector

ENTRYPOINT ["/secretless-sidecar-injector"]
