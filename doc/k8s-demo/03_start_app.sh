#!/usr/bin/env bash

. ./config.sh
. ./utils.sh

# store Secretless config
echo ">>--- Create and store Secretless configuratin"

kubectl create configmap quick-start-application-secretless-config \
  --namespace quick-start \
  --from-file=etc/secretless.yml

# start application
echo ">>--- Start application"

kubectl apply -f etc/quick-start.yml
wait_for_app quick-start-application quick-start
