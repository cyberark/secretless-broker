#!/usr/bin/env bash

. ./env.sh

helm install \
  --namespace "${MYSQL_NAMESPACE}" \
  --name "${MYSQL_RELEASE}" \
  --set mysqlRootPassword="${MYSQL_PASSWORD}" \
  stable/mysql
