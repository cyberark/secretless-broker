#!/bin/bash -ex

godep restore

echo "Building for darwin + amd64"
env GOOS=darwin GOARCH=amd64 go install ./cmd/secretless
echo "Building for linux + amd64"
env GOOS=linux GOARCH=amd64 go install ./cmd/secretless

mkdir -p bin/darwin/amd64
cp /go/bin/secretless bin/darwin/amd64
mkdir -p bin/linux/amd64
cp /go/bin/secretless bin/linux/amd64
