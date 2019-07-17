#!/usr/bin/env bash

# configurable env vars
MYSQL_NAMESPACE="kt-mysql"
MYSQL_RELEASE="kt-mysql-release"
MYSQL_USERNAME="root"
MYSQL_PASSWORD="mysqlRootPasswordx"
MYSQL_HOST="${MYSQL_RELEASE}.${MYSQL_NAMESPACE}.svc.cluster.local"
MYSQL_PORT="3306"
