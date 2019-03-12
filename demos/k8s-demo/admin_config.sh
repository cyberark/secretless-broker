#!/bin/bash


# application url accessible to local machine
export APPLICATION_URL=192.168.99.100:30002 # CHANGE to reflect endpoint exposed by application service

# database url accessible to kubernetes cluster and local machine
export DB_URL=192.168.99.100:30001/quick_start_db # CHANGE to reflect endpoint exposed by db service

# admin-user credentials
export DB_ADMIN_USER=postgres
export DB_ADMIN_PASSWORD=admin_password

# application-user credentials
export DB_USER=quick_start
export DB_INITIAL_PASSWORD=quick_start

# Run this to access postgres as admin_user
# docker run --rm -it -e PGPASSWORD=${DB_ADMIN_PASSWORD} postgres:9.6 psql -U ${DB_ADMIN_USER} "postgres://$DB_URL"
