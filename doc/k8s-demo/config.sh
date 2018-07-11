#!/bin/bash -e

DB_URL=192.168.99.100:30001/quick_start_db # Change IP address component accordingly if not running a minikube cluster
DB_ROOT_USER=postgres
DB_ROOT_PASSWORD=postgres
DB_USER=quick_start
DB_INITIAL_PASSWORD=quick_start

# Run this to access postgres
# docker run --rm -it postgres:9.6 env PGPASSWORD=$DB_ROOT_PASSWORD psql -U $DB_ROOT_USER "postgres://$DB_URL"
