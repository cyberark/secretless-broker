#!/usr/bin/env bash

. ./admin_config.sh
. ./utils.sh

# setup db
echo ">>--- Set up database"

cat << EOL | docker run \
 --rm \
 -i \
 -e PGPASSWORD=${DB_ADMIN_PASSWORD} \
 postgres:9.6 \
  psql \
  -U ${DB_ADMIN_USER} \
  "postgres://$DB_URL"
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
echo ">>--- Clean up quick-start namespace"

kubectl delete namespace quick-start --ignore-not-found=true
while [[ $(kubectl get namespace quick-start 2>/dev/null) ]] ; do
  echo "Waiting for quick-start namespace clean up"
  sleep 5
done
kubectl create namespace quick-start

echo Ready!

# store db credentials
kubectl --namespace quick-start \
 create secret generic \
 quick-start-backend-credentials \
 --from-literal=address="${DB_URL}" \
 --from-literal=username="${DB_USER}" \
 --from-literal=password="${DB_INITIAL_PASSWORD}"

# grant default service account in quick-start namespace access to quick-start-backend-credentials
cat << EOL | kubectl --namespace quick-start create -f /dev/stdin
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: quick-start-backend-credentials-reader
rules:
- apiGroups: [""] # "" indicates the core API group
  resources: ["secrets"]
  resourceNames: ["quick-start-backend-credentials"]
  verbs: ["get"]
---
# This role binding allows the default serviceAccount to read the "quick-start-backend-credentials" secret in this namespace.
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: read-quick-start-backend-credentials
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: Role
  name: quick-start-backend-credentials-reader
  apiGroup: rbac.authorization.k8s.io
EOL
