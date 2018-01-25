#!/bin/bash -ex

platform=$(go run ../print_platform.go)

cd ../..

go build -o "bin/$platform/amd64/secretless" ./cmd/secretless

cd -

./run_dev.sh &

sleep 2

go test -v .

kill "$!"
