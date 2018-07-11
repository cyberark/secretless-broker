#!/usr/bin/env bash

. ./config.sh
. ./utils.sh

# store Secretless config
echo ">>--- Create and store Secretless configuratin"

kubectl create configmap quick-start-application-secretless-config \
  --namespace quick-start \
  --from-file=secretless.yml

# build application
echo ">>--- Build application"

docker build -t quick-start-app:latest .

# start application
echo ">>--- Start application"

kubectl apply -f quick-start.yml
wait_for_app quick-start-application quick-start
