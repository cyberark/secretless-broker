#!/bin/bash -e

echo "Bringing up unconfigured Postgresql database in container 'pg'"

docker-compose up -d pg

echo "Configuring the database using Ansible"

db_password=$(<../secrets/db.password)
docker-compose run --rm --no-deps -e DB_PASSWORD="$db_password" ansible \
  ansible-playbook \
  --key-file "/root/id_insecure" \
  postgresql.yml
