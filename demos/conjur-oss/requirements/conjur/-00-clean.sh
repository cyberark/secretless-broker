#!/usr/bin/env bash

. ./env.sh

helm delete --purge "${OSS_CONJUR_RELEASE_NAME}"
kubectl delete ns "${OSS_CONJUR_NAMESPACE}"
docker rm -f "conjur-cli-${OSS_CONJUR_NAMESPACE}"
