#!/bin/bash -ex

db_password=$(<../secrets/db.password)

export DB_PASSWORD="$db_password"
export DB_HOST=pg

docker-compose run --rm \
  --entrypoint ./makedb.sh \
  myapp

docker-compose up --no-deps -d myapp
