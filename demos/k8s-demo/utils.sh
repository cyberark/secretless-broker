#!/bin/bash -e

get_first_pod_for_app() {
    echo $(kubectl get --namespace "$2" po -l=app="$1" -o=jsonpath='{$.items[0].metadata.name}')
}

wait_for_app() {
    while [[ ! $(kubectl get --namespace "$2" po -l=app="$1" -o=jsonpath='{$.items[0].status.containerStatuses.*.ready}' | grep -v "false") ]] ; do
        echo "Waiting for $1 to be ready"
        sleep 5
    done
    echo Ready!
}
