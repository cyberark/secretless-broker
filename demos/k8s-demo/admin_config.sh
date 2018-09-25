#!/bin/bash -e

# application url accessible to local machine
get_APPLICATION_URL() { # CHANGE to reflect endpoint exposed by application service
  local url=$(minikube service -n quick-start-application-ns quick-start-application --url)
  echo "${url#"http://"}"
}

# database url accessible to kubernetes cluster and local machine
get_REMOTE_DB_URL() { # CHANGE to reflect endpoint exposed by db service
  local url=$(minikube service -n quick-start-backend-ns quick-start-backend --url)
  echo "${url#"http://"}"/quick_start_db
}

# admin-user credentials
DB_ADMIN_USER=postgres
DB_ADMIN_PASSWORD=admin_password

# application-user credentials
DB_USER=quick_start
DB_INITIAL_PASSWORD=quick_start

# Run this to access postgres as admin_user
#
# kubectl run --rm -it \
# psql-client --env PGPASSWORD=${DB_ADMIN_PASSWORD} --image=postgres:9.6 --restart=Never \
# --command -- psql     -U ${DB_ADMIN_USER} "postgres://$DB_URL"
