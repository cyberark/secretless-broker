#!/usr/bin/env bash

. ./env.sh

kubectl delete ns "${APP_NAMESPACE}"
