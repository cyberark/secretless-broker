#!/usr/bin/env bash

. ./utils.sh

# create namespace
kubectl delete namespace quick-start
while [[ $(kubectl get namespace quick-start) ]] ; do
    echo "Waiting for quick-start namespace clean up"
    sleep 10
done
kubectl create namespace quick-start
echo Ready!

# create database
kubectl apply -f pg.yml
wait_for_app quick-start-backend
sleep 10

# setup db
kubectl --namespace quick-start \
    exec -it $(get_first_pod_for_app quick-start-backend) -- psql -U postgres -c "
/* Create Application User */
CREATE USER $DB_USERNAME PASSWORD '$DB_INITIAL_PASSWORD';

/* Create Table */
CREATE TABLE notes (
    id serial primary key,
    title varchar(256),
    description varchar(1024)
);

/* Grant Permissions */
GRANT ALL ON ALL TABLES IN SCHEMA PUBLIC TO $DB_USERNAME;
GRANT ALL ON ALL SEQUENCES IN SCHEMA PUBLIC TO $DB_USERNAME;
GRANT SELECT ON ALL TABLES IN SCHEMA PUBLIC TO $DB_USERNAME;
"

# store db credentials
update_password_k8s_secret $DB_INITIAL_PASSWORD
