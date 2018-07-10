FROM golang:1.10.3

RUN apt-get update && apt-get install -y vim curl jq postgresql-client

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/myapp

COPY . .

RUN dep ensure
RUN go build

EXPOSE 80

ENTRYPOINT [ "./myapp" ]
