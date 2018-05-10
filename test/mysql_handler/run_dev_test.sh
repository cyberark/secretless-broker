#!/bin/bash -ex

function finish {
  rm -rf run/mysql/*
  kill "$!"
}
trap finish EXIT

platform=$(go run ../print_platform.go)

pushd ../..
  go build -o "bin/$platform/amd64/secretless" ./cmd/secretless
popd

./run_dev.sh &

sleep 2

go test -v .
