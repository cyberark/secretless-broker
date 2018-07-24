#!/bin/bash -e

# application url accessible to local machine
APPLICATION_URL=192.168.99.100:30002 # CHANGE to reflect endpoint exposed by application service

# database url accessible to kubernetes cluster and local machine
DB_URL=192.168.99.100:30001/quick_start_db # CHANGE to reflect endpoint exposed by db service

# admin-user credentials
DB_ADMIN_USER=postgres
DB_ADMIN_PASSWORD=admin_password

# application-user credentials
DB_USER=quick_start
DB_INITIAL_PASSWORD=quick_start

# Run this to access postgres
# docker run --rm -it postgres:9.6 env PGPASSWORD=$DB_ADMIN_PASSWORD psql -U $DB_ADMIN_USER "postgres://$DB_URL"
