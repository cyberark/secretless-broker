#!/usr/bin/env bash

. ./env.sh

cat << EOL | docker exec -i conjur-cli bash -
# Apply Conjur policy
./conjur-policy.sh > tmp/conjur-policy.yml
conjur policy load root tmp/conjur-policy.yml

# Init authn-k8s stuff

## Generate OpenSSL private key
openssl genrsa -out tmp/ca.key 2048

## Generate root CA certificate
openssl req -x509 -new -nodes -key tmp/ca.key -sha1 -days 3650 -set_serial 0x0 -out tmp/ca.crt \
  -subj '/CN=conjur.authn-k8s.${AUTHENTICATOR_ID}/OU=Conjur Kubernetes CA/O=${CONJUR_ACCOUNT}' \
  -config openssl-config

## Load variable values
cat tmp/ca.key | conjur variable values add conjur/authn-k8s/${AUTHENTICATOR_ID}/ca/key
cat tmp/ca.crt | conjur variable values add conjur/authn-k8s/${AUTHENTICATOR_ID}/ca/cert
EOL
