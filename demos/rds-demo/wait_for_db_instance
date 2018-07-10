#!/bin/bash -ex

source ./settings.sh

if [[ ! -f tmp/db_instance.json ]]; then
  echo "File tmp/db_instance.json does not exist."
  exit 1
fi

echo "Waiting for the database to be available"

export http_proxy

while ! find_db_endpoint; do
  sleep 1
done

echo "Database is available at $db_endpoint"
