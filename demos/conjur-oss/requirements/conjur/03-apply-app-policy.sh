#!/usr/bin/env bash

. ./env.sh
. ../mysql/env.sh

cat << EOL | docker exec -i "conjur-cli-${OSS_CONJUR_NAMESPACE}" bash -
mkdir -p tmp

# Apply App policy
./app-policy.sh > tmp/app-policy.yml
conjur policy load root tmp/app-policy.yml

conjur variable values add "${APP_SECRETS_POLICY_BRANCH}/host" "${MYSQL_HOST}"
conjur variable values add "${APP_SECRETS_POLICY_BRANCH}/password" "${MYSQL_PASSWORD}"
conjur variable values add "${APP_SECRETS_POLICY_BRANCH}/username" "${MYSQL_USERNAME}"
conjur variable values add "${APP_SECRETS_POLICY_BRANCH}/port" "${MYSQL_PORT}"
EOL
