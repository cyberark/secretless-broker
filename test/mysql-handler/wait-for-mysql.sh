#!/bin/bash
set -e

cmd="$@"

until docker-compose exec mysql mysqladmin -psecurerootpass status; do
#PGPASSWORD=$POSTGRES_PASSWORD psql -h "$host" -U "postgres" -c '\q'; do
  >&2 echo "MySQL is unavailable - sleeping"
  sleep 1
done

>&2 echo "MySQL is up - executing command"
exec $cmd
