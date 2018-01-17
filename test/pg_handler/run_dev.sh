#!/bin/bash -ex

platform=$(go run ../print_platform.go)

pg_host_port=$(docker-compose port pg 5432)
pg_port=$(echo "$pg_host_port" | go run ../util/parse_port.go)

exec env PG_ADDRESS="localhost:$pg_port" \
  PG_PASSWORD=test \
  "../../bin/$platform/amd64/secretless" \
  -config secretless.dev.yml
