#!/bin/bash -e

start_step() { printf '\n\n>>--- %s\n\n' "$1"; }

first_pod() {
  kubectl get pods \
    --namespace "$2" \
    --selector=app="$1" \
    --output=jsonpath='{$.items[0].metadata.name}'
}

wait_for_app() {
  local waiting=false

  while [[ "$(kubectl get pods \
    --namespace "$2" \
    --selector app="$1" \
    --output jsonpath='{$.items[0].status.containerStatuses.*.ready}')" != *true* ]]
  do
    if [[ "$waiting" != "true" ]]; then
      printf "Waiting for %s to be ready" "$1"
      waiting=true
    fi
    printf "."
    sleep 3
  done

  if [[ "$waiting" = "true" ]]; then
    printf "Done"
  fi
}

# Note: In future versions of k8s we'll be able to replace this function
# with the k8s "wait" command:
#
# https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#wait
#
delete_ns_and_cleanup() {
  # We don't care about the output of this command, only its return return code:
  # success means there was something to delete, and that we must wait for
  # deletion to fully process.  Failure means there was nothing to delete and
  # we can return early.
  if ! kubectl delete namespace "$1" > /dev/null 2>&1; then
    return 0
  fi

  printf "Cleaning up old namespace"

  # As long as we can still "get" the namespace, the deletion isn't done.
  while kubectl get namespace "$1" > /dev/null 2>&1; do 
    printf "."
    sleep 3
  done

  printf '%s\n\n' "Done"
}
