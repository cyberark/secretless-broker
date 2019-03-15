#!/bin/bash

MINIKUBE_IP="$(minikube ip)"
export MINIKUBE_IP

# database url accessible to kubernetes cluster and local machine
# NOTE: Defined in pg.yml as nodePort
export DB_URL="$MINIKUBE_IP":30001/quick_start_db

# admin-user credentials
export DB_ADMIN_USER=postgres
export DB_ADMIN_PASSWORD=admin_password

# application-user credentials
export DB_USER=quick_start
export DB_INITIAL_PASSWORD=quick_start

# Run this to access postgres as admin_user
# docker run --rm -it -e PGPASSWORD=${DB_ADMIN_PASSWORD} postgres:9.6 psql -U ${DB_ADMIN_USER} "postgres://$DB_URL"
