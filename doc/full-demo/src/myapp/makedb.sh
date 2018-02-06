#!/bin/bash -ex

command="psql -U myapp"
if [[ "$DB_HOST" != "" ]]; then
  command="$command -h $DB_HOST"
fi
if [[ "$DB_PASSWORD" != "" ]]; then
  command="env PGPASSWORD=$DB_PASSWORD $command"
fi

$command postgres < userinfo.sql
