#!/bin/bash -ex

source ./settings.sh

db_password=$1

if [[ $db_password = "" ]]; then
  echo "usage: $0 <db-password>"
  exit 1
fi

if [[ ! -f tmp/db_instance.json ]]; then
  echo "File tmp/db_instance.json does not exist."
  exit 1
fi

export http_proxy

find_db_endpoint || exit "Unable to find the DB endpoint"

echo "Storing database password in Conjur"

conjur variable values add pg/url "$db_endpoint:5432/postgres"
conjur variable values add pg/username "$db_username"
conjur variable values add pg/password "$db_password"
