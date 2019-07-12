#!/usr/bin/env bash

. ./env.sh

cat << EOL
---
- !policy
  id: "${APP_NAME}"
  owner: !group devops
  annotations:
    description: This policy connects authn identities to an application identity. It defines a layer named for an application that contains the whitelisted identities that can authenticate to the authn-k8s endpoint. Any permissions granted to the application layer will be inherited by the whitelisted authn identities, thereby granting access to the authenticated identity.
  body:
  - !layer

 # add authn identities to application layer so authn roles inherit app's permissions
  - !grant
    role: !layer
    members:
    - !layer /conjur/authn-k8s/${AUTHENTICATOR_ID}/apps
- !policy
  id: "${APP_NAME}-db"
  owner: !group devops
  annotations:
    description: This policy contains the creds to access the summon init app DB

  body:
    - &init-variables
      - !variable password
      - !variable url
      - !variable username

    - !permit
      role: !layer "/${APP_NAME}"
      privileges: [ read, execute ]
      resources: *init-variables
EOL
