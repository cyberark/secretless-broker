FROM golang:1.8

WORKDIR /go/src/proxy_tls

EXPOSE 443

COPY . .

RUN go build

EXPOSE 443

ENTRYPOINT [ "./proxy_tls" ]
