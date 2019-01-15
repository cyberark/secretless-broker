FROM golang:1.11.4

EXPOSE 80

RUN apt-get update && \
    apt-get install -y curl \
                       jq \
                       postgresql-client \
                       vim

ENTRYPOINT [ "./myapp" ]

WORKDIR /go/src/myapp

COPY . .

ENV GO111MODULE=on

RUN go mod -sync

RUN go build
