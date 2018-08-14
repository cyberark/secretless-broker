FROM golang:1.11beta2-stretch as kube-mutating-webhook-builder

ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/cyberark/secretless-broker/kube-mutating-webhook
WORKDIR /go/src/github.com/cyberark/secretless-broker/kube-mutating-webhook

ADD go.* ./
ADD *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kube-mutating-webhook .

FROM alpine:latest

COPY --from=kube-mutating-webhook-builder /go/src/github.com/cyberark/secretless-broker/kube-mutating-webhook/kube-mutating-webhook /kube-mutating-webhook

ENTRYPOINT ["/kube-mutating-webhook"]
