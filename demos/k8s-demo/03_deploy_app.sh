#!/usr/bin/env bash

. ./utils.sh

# store Secretless config
echo ">>--- Create and store Secretless configuration"

kubectl --namespace quick-start \
 create configmap \
 quick-start-application-secretless-config \
 --from-file=etc/secretless.yml

# start application
echo ">>--- Start application"

kubectl --namespace quick-start \
 apply \
 -f etc/quick-start.yml

wait_for_app quick-start-application quick-start
