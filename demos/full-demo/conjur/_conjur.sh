#!/bin/bash -ex

function conjur_cli() {
  api_key="$1"
  shift

  docker-compose run --rm \
    --no-deps \
    -e CONJUR_AUTHN_API_KEY="$api_key" \
    conjur_client \
    "$@"
}
