#!/bin/bash -ex

docker-compose up -d pg

db_password=$(<../secrets/db.password)
docker-compose run --rm --no-deps -e DB_PASSWORD="$db_password" ansible \
  ansible-playbook \
  --key-file "/root/id_insecure" \
  postgresql.yml
