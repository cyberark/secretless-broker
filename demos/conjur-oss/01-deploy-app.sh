#!/usr/bin/env bash

. ./env.sh

# delete deployment, service account, ssl certificate and secretless config
kubectl --namespace "${APP_NAMESPACE}" \
  delete \
    deployment/"${APP_NAME}" \
    sa/"${APP_SERVICE_ACCOUNT_NAME}" \
    configmap/conjur-cert \
    configmap/secretless-config

# create application service account
kubectl \
 --namespace "${APP_NAMESPACE}" \
 create sa "${APP_SERVICE_ACCOUNT_NAME}"

# store conjur SSL certificate
kubectl \
  --namespace "${APP_NAMESPACE}" \
  create configmap \
  conjur-cert \
  --from-file=ssl-certificate="tmp/conjur.pem"

# store secretless.yml configmap
./secretless-config.sh > tmp/secretless.yml
kubectl \
  --namespace "${APP_NAMESPACE}" \
  create configmap \
  secretless-config \
  --from-file=secretless.yml="tmp/secretless.yml"

# create application deployment
./app-manifest.sh > tmp/app-manifest.yml
kubectl create -f tmp/app-manifest.yml

function app_pod_ready_candidates() {
  kubectl \
    --namespace "${APP_NAMESPACE}" \
    get pods \
    --field-selector=status.phase=Running \
    --selector="app=${APP_NAME}" \
    --output go-template --template='{{range .items}}{{if not .metadata.deletionTimestamp}}{{.metadata.name}} {{range .status.containerStatuses}}{{.ready}} {{end}}
{{end}}{{end}}' \
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

# run test on application
check_app_ready

kubectl \
  --namespace "${APP_NAMESPACE}" \
  exec "$(app_pod_name)" \
   -c app \
   -- \
   mysql -h 0.0.0.0 -P3000 -u x -px -e "status" --ssl-mode disable
