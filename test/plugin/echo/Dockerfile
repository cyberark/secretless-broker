FROM golang:1.24-alpine as builder

WORKDIR /go/src/echo

COPY main.go /go/src/echo/

RUN go build -o /go/bin/echo /go/src/echo/main.go

# =================== MAIN CONTAINER ===================
FROM scratch

COPY --from=builder /go/bin/echo /go/bin/echo
ENTRYPOINT ["/go/bin/echo"]
