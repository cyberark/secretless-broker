#!/bin/bash -e

source ./settings.sh

if [[ ! -f tmp/db_instance.json ]]; then
  echo "File tmp/db_instance.json does not exist."
  exit 1
fi

db_instance_id=$(cat tmp/db_instance.json | jq -r .DBInstance.DBInstanceIdentifier)

echo "Deleting RDS database $db_instance_id"

export http_proxy

aws rds delete-db-instance \
  --db-instance-identifier $db_instance_id \
  --endpoint-url $rds_endpoint_url \
  --skip-final-snapshot

rm tmp/db_instance.json

echo "Deleted database $db_instance_id"

