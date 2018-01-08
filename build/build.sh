#!/bin/bash -ex

dep ensure

echo "Building for linux + amd64"
mkdir -p bin/linux/amd64
env GOOS=linux GOARCH=amd64 go build -o bin/linux/amd64/secretless ./cmd/secretless

docker-compose -f build/docker-compose.yml build
