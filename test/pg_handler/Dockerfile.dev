FROM secretless-dev

RUN apt-get update && \
    apt-get install -y postgresql-client \
                       postgresql-contrib

RUN go get github.com/ajstarks/svgo/benchviz && \
    go get golang.org/x/tools/cmd/benchcmp
