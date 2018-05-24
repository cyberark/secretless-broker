#!/bin/bash -ex

platform=$(go run ../print_platform.go)

hcv_host_port=$(docker-compose port vault 8200)
hcv_port=$(echo "$hcv_host_port" | go run ../conjur_provider/parse_port.go)

exec env VAULT_ADDR="localhost:$hcv_port" \
  "../../bin/$platform/amd64/secretless" \
  -config secretless.dev.yml
