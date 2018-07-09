FROM golang:latest
RUN go get -u github.com/lib/pq github.com/gorilla/mux
RUN mkdir /app
ADD app /app/
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]
