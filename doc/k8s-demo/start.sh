#!/usr/bin/env bash

get_first_pod_for_app() {
    echo $(kubectl get --namespace quick-start po -l=app="$1" -o=jsonpath='{$.items[0].metadata.name}')
}

wait_for_app() {
    while [[ ! $(kubectl get --namespace quick-start pod -l=app="$1" | grep Running) ]] ; do
        echo "Waiting for $1"
        sleep 3
    done
    echo Ready!
}

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

cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: Secret
metadata:
  name: quick-start-backend-credentials
  namespace: quick-start
type: Opaque
data:
  address: $(echo -n quick-start-backend.quick-start.svc.cluster.local:5432 | base64)
  username: $(echo -n quick_start | base64)
  password: $(echo -n quick_start | base64)
EOF

kubectl apply -f pg.yml
wait_for_app quick-start-backend
sleep 10
# CREATE USER
kubectl --namespace quick-start \
    exec -it $(get_first_pod_for_app quick-start-backend) -- psql -U postgres -c "CREATE USER quick_start PASSWORD 'quick_start'; GRANT ALL ON SCHEMA public to quick_start;"

# CREATE TABLE
kubectl --namespace quick-start \
    exec -it $(get_first_pod_for_app quick-start-backend) -- psql -U postgres -c "create table notes (id serial primary key,title varchar(256),description varchar(1024));"
# GRANT PERMISSIONS
kubectl --namespace quick-start exec -it $(get_first_pod_for_app quick-start-backend) -- psql -U postgres -c "grant all on all tables in schema public to quick_start; grant all on all sequences in schema public to quick_start; grant select on all tables in schema public to quick_start;"


rotate_password() {
    db_username=quick_start
    new_password="password$RANDOM"

    cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: Secret
metadata:
  name: quick-start-backend-credentials
  namespace: quick-start
type: Opaque
data:
  address: $(echo -n quick-start-backend.quick-start.svc.cluster.local:5432 | base64)
  username: $(echo -n quick_start | base64)
  password: $(echo -n "$new_password" | base64)
EOF

    kubectl --namespace quick-start exec -it $(get_first_pod_for_app quick-start-backend) -- psql -U postgres -c "ALTER ROLE $db_username WITH PASSWORD '$new_password'; SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE pid <> pg_backend_pid();"
}

# BUILD APPLICATION
docker build -t quick-start-app:latest .

kubectl apply -f quick-start.yml
wait_for_app quick-start-application

# kubectl --namespace quick-start     exec -it $(get_first_pod_for_app quick-start-application) -c quick-start-application -- bash
# ./main
# curl localhost:8080/note
# curl -d "title=value1&description=value2" -X POST localhost:8080/note
