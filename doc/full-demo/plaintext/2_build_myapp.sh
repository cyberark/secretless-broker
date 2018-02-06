#!/bin/bash -e

db_password=$(<../secrets/db.password)

echo "Loaded database password: $db_password"

export DB_PASSWORD="$db_password"
export DB_HOST=pg

echo "Creating the database schema for 'myapp' by providing it DB_HOST and DB_PASSWORD"

docker-compose run --rm \
  --entrypoint ./makedb.sh \
  myapp

echo "Running 'myapp', providing it DB_HOST and DB_PASSWORD"

docker-compose up --no-deps -d myapp
