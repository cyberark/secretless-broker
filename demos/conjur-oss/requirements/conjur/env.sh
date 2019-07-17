# configurable env vars
AUTHENTICATOR_ID=example

APP_SECRETS_POLICY_BRANCH=apps/secrets/test
APP_SECRETS_READER_LAYER=apps/layers/myapp

CONJUR_ACCOUNT=example_acc

OSS_CONJUR_SERVICE_ACCOUNT_NAME=conjur-sa
OSS_CONJUR_NAMESPACE=kumbi-conjur
OSS_CONJUR_RELEASE_NAME=sealing-whale

OSS_CONJUR_HELM_FULLNAME=$(echo "${OSS_CONJUR_RELEASE_NAME}-conjur-oss" |  cut -c 1-63 | sed -e "s/\--*$//"); # because DNS, see helm chart
