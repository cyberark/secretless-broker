#!/bin/bash -e

get_first_pod_for_app() {
    echo $(kubectl get --namespace "$2" po -l=app="$1" -o=jsonpath='{$.items[0].metadata.name}')
}

wait_for_app() {
    while [[ ! $(kubectl get --namespace "$2" pod -l=app="$1" | grep Running) ]] ; do
        echo "Waiting for $1"
        sleep 3
    done
    echo Ready!
}

update_password_k8s_secret() {
    cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: Secret
metadata:
    name: quick-start-backend-credentials
    namespace: quick-start
type: Opaque
data:
    address: $(echo -n $DB_URL | base64)
    username: $(echo -n $DB_USER | base64)
    password: $(echo -n "$1" | base64)
EOF
}
