#!/usr/bin/env bash

set -e -o nounset

. ./env.sh

kubectl delete ns "${APP_NAMESPACE}"
