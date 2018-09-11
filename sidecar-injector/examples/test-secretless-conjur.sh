# proceed after 6_deploy_test_app.sh (because this needs the backend) of the kubernetes-conjur-demo

cd $(dirname $0)/..
. .env

# derived values
CONJUR_APPLIANCE_URL="https://conjur-follower.$CONJUR_NAMESPACE_NAME.svc.cluster.local/api"
CONJUR_AUTHN_URL="https://conjur-follower.$CONJUR_NAMESPACE_NAME.svc.cluster.local/api/authn-k8s/$AUTHENTICATOR_ID"
CONJUR_AUTHN_LOGIN="$TEST_APP_NAMESPACE_NAME/service_account/test-app-summon-init"
CONJUR_SSL_CERTIFICATE=$(follower_pod_name=$(kubectl -n $CONJUR_NAMESPACE_NAME get pods -l role=follower --no-headers | awk '{ print $1 }' | head -1); kubectl exec -n $CONJUR_NAMESPACE_NAME $follower_pod_name -- cat /opt/conjur/etc/ssl/conjur.pem)

if [ $CONJUR_VERSION = '4' ]; then
  CONJUR_AUTHN_LOGIN=$TEST_APP_NAMESPACE_NAME/service_account/$TEST_APP_SERVICE_ACCOUNT
else
  CONJUR_AUTHN_LOGIN=host/conjur/authn-k8s/$AUTHENTICATOR_ID/apps/$TEST_APP_NAMESPACE_NAME/service_account/$TEST_APP_SERVICE_ACCOUNT
fi

cat << EOL > test-secretless-conjur.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: "secretless-conjur-config"
data:
  CONJUR_ACCOUNT: "$CONJUR_ACCOUNT"
  CONJUR_VERSION: "$CONJUR_VERSION"
  CONJUR_APPLIANCE_URL: "$CONJUR_APPLIANCE_URL"
  CONJUR_AUTHN_URL: "$CONJUR_AUTHN_URL"
  CONJUR_SSL_CERTIFICATE: |
$(echo "$CONJUR_SSL_CERTIFICATE" | awk '{ print "    " $0 }')
  CONJUR_AUTHN_LOGIN: "$CONJUR_AUTHN_LOGIN"
  secretless.yml: |
    listeners:
      - name: test-app-pg-listener
        protocol: pg
        address: 0.0.0.0:5432

    handlers:
      - name: test-app-pg-handler
        listener: test-app-pg-listener
        credentials:
          - name: address
            provider: conjur
            id: test-secretless-app-db/url
          - name: username
            provider: conjur
            id: test-secretless-app-db/username
          - name: password
            provider: conjur
            id: test-secretless-app-db/password

---
apiVersion: v1
kind: Pod
metadata:
  name: "example-usage-secretless-conjur"
  annotations:
    sidecar-injector.cyberark.com/inject: "yes"
    sidecar-injector.cyberark.com/conjurConnConfig: "secretless-conjur-config"
    sidecar-injector.cyberark.com/conjurAuthConfig: "secretless-conjur-config"
    sidecar-injector.cyberark.com/secretlessConfig: "secretless-conjur-config"
    sidecar-injector.cyberark.com/injectType: "secretless"
  labels:
    app: example-usage-secretless-conjur
spec:
  serviceAccountName: $TEST_APP_SERVICE_ACCOUNT
  containers:
  - name: app
    env:
      - name: DB_URL
        value: postgresql://localhost:5432/postgres
    image: postgres:9.6-alpine
    command: ["sleep", "100000"]
EOL

cat << EOL
Run the following to test:

kubectl label namespace $TEST_APP_NAMESPACE_NAME cyberark-sidecar-injector=enabled
kubectl create sa -n $TEST_APP_NAMESPACE_NAME $TEST_APP_SERVICE_ACCOUNT
kubectl -n $TEST_APP_NAMESPACE_NAME delete --force --grace-period 0 -f test-secretless-conjur.yaml; kubectl -n $TEST_APP_NAMESPACE_NAME create -f test-secretless-conjur.yaml -o yaml
kubectl -n $TEST_APP_NAMESPACE_NAME exec -it example-usage-secretless-conjur -c app -- sh -c 'psql \$DB_URL?sslmode=disable -c "\dt"'
EOL
