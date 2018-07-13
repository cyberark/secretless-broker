FROM golang:1.10.3

EXPOSE 80

RUN apt-get update && \
    apt-get install -y curl \
                       jq \
                       postgresql-client \
                       vim

RUN go get -u github.com/golang/dep/cmd/dep

ENTRYPOINT [ "./myapp" ]

WORKDIR /go/src/myapp

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . .
RUN go build
