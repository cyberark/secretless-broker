#!/bin/bash -ex

dep ensure

echo "Building for darwin + amd64"
mkdir -p bin/darwin/amd64
env GOOS=darwin GOARCH=amd64 go build -o bin/darwin/amd64/secretless ./cmd/secretless
env GOOS=darwin GOARCH=amd64 go build -o bin/darwin/amd64/summon2 ./cmd/summon2
