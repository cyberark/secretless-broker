#!/bin/bash -ex

godep restore

echo "Building for linux + amd64"
env GOOS=linux GOARCH=amd64 go install ./cmd/secretless

mkdir -p bin/linux/amd64
cp $GOPATH/bin/secretless bin/linux/amd64
