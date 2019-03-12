#!/usr/bin/env bash

. ./utils.sh

# start application
echo ">>--- Start application"

kubectl --namespace quick-start-application-ns \
 apply \
 -f etc/quick-start-application.yml

wait_for_app quick-start-application quick-start-application-ns
