# proceed after 4_store_conjur_cert.sh of the kubernetes-conjur-demo

cd $(dirname $0)/..
. .env

cat << EOL > test-secretless.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: "secretless-config"
data:
  secretless.yml: |
    listeners:
      - name: http_good_basic_auth
        debug: true
        protocol: http
        address: 0.0.0.0:3000

    handlers:
      - name: http_good_basic_auth_handler
        type: basic_auth
        listener: http_good_basic_auth
        debug: true
        match:
          - ^http.*
        credentials:
          - name: username
            provider: literal
            id: not-so-secret-username
          - name: password
            provider: literal
            id: should-be-secret-password

---
apiVersion: v1
kind: Pod
metadata:
  name: "example-usage-secretless"
  annotations:
    sidecar-injector.cyberark.com/inject: "yes"
    sidecar-injector.cyberark.com/secretlessConfig: "secretless-config"
    sidecar-injector.cyberark.com/injectType: "secretless"
    sidecar-injector.cyberark.com/containerName: "yes-container"
  labels:
    app: example-usage-secretless
spec:
  containers:
  - name: app
    env:
      - name: http_proxy
        value: "http://0.0.0.0:3000"
    image: googlecontainer/echoserver:1.1

EOL

cat << EOL
Run the following to test:

kubectl label namespace $TEST_APP_NAMESPACE_NAME cyberark-sidecar-injector=enabled
kubectl -n $TEST_APP_NAMESPACE_NAME delete --force --grace-period 0 -f test-secretless.yaml; kubectl -n $TEST_APP_NAMESPACE_NAME create -f test-secretless.yaml -o yaml
kubectl -n $TEST_APP_NAMESPACE_NAME exec -it example-usage-secretless -c app -- curl -i localhost:8080

EOL
