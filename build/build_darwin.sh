#!/bin/bash -ex

godep restore

echo "Building for darwin + amd64"
env GOOS=darwin GOARCH=amd64 go install ./cmd/secretless

mkdir -p bin/darwin/amd64
cp $GOPATH/bin/secretless bin/darwin/amd64
