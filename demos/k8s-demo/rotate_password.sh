#!/bin/bash -e

new_password="$1"

if [[ -z $new_password ]]; then
  echo "usage: $0 [new-password]"
  exit 1
fi

. ./admin_config.sh
. ./utils.sh
REMOTE_DB_URL=$(get_REMOTE_DB_URL)

# update db credentials
kubectl run \
 --rm \
 -i \
 update-db-credentials-db-client-${RANDOM} \
   --env PGPASSWORD=${DB_ADMIN_PASSWORD} \
   --image=postgres:9.6 \
   --restart=Never \
   --command \
   -- psql \
     -U ${DB_ADMIN_USER} \
     "postgres://$REMOTE_DB_URL" \
     << EOL
ALTER ROLE $DB_USER WITH PASSWORD '$new_password';
EOL

base64_new_password=$(echo -n "${new_password}" | base64)
new_password_json='{"data":{"password": "'${base64_new_password}'"}}'

# update stored credentials
kubectl --namespace quick-start-application-ns \
 patch secret \
 quick-start-backend-credentials \
 -p="${new_password_json}"

# prune open connections
kubectl run \
 --rm \
 -i \
 prune-open-connections-db-client-${RANDOM} \
   --env PGPASSWORD=${DB_ADMIN_PASSWORD} \
   --image=postgres:9.6 \
   --restart=Never \
   --command \
   -- psql \
     -U ${DB_ADMIN_USER} \
     "postgres://$REMOTE_DB_URL" \
     << EOL
SELECT
    pg_terminate_backend(pid)
FROM
    pg_stat_activity
WHERE
    pid <> pg_backend_pid()
AND
    usename='$DB_USER';
EOL
