#!/bin/bash -ex

docker-compose run --rm --no-deps \
  -e CONJUR_APPLIANCE_URL=http://conjur \
  test bash -c "env http_proxy= dep ensure && cd test/conjur && go test"
