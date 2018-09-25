#!/bin/bash

set -e

usage() {
    cat <<EOF >&2
Generate MutatingWebhookConfiguration for CyberArk sidecar injector webhook service.

This script uses generates a MutatingWebhookConfiguration using the provided service name of the webhook and the namespace where the webhook service resides.

usage: ${0} [OPTIONS]

The following flags are required.

       --service          Service name of webhook.
       --namespace        Namespace where webhook service resides.
EOF
    exit 1
}

while [[ $# -gt 0 ]]; do
    case ${1} in
        --service)
            service="$2"
            shift
            ;;
        --namespace)
            namespace="$2"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done

if [ -z ${service} ] || [ -z ${namespace} ]
then
    usage
fi

ROOT=$(cd $(dirname $0)/../../; pwd)

set -o errexit
set -o nounset
set -o pipefail

export CA_BUNDLE=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n')

if command -v envsubst >/dev/null 2>&1; then
    envsubst
else
    sed \
        -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g" \
        -e "s|\${namespace}|${namespace}|g" \
        -e "s|\${service}|${service}|g"
fi
