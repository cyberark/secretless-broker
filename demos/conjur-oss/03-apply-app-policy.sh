#!/usr/bin/env bash

. ./env.sh

cat << EOL | docker exec -i conjur-cli bash -
# Apply App policy
./app-policy.sh > tmp/app-policy.yml
conjur policy load root tmp/app-policy.yml

# Init App variables
conjur variable values add "${APP_NAME}-db/password" abcxyzpassword
conjur variable values add "${APP_NAME}-db/url" abcxyzurl
conjur variable values add "${APP_NAME}-db/username" abcxyzusername
EOL
