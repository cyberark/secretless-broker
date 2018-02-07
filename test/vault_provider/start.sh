#!/bin/bash -ex

docker-compose build
docker-compose up -d vault

function wait_for_vault() {
  for _ in $(seq 20); do
    # curl for /v1/sys/health hangs
    if ! docker-compose run --rm -T dev curl -s http://vault:8200 > /dev/null; then
      echo .
      sleep 2
    else
      break
    fi
  done

  # Fail if the server isn't up yet
  docker-compose run --rm -T dev curl -s http://vault:8200 > /dev/null
}

wait_for_vault

root_token=$(docker-compose logs vault | grep "Root Token:" | tail -n 1 |  cut -c 33-)

function vault_cmd() {
  docker-compose run --rm -T -e VAULT_ADDR=http://vault:8200 -e VAULT_TOKEN="$root_token" --entrypoint vault vault \
    "$@"  
}

vault_host_port=$(docker-compose port vault 8200)
vault_port=$(echo "$vault_host_port" | go run ../util/parse_port.go)

rm -rf tmp
mkdir -p tmp

cat <<ENV > .env
VAULT_ADDR=http://localhost:$vault_port
VAULT_TOKEN=$root_token
ENV

vault_cmd mount kv
vault_cmd write kv/db/password 'password=secret'
vault_cmd write kv/frontend/admin-password 'password=secret'
