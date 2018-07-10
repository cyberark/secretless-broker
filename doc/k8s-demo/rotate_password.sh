#!/bin/bash -e

new_password="$1"
#new_password="password$RANDOM"
if [[ -z $new_password ]]; then
  echo "usage: $0 <new-password>"
  exit 1
fi

. ./utils.sh

qs_app=$(get_first_pod_for_app quick-start-application)
qs_backend=$(get_first_pod_for_app quick-start-backend)

update_password_k8s_secret "$new_password"

# wait for secretless to be propagated
while [[ ! "$(kubectl --namespace quick-start exec -it $qs_app -c secretless -- cat /etc/secret/password)" == "$new_password" ]] ; do
    echo "Waiting for secret to be propagated"
    sleep 10
done
echo Ready!

# prune open connections
kubectl --namespace quick-start exec -it $qs_backend -- psql -U postgres -c "ALTER ROLE $DB_USERNAME WITH PASSWORD '$new_password'; SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE pid <> pg_backend_pid();"
