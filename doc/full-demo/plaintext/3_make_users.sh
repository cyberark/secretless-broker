#!/bin/bash -e

echo "Creating users 'alice' and 'bob' using a 'curl' request to 'http://myapp'"

docker-compose run --rm client curl -X POST myapp --data @alice.json
docker-compose run --rm client curl -X POST myapp --data @bob.json

echo "Listing all users using a 'curl' request to 'http://myapp'"

docker-compose run --rm client curl myapp
