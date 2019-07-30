#!/usr/bin/env bash

set -e -o nounset

# Configurable env vars
MYSQL_NAMESPACE="mysql"
MYSQL_RELEASE="mysql-release"
MYSQL_USERNAME="root"
MYSQL_PASSWORD="mysqlRootPasswordx"
MYSQL_HOST="${MYSQL_RELEASE}.${MYSQL_NAMESPACE}.svc.cluster.local"
MYSQL_PORT="3306"
