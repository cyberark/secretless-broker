#!/usr/bin/env bash

. ./utils.sh

kubectl delete namespace quick-start
while [[ $(kubectl get namespace quick-start) ]] ; do
    echo "Waiting for quick-start namespace clean up"
    sleep 10
done
echo Ready!

kubectl create namespace quick-start

kubectl create configmap quick-start-application-secretless-config \
  --namespace quick-start \
  --from-file=secretless.yml

update_password_k8s_secret quick_start

kubectl apply -f pg.yml
wait_for_app quick-start-backend
sleep 10
# CREATE USER
kubectl --namespace quick-start \
    exec -it $(get_first_pod_for_app quick-start-backend) -- psql -U postgres -c "CREATE USER $DB_USERNAME PASSWORD '$DB_INITIAL_PASSWORD'; GRANT ALL ON SCHEMA public to $DB_USERNAME;"

# CREATE TABLE
kubectl --namespace quick-start \
    exec -it $(get_first_pod_for_app quick-start-backend) -- psql -U postgres -c "create table notes (id serial primary key,title varchar(256),description varchar(1024));"
# GRANT PERMISSIONS
kubectl --namespace quick-start exec -it $(get_first_pod_for_app quick-start-backend) -- psql -U postgres -c "grant all on all tables in schema public to quick_start; grant all on all sequences in schema public to quick_start; grant select on all tables in schema public to quick_start;"


# BUILD APPLICATION
docker build -t quick-start-app:latest .

kubectl apply -f quick-start.yml
wait_for_app quick-start-application

# kubectl --namespace quick-start exec -it $(get_first_pod_for_app quick-start-application) -c quick-start-application -- bash
# ./main
# curl localhost:8080/note
# curl -d '{"title":"value1", "description":"value2"}' -H "Content-Type: application/json" -X POST http://localhost:8080/note
