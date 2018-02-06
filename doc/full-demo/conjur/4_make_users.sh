#!/bin/bash -ex

docker-compose run --rm client curl -X POST myapp --data @alice.json
docker-compose run --rm client curl -X POST myapp --data @bob.json

docker-compose run --rm client curl myapp
