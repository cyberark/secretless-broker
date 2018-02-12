#!/bin/bash -e

# shellcheck disable=SC2034
db_username=alice
db_subnet_group_name=default
rds_endpoint_url=http://rds.amazonaws.com

function find_db_endpoint() {
  local db_instance_id
  local endpoint

  db_instance_id=$(jq -r .DBInstance.DBInstanceIdentifier < tmp/db_instance.json)
  endpoint=$(aws rds describe-db-instances \
    --db-instance-identifier "$db_instance_id" \
    --endpoint-url "$rds_endpoint_url" | jq -r .DBInstances[0].Endpoint.Address)

  if [[ "$endpoint" = "null" || "$endpoint" == "" ]]; then
    printf '.'
    return 3
  else
    db_endpoint="$endpoint"
    return 0
  fi
}
