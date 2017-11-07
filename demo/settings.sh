#!/bin/bash -ex

db_username=alice
db_subnet_group_name=default-vpc-eb2d9893

rds_endpoint_url=http://rds.amazonaws.com

function find_db_endpoint() {
  db_instance_id=$(cat tmp/db_instance.json | jq -r .DBInstance.DBInstanceIdentifier)

  local endpoint=$(aws rds describe-db-instances \
    --db-instance-identifier $db_instance_id \
    --endpoint-url $rds_endpoint_url | jq -r .DBInstances[0].Endpoint.Address)

  if [[ "$endpoint" = "null" ]]; then
    printf '.'
    return 3
  else
    db_endpoint="$endpoint"
    return 0
  fi
}
