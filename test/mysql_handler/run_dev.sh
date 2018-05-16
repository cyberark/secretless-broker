#!/bin/bash -ex

platform=$(go run ../print_platform.go)

pushd ../..
  go build -o "bin/$platform/amd64/secretless" ./cmd/secretless
popd

mysql_host_port=$(docker-compose port mysql 3306)
mysql_port=$(echo "$mysql_host_port" | go run ../util/parse_port.go)

exec env MYSQL_HOST="localhost" \
  MYSQL_PORT="$mysql_port" \
  MYSQL_PASSWORD=testpass \
  "../../bin/$platform/amd64/secretless" \
  -f secretless.dev.yml
