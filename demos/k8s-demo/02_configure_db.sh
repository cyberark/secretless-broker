#!/usr/bin/env bash

. ./admin_config.sh
. ./utils.sh
REMOTE_DB_URL=$(get_REMOTE_DB_URL)

# setup db
echo ">>--- Set up database"

kubectl run \
 --rm \
 -i \
 set-up-db-db-client-${RANDOM} \
   --env PGPASSWORD=${DB_ADMIN_PASSWORD} \
   --image=postgres:9.6 \
   --restart=Never \
   --command \
   -- psql \
     -U ${DB_ADMIN_USER} \
     "postgres://$REMOTE_DB_URL" \
     << EOL
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
EOL

# create namespace
echo ">>--- Clean up quick-start-application-ns namespace"

kubectl delete namespace quick-start-application-ns --ignore-not-found=true
while [[ $(kubectl get namespace quick-start-application-ns 2>/dev/null) ]] ; do
  echo "Waiting for quick-start-application-ns namespace clean up"
  sleep 5
done
kubectl create namespace quick-start-application-ns

echo Ready!

# store db credentials
kubectl --namespace quick-start-application-ns \
 create secret generic \
 quick-start-backend-credentials \
 --from-literal=address="${REMOTE_DB_URL}" \
 --from-literal=username="${DB_USER}" \
 --from-literal=password="${DB_INITIAL_PASSWORD}"

# create application service account
kubectl --namespace quick-start-application-ns \
  create serviceaccount \
  quick-start-application

# grant quick-start-application service account
# in quick-start-application-ns namespace
# access to quick-start-backend-credentials
kubectl --namespace quick-start-application-ns \
 create \
 -f etc/quick-start-application-entitlements.yml
