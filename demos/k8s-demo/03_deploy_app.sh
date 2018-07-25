#!/usr/bin/env bash

. ./utils.sh

readonly REGISTRY_URL="registry2.itci.conjur.net"

# log into private registry (TODO - remove once images are public!!)
echo ">>--- Logging into private Docker registry"

kubectl get secret --namespace quick-start "${REGISTRY_URL}" &>/dev/null || {
  if [[ -z "${REGISTRY_USERNAME}" || -z "${REGISTRY_PASSWORD}" ]]; then
    echo -e "\nLogin to private registry [${REGISTRY_URL}]"
    printf "Username: "
    read REGISTRY_USERNAME
    printf "Password: "
    read -s REGISTRY_PASSWORD
  fi

  kubectl create secret docker-registry "${REGISTRY_URL}" \
    --namespace quick-start \
    --docker-server="${REGISTRY_URL}" \
    --docker-username="${REGISTRY_USERNAME}" \
    --docker-password="${REGISTRY_PASSWORD}" \
    --docker-email=nil &>/dev/null

  echo ""
  echo ""
}

# store Secretless config
echo ">>--- Create and store Secretless configuration"

kubectl create configmap quick-start-application-secretless-config \
  --namespace quick-start \
  --from-file=etc/secretless.yml

# start application
echo ">>--- Start application"

kubectl apply -f etc/quick-start.yml
wait_for_app quick-start-application quick-start
