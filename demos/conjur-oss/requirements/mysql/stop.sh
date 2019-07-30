#!/usr/bin/env bash

set -e -o nounset

. ./env.sh

helm delete --purge "${MYSQL_RELEASE}"
kubectl delete ns "${MYSQL_NAMESPACE}"
