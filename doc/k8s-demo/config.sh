#!/bin/bash -e

DB_URL=$(minikube -n quick-start-db service quick-start-backend --url --format "//{{.IP}}:{{.Port}}" | sed -e 's/^\/\///')/quick_start_db
DB_ROOT_USER=postgres
DB_ROOT_PASSWORD=postgres
DB_USER=quick_start
DB_INITIAL_PASSWORD=quick_start

DB_URL=quick-start-db-example.cog0ithmhjbw.us-east-1.rds.amazonaws.com:5432/quick_start_db
DB_ROOT_USER=quick_start_db
DB_ROOT_PASSWORD=quick_start_db
DB_USER=quick_start_user
DB_INITIAL_PASSWORD=quick_start_user

# Run this to access postgres
# docker run --rm -it postgres:9.6 env PGPASSWORD=$DB_ROOT_PASSWORD psql -U $DB_ROOT_USER "postgres://$DB_URL"
