#!/usr/bin/env bash

. ./env.sh

helm delete --purge "${MYSQL_RELEASE}"
kubectl delete ns "${MYSQL_NAMESPACE}"
