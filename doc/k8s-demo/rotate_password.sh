#!/bin/bash -e

new_password="$1"
#new_password="password$RANDOM"
if [[ -z $new_password ]]; then
  echo "usage: $0 <new-password>"
  exit 1
fi

. ./utils.sh

# update stored credentials
update_password_k8s_secret "$new_password"

# wait for stored credentials to be propagated to application workloads
while [[ ! "$(kubectl --namespace quick-start exec -it $(get_first_pod_for_app quick-start-application quick-start) -c secretless -- cat /etc/secret/password)" == "$new_password" ]] ; do
    echo "Waiting for secret to be propagated"
    sleep 10
done
echo Ready!

# prune open connections
docker run --rm -it postgres:9.6 env \
 PGPASSWORD=$DB_ROOT_PASSWORD psql -U $DB_ROOT_USER "postgres://$DB_URL" -c "
ALTER ROLE $DB_USER WITH PASSWORD '$new_password';
SELECT
    pg_terminate_backend(pid)
FROM
    pg_stat_activity
WHERE
    pid <> pg_backend_pid()
AND
    usename='$DB_USER';
"
