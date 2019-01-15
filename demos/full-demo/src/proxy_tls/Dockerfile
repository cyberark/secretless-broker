FROM golang:1.11.4

EXPOSE 443

WORKDIR /go/src/proxy_tls
ENTRYPOINT [ "./proxy_tls" ]

COPY . .
RUN go build
