#!/usr/bin/env bash

. ./utils.sh
. ./admin_config.sh

# create namespace
echo ">>--- Clean up quick-start-backend-ns namespace"

kubectl delete namespace quick-start-backend-ns
while [[ $(kubectl get namespace quick-start-backend-ns 2>/dev/null) ]] ; do
  echo "Waiting for quick-start-backend-ns namespace clean up"
  sleep 5
done
kubectl create namespace quick-start-backend-ns

echo Ready!

# add pg certificates to kubernetes secrets
kubectl --namespace quick-start-backend-ns \
  create secret generic \
  quick-start-backend-certs \
  --from-file=etc/pg_server.crt \
  --from-file=etc/pg_server.key

# create database
echo ">>--- Create database"

kubectl --namespace quick-start-backend-ns \
 apply -f etc/pg.yml

# Wait for it
wait_for_app quick-start-backend quick-start-backend-ns

kubectl --namespace quick-start-backend-ns \
 exec \
 -i \
 $(get_first_pod_for_app quick-start-backend quick-start-backend-ns) \
 -- \
  psql \
  -U ${DB_ADMIN_USER} \
  -c "
CREATE DATABASE quick_start_db;
"
