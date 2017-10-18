#!/bin/bash -ex

godep restore
go install

mkdir -p bin/linux/amd64
cp /go/bin/secretless-pg bin/linux/amd64
