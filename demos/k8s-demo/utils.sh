#!/bin/bash -e

first_pod() {
  kubectl get pods \
    --namespace "$2" \
    --selector=app="$1" \
    --output=jsonpath='{$.items[0].metadata.name}'
}

wait_for_app() {
  while kubectl get pods \
    --namespace "$2" \
    --selector=app="$1" \
    --output=jsonpath='{$.items[0].status.containerStatuses.*.ready}' \
      | grep -q "false"
  do
    echo "Waiting for $1 to be ready"
    sleep 5
  done
  echo "$1" Ready!
}

# Usage: repeat_str 3 hi (returns hihihi)
repeat_str() {
  local i
  for ((i=0; i<"$1"; i++)); do
    printf "%s" "$2"
  done
}

# repeats a cmd
# Usage: repeat 3 echo hi
# hi
# hi
# hi
repeat() { local i n; n=$1; shift; for ((i=1; i<=n; i++)); do "$@"; done; }
