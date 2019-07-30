#!/usr/bin/env bash

set -e -o nounset

. ./env.sh

function cleanup() {
    kubectl --namespace "${APP_NAMESPACE}" delete pod conjur-cli
}
trap cleanup EXIT INT TERM

# Ensure conjur authenticator is able to operate in application namespace
# via kubernetes RBAC
./conjur-authenticator-role.sh > tmp/conjur-authenticator-role.yml
kubectl create -f tmp/conjur-authenticator-role.yml

# Create application namespace
kubectl create ns "${APP_NAMESPACE}"

# Create CLI container
kubectl \
  --namespace "${APP_NAMESPACE}" \
  run conjur-cli \
   --image=cyberark/conjur-cli:5 \
   --restart=Never \
   --attach=false \
   --command -- sleep infinity

cat << EOL | kubectl --namespace "${APP_NAMESPACE}" exec -i conjur-cli -- bash -
# Here you connect to the endpoint of your Conjur service.
yes yes | conjur init -u '${CONJUR_APPLIANCE_URL}' -a '${CONJUR_ACCOUNT}'

# API key here is the key that creation of the account provided you in step #2
conjur authn login -u '${CONJUR_ADMIN_AUTHN_LOGIN}' -p '${CONJUR_ADMIN_API_KEY}'

# Check that you are identified as the admin user
conjur authn whoami

EOL

# Retrieve Conjur SSL certificate
kubectl --namespace "${APP_NAMESPACE}" exec -i conjur-cli -- bash -c "cat /root/*.pem" > tmp/conjur.pem

# Generate application policy
./app-policy.sh > tmp/app-policy.yml

# Apply application policy
kubectl --namespace "${APP_NAMESPACE}" exec -i conjur-cli -- rm -rf /tmp
kubectl --namespace "${APP_NAMESPACE}" cp $PWD/tmp conjur-cli:/tmp
kubectl --namespace "${APP_NAMESPACE}" exec -i conjur-cli -- conjur policy load root /tmp/app-policy.yml
