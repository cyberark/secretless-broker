#!/bin/bash -e

step() {
  echo
  echo ">>--- $1"
  echo
}

first_pod() {
  kubectl get pods \
    --namespace "$2" \
    --selector=app="$1" \
    --output=jsonpath='{$.items[0].metadata.name}'
}

wait_for_app() {
  local waiting=false

  until [[ "$(kubectl get pods \
    --namespace "$2" \
    --selector app="$1" \
    --output jsonpath='{$.items[0].status.containerStatuses.*.ready}')" =~ (true ?)+ ]]
  do
    if [[ "$waiting" != "true" ]]; then
      echo "Waiting for $1 to be ready"
      waiting=true
    fi
    echo -n "."
    sleep 3
  done

  if [[ "$waiting" = "true" ]]; then
    echo "OK"
  fi
}
