export APP_NAME=juxtaposer
export APP_SERVICE_ACCOUNT=secretless-xa
export AUTHENTICATOR_ID=openshift/secretless-xa
export CONFIG_TEMPLATE=pg
export DAP_ACCOUNT=xa
export DAP_NAMESPACE_NAME=xa-secretless
export DAP_SSL_CERT_CONFIG_MAP=dap-ssl-cert
export DOCKER_REGISTRY_PATH=REPLACEME
export TEST_APP_NAMESPACE_NAME=srdjan-secretless-xa

OC_REPOSITORY="docker-registry.default.svc:5000/$TEST_APP_NAMESPACE_NAME"
TAG_NAME="$TEST_APP_NAMESPACE_NAME"
export PERFTOOL_IMAGE="$OC_REPOSITORY/$APP_NAME:$TAG_NAME"
