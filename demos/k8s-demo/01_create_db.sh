#!/usr/bin/env bash

. ./utils.sh
. ./admin_config.sh

##################################################
start_step "Create a new namespace"

delete_ns_and_cleanup "quick-start-backend-ns"
kubectl create namespace "quick-start-backend-ns"

##################################################
start_step "Add certificates to Kubernetes Secrets"

# add pg certificates to kubernetes secrets
kubectl --namespace quick-start-backend-ns \
  create secret generic \
  quick-start-backend-certs \
  --from-file=etc/pg_server.crt \
  --from-file=etc/pg_server.key

##################################################
start_step "Create StatefulSet for Database"

kubectl --namespace quick-start-backend-ns apply -f "etc/pg.yml"

wait_for_app "quick-start-backend" "quick-start-backend-ns"

##################################################
start_step "Create Application Database"

kubectl --namespace quick-start-backend-ns \
 exec -i \
 "$(first_pod quick-start-backend quick-start-backend-ns)" \
 -- \
  psql -U ${DB_ADMIN_USER} \
  -c "CREATE DATABASE quick_start_db;"

##################################################
start_step "Create Database Table and Permissions"

docker run --rm -i \
 -e PGPASSWORD=${DB_ADMIN_PASSWORD} \
 "postgres:9.6" \
    psql \
    -U ${DB_ADMIN_USER} \
    "postgres://$DB_URL" \
    <<EOSQL
/* Create Application User */
CREATE USER ${DB_USER} PASSWORD '${DB_INITIAL_PASSWORD}';

/* Create Table */
CREATE TABLE pets (
    id serial primary key,
    name varchar(256)
);

/* Grant Permissions */
GRANT SELECT, INSERT ON public.pets TO ${DB_USER};
GRANT USAGE, SELECT ON SEQUENCE public.pets_id_seq TO ${DB_USER};
EOSQL

##################################################
start_step "Store DB credentials in Kubernetes Secrets"

delete_ns_and_cleanup "quick-start-application-ns"
kubectl create namespace "quick-start-application-ns"

# Store the credentials
kubectl --namespace quick-start-application-ns \
  create secret generic "quick-start-backend-credentials" \
  --from-literal=address="${DB_URL}" \
  --from-literal=username="${DB_USER}" \
  --from-literal=password="${DB_INITIAL_PASSWORD}"

##################################################
start_step "Create Application Service Account"

# create application service account
kubectl --namespace quick-start-application-ns \
  create serviceaccount "quick-start-application"

# grant "quick-start-application" service account in
# "quick-start-application-ns" namespace access to
# "quick-start-backend-credentials"
kubectl --namespace quick-start-application-ns \
  create -f "etc/quick-start-application-entitlements.yml"

##################################################
start_step "Create and Store Secretless Configuration"

kubectl --namespace quick-start-application-ns \
  create configmap "quick-start-application-secretless-config" \
  --from-file=etc/secretless.yml
