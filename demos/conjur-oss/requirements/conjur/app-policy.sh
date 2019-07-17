#!/usr/bin/env bash

. ./env.sh

cat << EOL
---

# Policy containing your application secrets reader layer
- !policy
  id: "${APP_SECRETS_READER_LAYER}"
  annotations:
    description: This policy houses the layer with privileges to read app secrets
  body:
  - !layer

# Policy containing your application secrets
- !policy
  id: "${APP_SECRETS_POLICY_BRANCH}"
  annotations:
    description: This policy contains the creds required by the app

  body:
    - &init-variables
      - !variable password
      - !variable host
      - !variable port
      - !variable username

    - !permit
      role: !layer "/${APP_SECRETS_READER_LAYER}"
      privileges: [ read, execute ]
      resources: *init-variables
EOL
