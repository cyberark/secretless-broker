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

conjur_host_port=$(docker-compose port conjur 80)
conjur_port=$(echo $conjur_host_port | go run parse_port.go)

rm -rf tmp
mkdir -p tmp

cat <<CONJURRC > tmp/.conjurrc
url: http://localhost:$conjur_port
account: dev
api_key: $admin_api_key
CONJURRC

docker-compose up -d secretless

sleep 2

docker-compose run --no-deps --rm \
  test conjur variable values add db/password secret
