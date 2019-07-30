#!/usr/bin/env bash

set -e -o nounset

. ./env.sh

# Delete deployment, service account, ssl certificate and secretless config
kubectl --namespace "${APP_NAMESPACE}" \
  delete \
    deployment/"${APP_NAME}" \
    sa/"${APP_SERVICE_ACCOUNT_NAME}" \
    configmap/conjur-cert \
    configmap/secretless-config

# Create application service account
kubectl \
 --namespace "${APP_NAMESPACE}" \
 create sa "${APP_SERVICE_ACCOUNT_NAME}"

# Store Conjur SSL certificate
kubectl \
  --namespace "${APP_NAMESPACE}" \
  create configmap \
  conjur-cert \
  --from-file=ssl-certificate="tmp/conjur.pem"

# Store secretless.yml configmap
./secretless-config.sh > tmp/secretless.yml
kubectl \
  --namespace "${APP_NAMESPACE}" \
  create configmap \
  secretless-config \
  --from-file=secretless.yml="tmp/secretless.yml"

# Create application deployment
./app-manifest.sh > tmp/app-manifest.yml
kubectl create -f tmp/app-manifest.yml


function app_pod_ready_candidates() {
  # Prints lines with pods
  # that match the criteria of the kubectl command below
  # and have all their containers "ready".
  #
  # The template of the kubectl command below generates
  # lines that take the form
  # <pod_name> <true|false> <true|false> <true|false>
  #
  # , where the n-th <true|false> represents the readiness
  # of the n-th container in the pod
  #
  # The output is piped to grep, which uses extended regexp to match
  # pods that have all their containers ready i.e.
  #
  # <pod_name> true true true true
  #
  local outputTemplate="{{range .items}}{{if not .metadata.deletionTimestamp}}{{.metadata.name}} {{range .status.containerStatuses}}{{.ready}} {{end}}
{{end}}{{end}}"

  kubectl \
    --namespace "${APP_NAMESPACE}" \
    get pods \
    --field-selector=status.phase=Running \
    --selector="app=${APP_NAME}" \
    --output go-template --template="${outputTemplate}" \
    | grep -E "^.+ (true ?)+$"
}

function app_pod_name() {
  app_pod_ready_candidates | head -n 1 | awk '{print $1}'
}

function check_app_ready() {
  echo "Waiting for application ..."
  until app_pod_ready_candidates &> /dev/null; do
    printf "."
    sleep 5
  done

  printf "\n%s\n" "Application ready."
}

# Run test on application
check_app_ready

kubectl \
  --namespace "${APP_NAMESPACE}" \
  exec "$(app_pod_name)" \
   -c app \
   -- \
   mysql -h 0.0.0.0 -P3000 -u x -px -e "status" --ssl-mode disable
