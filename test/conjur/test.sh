#!/bin/bash -ex

docker-compose build
docker-compose up -d conjur

function wait_for_conjur() {
  for i in $(seq 20); do
    if ! docker-compose exec conjur curl -o /dev/null -fs -X OPTIONS http://localhost > /dev/null; then
      echo .
      sleep 2
    else
      break
    fi
  done

  # Fail if the server isn't up yet
  docker-compose exec conjur curl -o /dev/null -fs -X OPTIONS http://localhost > /dev/null
}

wait_for_conjur

admin_api_key=$(docker-compose exec conjur conjurctl role retrieve-key dev:user:admin | tr -d '\r')
export CONJUR_AUTHN_API_KEY=$admin_api_key

docker-compose up -d secretless

sleep 2

docker-compose run --rm \
  test conjur variable values add db/password secret

docker-compose run --rm \
  -e TEST_CONJUR_AUTHN_API_KEY=$admin_api_key \
  test bash -c "env http_proxy= godep restore && cd test/conjur && go test"
