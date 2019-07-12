#!/usr/bin/env bash

. ./env.sh

cat << EOL
version: "2"
services:
  http_basic_auth:
    protocol: http
    listenOn: tcp://0.0.0.0:3000
    credentials:
      username:
        from: conjur
        get: ${APP_NAME}-db/username
      password:
        from: conjur
        get: ${APP_NAME}-db/password
    config:
      authenticationStrategy: basic_auth
      authenticateURLsMatching:
        - .*
EOL
