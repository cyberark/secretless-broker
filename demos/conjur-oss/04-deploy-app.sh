#!/usr/bin/env bash

. ./env.sh

kubectl delete --wait=true ns "${APP_NAMESPACE}"
kubectl create ns "${APP_NAMESPACE}"

kubectl \
 --namespace "${APP_NAMESPACE}" \
 create sa "${APP_SERVICE_ACCOUNT_NAME}"

docker exec conjur-cli bash -c "cat /root/*.pem" > tmp/conjur.pem
kubectl \
  --namespace "${APP_NAMESPACE}" \
  create configmap \
  conjur-cert \
  --from-file=ssl-certificate="tmp/conjur.pem"

./conjur-authenticator-role.sh > tmp/conjur-authenticator-role.yml
kubectl create -f tmp/conjur-authenticator-role.yml

./secretless-config.sh > tmp/secretless.yml
kubectl \
  --namespace "${APP_NAMESPACE}" \
  create configmap \
  secretless-config \
  --from-file=secretless.yml="tmp/secretless.yml"

./app-manifest.sh > tmp/app-manifest.yml
kubectl create -f tmp/app-manifest.yml

POD_NAME=$(kubectl \
  --namespace "${APP_NAMESPACE}" \
  get pods \
  -l "app=${APP_NAME}" \
  -o jsonpath="{.items[0].metadata.name}")

kubectl \
  --namespace "${APP_NAMESPACE}" \
  exec "${POD_NAME}" \
   -c app \
   -- \
   curl --silent localhost:8080 \
    | grep authorization \
    | sed -e "s/^authorization=Basic\ //" \
    | base64 --decode; echo
