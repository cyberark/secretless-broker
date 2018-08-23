#!/bin/bash -e

new_password="$1"

if [[ -z $new_password ]]; then
  echo "usage: $0 [new-password]"
  exit 1
fi

. ./utils.sh
. ./admin_config.sh

# update db credentials
docker run \
 --rm \
 -i \
 -e PGPASSWORD=${DB_ADMIN_PASSWORD} \
 postgres:9.6 \
  psql \
  -U ${DB_ADMIN_USER} \
  "postgres://$DB_URL" \
  -c "
ALTER ROLE $DB_USER WITH PASSWORD '$new_password';
"

base64_new_password=$(echo -n "${new_password}" | base64)
new_password_json='{"data":{"password": "'${base64_new_password}'"}}'

# update stored credentials
kubectl --namespace quick-start-application-ns \
 patch secret \
 quick-start-backend-credentials \
 -p="${new_password_json}"

# prune open connections
docker run \
 --rm \
 -i \
 -e PGPASSWORD=${DB_ADMIN_PASSWORD} \
 postgres:9.6 \
  psql \
  -U ${DB_ADMIN_USER} \
  "postgres://$DB_URL" \
  -c "
SELECT
    pg_terminate_backend(pid)
FROM
    pg_stat_activity
WHERE
    pid <> pg_backend_pid()
AND
    usename='$DB_USER';
"
