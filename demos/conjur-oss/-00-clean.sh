#!/usr/bin/env bash

. ./env.sh

kubectl delete ns "${APP_NAMESPACE}"
helm delete --purge "${OSS_CONJUR_RELEASE_NAME}"
kubectl delete ns "${OSS_CONJUR_NAMESPACE}"
docker rm -f conjur-cli
