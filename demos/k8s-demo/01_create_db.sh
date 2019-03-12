#!/usr/bin/env bash

. ./utils.sh
. ./admin_config.sh

start_step() { printf '\n\n>>--- %s\n\n' "$1"; }
finish_step() { printf '\n---\n\n%s' ''; }

start_step "Create a new namespace..."

if kubectl delete namespace quick-start-backend-ns 2> /dev/null; then
  printf '\n%s' "Cleaning up old namespace"

  while kubectl get namespace quick-start-backend-ns > /dev/null 2>&1; do 
    printf "."
    sleep 3
  done

  printf '%s\n\n' "Done"
fi

kubectl create namespace quick-start-backend-ns
finish_step

##################################################

# add pg certificates to kubernetes secrets
kubectl --namespace quick-start-backend-ns \
  create secret generic \
  quick-start-backend-certs \
  --from-file=etc/pg_server.crt \
  --from-file=etc/pg_server.key

##################################################
start_step "Create database"

kubectl --namespace quick-start-backend-ns \
 apply -f etc/pg.yml

# Wait for it
wait_for_app quick-start-backend quick-start-backend-ns
finish_step

##################################################
# kubectl --namespace quick-start-backend-ns \
#  exec -i \
#  "$(first_pod quick-start-backend quick-start-backend-ns)" \
#  -- \
#   psql -U ${DB_ADMIN_USER} \
#   -c "CREATE DATABASE quick_start_db;"
kubectl --namespace quick-start-backend-ns \
exec \
-i \
"$(first_pod quick-start-backend quick-start-backend-ns)" \
-- \
 psql \
 -U ${DB_ADMIN_USER} \
 -c "
CREATE DATABASE quick_start_db;
"
