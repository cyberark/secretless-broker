#!/usr/bin/env bash

. ./env.sh

cat << EOL
version: "2"
services:
  app_db:
    protocol: mysql
    listenOn: tcp://0.0.0.0:3000
    credentials:
      host:
        from: conjur
        get: ${APP_SECRETS_POLICY_BRANCH}/host
      port:
        from: conjur
        get: ${APP_SECRETS_POLICY_BRANCH}/port
      username:
        from: conjur
        get: ${APP_SECRETS_POLICY_BRANCH}/username
      password:
        from: conjur
        get: ${APP_SECRETS_POLICY_BRANCH}/password
      sslmode: disable
EOL
