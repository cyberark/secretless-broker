#!/usr/bin/env bash

. ./utils.sh
. ./admin_config.sh

# create namespace
echo ">>--- Clean up quick-start-db namespace"

kubectl delete namespace quick-start-db
while [[ $(kubectl get namespace quick-start-db 2>/dev/null) ]] ; do
  echo "Waiting for quick-start-db namespace clean up"
  sleep 5
done
kubectl create namespace quick-start-db

echo Ready!

# create database
echo ">>--- Create database"

kubectl --namespace quick-start-db \
 apply -f etc/pg.yml

# Wait for it
wait_for_app quick-start-backend quick-start-db

kubectl --namespace quick-start-db \
 exec \
 -i \
 $(get_first_pod_for_app quick-start-backend quick-start-db) \
 -- \
  psql \
  -U ${DB_ADMIN_USER} \
  -c "
CREATE DATABASE quick_start_db;
"
