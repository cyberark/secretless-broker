#!/usr/bin/env bash

. ./env.sh

docker run \
  --name conjur-cli \
  --rm \
  -d \
  -w /work \
  -v $PWD:/work \
  --entrypoint sleep \
  cyberark/conjur-cli:5 \
    infinity

POD_NAME=$(kubectl get pods --namespace "${OSS_CONJUR_NAMESPACE}" \
                                         -l "app=conjur-oss,release=${OSS_CONJUR_RELEASE_NAME}" \
                                         -o jsonpath="{.items[0].metadata.name}")

kubectl exec \
  --namespace "${OSS_CONJUR_NAMESPACE}" \
 "${POD_NAME}" \
  --container=conjur-oss \
  conjurctl wait

kubectl exec \
  --namespace "${OSS_CONJUR_NAMESPACE}" \
 "${POD_NAME}" \
  --container=conjur-oss \
  conjurctl account create "${CONJUR_ACCOUNT}"

CONJUR_ADMIN_API_KEY=`
kubectl exec \
  --namespace "${OSS_CONJUR_NAMESPACE}" \
 "${POD_NAME}" \
  --container=conjur-oss \
  conjurctl role retrieve-key ${CONJUR_ACCOUNT}:user:admin
`

OSS_CONJUR_SERVICE_IP=""
echo "Waiting for end point..."
while [[ -z "${OSS_CONJUR_SERVICE_IP}" ]]; do
  OSS_CONJUR_SERVICE_IP=`
kubectl get svc \
 --namespace "${OSS_CONJUR_NAMESPACE}" \
 "${OSS_CONJUR_HELM_FULLNAME}-ingress" \
 -o jsonpath='{.status.loadBalancer.ingress[].ip}'
` || OSS_CONJUR_SERVICE_IP=''

  # sleep if condition still not met
  [[ -z "${OSS_CONJUR_SERVICE_IP}" ]] && sleep 5
done
echo "End point ready: ${OSS_CONJUR_SERVICE_IP}"

cat << EOL | docker exec -i conjur-cli bash -
echo '${OSS_CONJUR_SERVICE_IP} conjur.myorg.com' >> /etc/hosts

# Here you connect to the endpoint of your Conjur service.
yes yes | conjur init -u https://conjur.myorg.com -a '${CONJUR_ACCOUNT}'

# API key here is the key that creation of the account provided you in step #2
conjur authn login -u admin -p '${CONJUR_ADMIN_API_KEY}'

# Check that you are identified as the admin user
conjur authn whoami
EOL
