#!/usr/bin/env bash

username=$1
password=$2

: "${username:?"Need to provide non-empty username as first argument"}"
: "${password:?"Need to provide non-empty password as second argument"}"

cat << EOL
apiVersion: "secretless${SECRETLESS_CRD_SUFFIX}.io/v1"
kind: "Configuration"
metadata:
  name: first
spec:
  listeners:
    - name: http_config_1_listener
      protocol: http
      address: localhost:8000

  handlers:
    - name: http_config_1_handler
      type: basic_auth
      listener: http_config_1_listener
      # The echo-server upstream is plaintext http, so opt out of the
      # forceSSL default (true) to avoid a TLS handshake against it.
      forceSSL: false
      match:
        - ^http.*
      credentials:
        - name: username
          provider: literal
          id: "${username}"
        - name: password
          provider: literal
          id: "${password}"
EOL
