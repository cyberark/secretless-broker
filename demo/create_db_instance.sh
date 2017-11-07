#!/bin/bash -e

source ./settings.sh

db_instance_id=$1

password=$2

if [[ $db_instance_id = "" || $password = "" ]]; then
  echo "usage: $0 <db-name> <password>"
  exit 1
fi

if [[ -f tmp/db_instance.json ]]; then
  echo "File tmp/db_instance.json already exists. It looks like you have a demo in progress already"
  exit 1
fi

echo "Creating RDS database with admin user '$db_username' and admin password $password"

export http_proxy

aws rds create-db-instance \
  --db-instance-identifier $db_instance_id \
  --db-instance-class db.t2.micro \
  --engine postgres \
  --publicly-accessible \
  --endpoint-url $rds_endpoint_url \
  --master-username $db_username \
  --master-user-password $password \
  --allocated-storage 10 \
  --db-subnet-group-name $db_subnet_group_name | tee tmp/db_instance.json


db_arn=$(cat tmp/db_instance.json | jq -r .DBInstance.DBInstanceArn)

