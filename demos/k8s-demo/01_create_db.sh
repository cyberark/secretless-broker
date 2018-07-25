#!/usr/bin/env bash

. ./utils.sh
. ./admin_config.sh

# create namespace
echo ">>--- Clean up quick-start-db namespace"

kubectl delete namespace quick-start-db
while [[ $(kubectl get namespace quick-start-db 2>/dev/null) ]] ; do
    echo "Waiting for quick-start-db namespace clean up"
    sleep 10
done
kubectl create namespace quick-start-db

echo Ready!

# create database
echo ">>--- Create database"

kubectl apply -f etc/pg.yml

# Wait for it
wait_for_app quick-start-backend quick-start-db
sleep 3

kubectl --namespace quick-start-db \
    exec -it $(get_first_pod_for_app quick-start-backend quick-start-db) -- psql -U ${DB_ADMIN_USER} -c "
CREATE DATABASE quick_start_db;
"
