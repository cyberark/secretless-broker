#!/usr/bin/env bash

# configurable env vars
AUTHENTICATOR_ID=example

APP_NAME=my-app
APP_NAMESPACE=kumbi-app
APP_AUTHENTICATION_CONTAINER_NAME=secretless
APP_SERVICE_ACCOUNT_NAME=my-app-sa

CONJUR_ACCOUNT=example_acc

OSS_CONJUR_SERVICE_ACCOUNT_NAME=conjur-sa
OSS_CONJUR_NAMESPACE=kumbi-conjur
OSS_CONJUR_RELEASE_NAME=sealing-whale

# generated env vars

OSS_CONJUR_HELM_FULLNAME=$(echo "${OSS_CONJUR_RELEASE_NAME}-conjur-oss" |  cut -c 1-63 | sed -e "s/\--*$//"); # because DNS, see helm chart
CONJUR_APPLIANCE_URL="https://${OSS_CONJUR_HELM_FULLNAME}.${OSS_CONJUR_NAMESPACE}.svc.cluster.local"
CONJUR_AUTHN_URL="${CONJUR_APPLIANCE_URL}/authn-k8s/${AUTHENTICATOR_ID}"
CONJUR_AUTHN_LOGIN="host/conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/${APP_NAMESPACE}/service_account/${APP_SERVICE_ACCOUNT_NAME}"
CONJUR_VERSION=5

mkdir -p tmp
