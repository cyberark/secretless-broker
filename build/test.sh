#!/bin/bash -ex

docker-compose up -d pg conjur

function cleanup {

  if [ ! $? -eq 0 ]; then
    echo "Something went wrong..."
    echo "Logs from conjur:"
    docker-compose logs conjur

    echo
    echo "Logs from secretless_test:"

    docker-compose logs secretless_test
  fi

  if [[ "$DEBUG" != "true" ]]; then
    docker-compose down || true
    rm -f ./run/postgresql/.s.PGSQL.5432
  fi
}
trap cleanup EXIT

while ! docker-compose logs conjur | grep "API key for admin"; do
  sleep 1
done

admin_api_key=$(docker-compose logs conjur | sed -n "s/^.*API key for admin\: \(.*\)$/\1/p")

docker-compose run -T -e CONJUR_AUTHN_API_KEY="$admin_api_key" --rm --no-deps --entrypoint bash client ./example/conjur.sh
secretless_api_key=$(docker-compose run -T -e CONJUR_AUTHN_API_KEY="$admin_api_key" --rm --no-deps client host rotate_api_key -h secretless)

env CONJUR_AUTHN_API_KEY="$secretless_api_key" docker-compose up -d --no-deps secretless_test

sleep 2

docker-compose run -T -e CONJUR_AUTHN_API_KEY="$admin_api_key" --rm --no-deps secretless_dev ./build/test_in_container.sh
