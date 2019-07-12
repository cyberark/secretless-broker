#!/usr/bin/env bash

. ./env.sh

helm install \
  --name "${OSS_CONJUR_RELEASE_NAME}"\
  --set dataKey="$(docker run --rm cyberark/conjur data-key generate)" \
  --set serviceAccount.name="${OSS_CONJUR_SERVICE_ACCOUNT_NAME}" \
  --set account="${CONJUR_ACCOUNT}" \
  --set authenticators="authn-k8s/${AUTHENTICATOR_ID}\,authn" \
  https://github.com/cyberark/conjur-oss-helm-chart/releases/download/v1.3.7/conjur-oss-1.3.7.tgz \
    --namespace "${OSS_CONJUR_NAMESPACE}"
