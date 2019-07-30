#!/usr/bin/env bash

set -e -o nounset

. ./mysql/env.sh
. ./conjur/env.sh

POD_NAME=$(kubectl get pods --namespace "${OSS_CONJUR_NAMESPACE}" \
                                         -l "app=conjur-oss,release=${OSS_CONJUR_RELEASE_NAME}" \
                                         -o jsonpath="{.items[0].metadata.name}")
CONJUR_ADMIN_API_KEY=`
kubectl exec \
  --namespace "${OSS_CONJUR_NAMESPACE}" \
 "${POD_NAME}" \
  --container=conjur-oss \
  conjurctl role retrieve-key ${CONJUR_ACCOUNT}:user:admin
`
CONJUR_ADMIN_AUTHN_LOGIN="admin"

CONJUR_APPLIANCE_URL="https://${OSS_CONJUR_HELM_FULLNAME}.${OSS_CONJUR_NAMESPACE}.svc.cluster.local"


cat << EOL > ../pre-env.sh
# Prerequisites generated from requirements/gen-env
AUTHENTICATOR_ID="${AUTHENTICATOR_ID}"

APP_SECRETS_POLICY_BRANCH="${APP_SECRETS_POLICY_BRANCH}"
APP_SECRETS_READER_LAYER="${APP_SECRETS_READER_LAYER}"

CONJUR_ACCOUNT="${CONJUR_ACCOUNT}"
CONJUR_APPLIANCE_URL="${CONJUR_APPLIANCE_URL}"
CONJUR_ADMIN_AUTHN_LOGIN="${CONJUR_ADMIN_AUTHN_LOGIN}"
CONJUR_ADMIN_API_KEY="${CONJUR_ADMIN_API_KEY}"

OSS_CONJUR_SERVICE_ACCOUNT_NAME="${OSS_CONJUR_SERVICE_ACCOUNT_NAME}"
OSS_CONJUR_NAMESPACE="${OSS_CONJUR_NAMESPACE}"
EOL