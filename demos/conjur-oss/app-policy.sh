#!/usr/bin/env bash

. ./env.sh

cat << EOL
---
# Policy enabling the Kubernetes authenticator for your application
- !policy
  id: conjur/authn-k8s/${AUTHENTICATOR_ID}/apps
  body:
    - &hosts
      - !host
        id: ${APP_NAMESPACE}/service_account/${APP_SERVICE_ACCOUNT_NAME}
        annotations:
          kubernetes/authentication-container-name: ${APP_AUTHENTICATION_CONTAINER_NAME}
          kubernetes: "true"
    - !grant
      role: !layer
      members: *hosts

# Grant application's authn identity membership to the application secrets reader layer so authn identity inherits read privileges on application secrets
- !grant
  role: !layer ${APP_SECRETS_READER_LAYER}
  members:
  - !host /conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/${APP_NAMESPACE}/service_account/${APP_SERVICE_ACCOUNT_NAME}

EOL
