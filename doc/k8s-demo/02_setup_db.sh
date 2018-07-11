#!/usr/bin/env bash

. ./config.sh
. ./utils.sh

# setup db
echo ">>--- Set up database"

docker run --rm -it postgres:9.6 env \
 PGPASSWORD=$DB_ROOT_PASSWORD psql -U $DB_ROOT_USER "postgres://$DB_URL" -c "
/* Clean DB */
REVOKE ALL ON ALL TABLES IN SCHEMA PUBLIC FROM $DB_USER;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA PUBLIC FROM $DB_USER;
REVOKE SELECT ON ALL TABLES IN SCHEMA PUBLIC FROM $DB_USER;
REVOKE ALL ON SCHEMA PUBLIC FROM $DB_USER;

DROP TABLE IF EXISTS notes;
DROP USER IF EXISTS $DB_USER;

/* Create Application User */
CREATE USER $DB_USER PASSWORD '$DB_INITIAL_PASSWORD';
GRANT ALL ON SCHEMA PUBLIC TO $DB_USER;

/* Create Table */
CREATE TABLE notes (
    id serial primary key,
    title varchar(256),
    description varchar(1024)
);

/* Grant Permissions */
GRANT ALL ON ALL TABLES IN SCHEMA PUBLIC TO $DB_USER;
GRANT ALL ON ALL SEQUENCES IN SCHEMA PUBLIC TO $DB_USER;
GRANT SELECT ON ALL TABLES IN SCHEMA PUBLIC TO $DB_USER;
"

# create namespace
echo ">>--- Clean up quick-start namespace"

kubectl delete namespace quick-start --ignore-not-found=true
while [[ $(kubectl get namespace quick-start 2>/dev/null) ]] ; do
    echo "Waiting for quick-start namespace clean up"
    sleep 10
done
kubectl create namespace quick-start

echo Ready!

# store db credentials
update_password_k8s_secret $DB_INITIAL_PASSWORD
