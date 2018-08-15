FROM golang:1.11beta2-stretch as mutating-webhook-service-builder

ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/cyberark/secretless-broker/cmd/mutating-webhook-service
WORKDIR /go/src/github.com/cyberark/secretless-broker/cmd/mutating-webhook-service

ADD go.* ./
ADD pkg ./pkg
ADD *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mutating-webhook-service .

FROM alpine:latest

COPY --from=mutating-webhook-service-builder /go/src/github.com/cyberark/secretless-broker/cmd/mutating-webhook-service/mutating-webhook-service /mutating-webhook-service

ENTRYPOINT ["/mutating-webhook-service"]
