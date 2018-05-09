#!/bin/bash -ex

function finish {
	kill "$!"
}
trap finish EXIT

platform=$(go run ../print_platform.go)

#rm run/mysql/mysql.sock > /dev/null 2>&1

pushd ../..
  go build -o "bin/$platform/amd64/secretless" ./cmd/secretless
popd

./run_dev.sh &

sleep 2

go test -v .
