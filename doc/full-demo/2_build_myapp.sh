#!/bin/bash -ex

docker-compose build myapp

db_password=$(<secrets/db.password)
env DB_PASSWORD="$db_password" docker-compose up --no-deps -d myapp_secretless

docker-compose run --rm --entrypoint ./makedb.sh myapp

docker-compose up --no-deps -d myapp
