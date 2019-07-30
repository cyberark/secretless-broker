#!/usr/bin/env bash

set -e -o nounset

. ./env.sh

cat << EOL
---
# Initializes users
## bob - devops admin

- !group devops
- !group ops

## kube_admin and devops groups are members of the ops admin group
- !grant
  role: !group ops
  members:
  - !group devops

## TODO: bob's credentials can be used to apply the policy instead of admin
- !user bob
- !grant
  role: !group devops
  member: !user bob

# This policy defines an authn-k8s endpoint, CA creds and a layer for whitelisted identities permitted to authenticate to it
- !policy
  id: conjur/authn-k8s/${AUTHENTICATOR_ID}
  owner: !group devops
  annotations:
    description: Namespace defs for the Conjur cluster in dev
  body:
  - !webservice
    annotations:
      description: authn service for cluster

  ## Permit a layer of whitelisted authn ids to call authn service
  - !permit
    resource: !webservice
    privilege: [ read, authenticate ]
    role: !layer apps

  ## CA cert and key for creating client certificates
  - !policy
    id: ca
    body:
    - !variable
      id: cert
      annotations:
        description: CA cert for Kubernetes Pods.
    - !variable
      id: key
      annotations:
        description: CA key for Kubernetes Pods.

  # This policy defines a layer of whitelisted identities permitted to authenticate to the authn-k8s endpoint.
  - !policy
    id: apps
    owner: !group /devops
    annotations:
      description: Identities permitted to authenticate
    body:
    - !layer
      annotations:
        description: Layer of authenticator identities permitted to call authn svc

EOL
