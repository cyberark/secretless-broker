export CONFIG_TEMPLATE=CHANGEME             # Examples: mssql, psql, etc.
export TEST_DURATION=1h

export APP_NAME=juxtaposer-${CONFIG_TEMPLATE}

export APP_SERVICE_ACCOUNT=secretless-xa
export AUTHENTICATOR_ID=CHANGEME            # Example: openshift/xa-secretless
export DAP_ACCOUNT=CHANGEME                 # Example: xa
export DAP_NAMESPACE_NAME=CHANGEME          # Example: xa-secretless
export DAP_SSL_CERT_CONFIG_MAP=dap-ssl-cert
export DOCKER_REGISTRY_PATH=CHANGEME
export SECRETLESS_IMAGE=CHANGEME            # Example: cyberark/secretless:1.6.0
export TEST_APP_NAMESPACE_NAME=CHANGEME     # Example: test-xa-namespace

OC_REPOSITORY="docker-registry.default.svc:5000/$TEST_APP_NAMESPACE_NAME"
TAG_NAME="$TEST_APP_NAMESPACE_NAME"
export PERFTOOL_IMAGE="$OC_REPOSITORY/$APP_NAME:$TAG_NAME"
