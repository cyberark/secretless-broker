#!/usr/bin/env bash

. ./utils.sh

# store secretless config
kubectl create configmap quick-start-application-secretless-config \
  --namespace quick-start \
  --from-file=secretless.yml

# build application
docker build -t quick-start-app:latest .

# start application
kubectl apply -f quick-start.yml
wait_for_app quick-start-application
