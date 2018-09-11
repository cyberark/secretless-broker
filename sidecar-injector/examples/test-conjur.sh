# proceed after 4_store_conjur_cert.sh of the kubernetes-conjur-demo

CONJUR_NAMESPACE_NAME=conjur-ktanekha
CONJUR_ACCOUNT=my-account
CONJUR_VERSION=4
AUTHENTICATOR_ID=gke-test
TEST_APP_NAMESPACE_NAME=test-app-ktanekha
TEST_APP_SERVICE_ACCOUNT=test-app-summon-init

# derived values
CONJUR_APPLIANCE_URL="https://conjur-follower.$CONJUR_NAMESPACE_NAME.svc.cluster.local/api"
CONJUR_AUTHN_URL="https://conjur-follower.$CONJUR_NAMESPACE_NAME.svc.cluster.local/api/authn-k8s/$AUTHENTICATOR_ID"
CONJUR_AUTHN_LOGIN="$TEST_APP_NAMESPACE_NAME/service_account/test-app-secretless"
CONJUR_SSL_CERTIFICATE=$(follower_pod_name=$(kubectl -n $CONJUR_NAMESPACE_NAME get pods -l role=follower --no-headers | awk '{ print $1 }' | head -1); kubectl exec -n $CONJUR_NAMESPACE_NAME $follower_pod_name -- cat /opt/conjur/etc/ssl/conjur.pem)

if [ $CONJUR_VERSION = '4' ]; then
  CONJUR_AUTHN_LOGIN=$TEST_APP_NAMESPACE_NAME/service_account/$TEST_APP_SERVICE_ACCOUNT
else
  CONJUR_AUTHN_LOGIN=host/conjur/authn-k8s/$AUTHENTICATOR_ID/apps/$TEST_APP_NAMESPACE_NAME/service_account/$TEST_APP_SERVICE_ACCOUNT
fi

cat << EOL > test-conjur.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: "authenticator-config"
data:
  CONJUR_ACCOUNT: "$CONJUR_ACCOUNT"
  CONJUR_VERSION: "$CONJUR_VERSION"
  CONJUR_APPLIANCE_URL: "$CONJUR_APPLIANCE_URL"
  CONJUR_AUTHN_URL: "$CONJUR_AUTHN_URL"
  CONJUR_SSL_CERTIFICATE: |
$(echo "$CONJUR_SSL_CERTIFICATE" | awk '{ print "    " $0 }')
  CONJUR_AUTHN_LOGIN: "$CONJUR_AUTHN_LOGIN"
---
apiVersion: v1
kind: Pod
metadata:
  name: "example-usage"
  annotations:
    sidecar-injector.cyberark.com/inject: "yes"
    sidecar-injector.cyberark.com/config: "authenticator-config"
    sidecar-injector.cyberark.com/injectType: "authenticator"
  labels:
    app: example-usage
spec:
  serviceAccountName: test-app-secretless
  containers:
  - name: app
    env:
      - name: http_proxy
        value: "http://0.0.0.0:3000"
    image: googlecontainer/echoserver:1.1
    volumeMounts:
      - mountPath: /run/conjur
        name: conjur-access-token
EOL

cat << EOL
Run the following to test:

kubectl label namespace $TEST_APP_NAMESPACE_NAME cyberark-sidecar-injector=enabled
kubectl create sa -n $TEST_APP_NAMESPACE_NAME $TEST_APP_SERVICE_ACCOUNT
kubectl -n $TEST_APP_NAMESPACE_NAME delete --force --grace-period 0 -f test-conjur.yaml; kubectl -n $TEST_APP_NAMESPACE_NAME create -f test-conjur.yaml -o yaml
kubectl -n $TEST_APP_NAMESPACE_NAME exec -i example-usage -c app -- cat /run/conjur/access-token
EOL
