FROM golang:1.10.3

EXPOSE 443

WORKDIR /go/src/proxy_tls
ENTRYPOINT [ "./proxy_tls" ]

COPY . .
RUN go build
