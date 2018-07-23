# build stage
FROM golang:latest AS build-env
RUN go get -u github.com/lib/pq github.com/gorilla/mux
RUN mkdir /src
ADD . /src/
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# final stage
FROM scratch
COPY --from=build-env /src/app /
ENTRYPOINT [ "/app" ]
