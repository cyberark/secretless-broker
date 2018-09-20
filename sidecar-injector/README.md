# Secretless Broker Sidecar Injector MutatingAdmissionWebhook

This document shows how to build and deploy the Secretless Broker Sidecar Injector [MutatingAdmissionWebhook](https://kubernetes.io/docs/admin/admission-controllers/#mutatingadmissionwebhook-beta-in-19) which injects a Secretless-broker sidecar container into a pod prior to persistence of the underlying object.

## Prerequisites

Kubernetes 1.9.0 or above with the `admissionregistration.k8s.io/v1beta1` API enabled. Verify that by the following command:
```
~$ kubectl api-versions | grep admissionregistration.k8s.io/v1beta1
```
The result should be:
```
admissionregistration.k8s.io/v1beta1
```

In addition, the `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` admission controllers should be added and listed in the correct order in the admission-control flag of kube-apiserver. Please see the [Kubernetes documentation]( https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/). It is likely that this is set by default if your cluster is running on GKE.

If using `minikube`, start your cluster as follows:
```bash
~$ minikube start --memory=8192 --kubernetes-version=v1.10.0
```

## Docker image

The docker image for the mutating webhook admission controller is publicly available on Dockerhub as [cyberark/sidecar-injector](https://hub.docker.com/r/cyberark/sidecar-injector/) .

## Installing the Sidecar Injector (Manually)

### Dedicated Namespace

Create a namespace "cyberark-sidecar-injector", where you will deploy the CyberArk Sidecar Injector Webhook components.

1. Create namespace
    ```bash
    ~$ kubectl create namespace cyberark-sidecar-injector
    ```

### Deploy Sidecar Injector

1. Create a signed cert/key pair and store it in a Kubernetes `secret` that will be consumed by sidecar deployment
    ```bash
    ~$ ./deployment/webhook-create-signed-cert.sh \
        --service cyberark-sidecar-injector \
        --secret cyberark-sidecar-injector \
        --namespace cyberark-sidecar-injector
    ```

2. Patch the `MutatingWebhookConfiguration` by setting `caBundle` with correct value from Kubernetes cluster
    ```bash
    ~$ cat deployment/mutatingwebhook.yaml | \
        deployment/webhook-patch-ca-bundle.sh \
          --service cyberark-sidecar-injector \
          --namespace cyberark-sidecar-injector > \
        deployment/mutatingwebhook-ca-bundle.yaml
    ```

3. Deploy resources
    ```bash
    ~$ kubectl -n cyberark-sidecar-injector apply -f deployment/deployment.yaml
    ~$ kubectl -n cyberark-sidecar-injector apply -f deployment/service.yaml
    ~$ kubectl -n cyberark-sidecar-injector apply -f deployment/mutatingwebhook-ca-bundle.yaml
    ```

### Verify Sidecar Injector Installation

1. The sidecar injector webhook should be running
    ```bash
    ~$ kubectl -n cyberark-sidecar-injector get pods
    ```
    ```
    NAME                                                  READY     STATUS    RESTARTS   AGE
    cyberark-sidecar-injector-bbb689d69-882dd   1/1       Running   0          5m
    ```
    ```bash
    ~$ kubectl -n cyberark-sidecar-injector get deployment
    ```
    ```
    NAME                                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
    cyberark-sidecar-injector             1         1         1            1           5m
    ```

## Installing the Sidecar Injector (Helm)

+ Helm is **required**

To install the sidecar injector in the "injectors" namespace run the following:

```
helm --namespace injectors \
 install \
 --set "caBundle=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}')" \
 --set csrEnabled=true \
 ./charts/cyberark-sidecar-injector/
```

```
...

NOTES:
## Instructions
Before you can proceed to use the sidecar-injector, there's one last step.
You will need to approve the CSR (Certificate Signing Request) made by the sidecar-injector.
This allows the sidecar-injector to communicate securely with the Kubernetes API.

### Watch initContainer logs for when CSR is created
kubectl -n injectors logs deployment/vigilant-numbat-cyberark-sidecar-injector -c init-webhook -f

### You can check and inspect the CSR
kubectl describe csr "vigilant-numbat-cyberark-sidecar-injector.injectors"

### Approve the CSR
kubectl certificate approve "vigilant-numbat-cyberark-sidecar-injector.injectors"

Now that everything is setup you can enjoy the Cyberark Sidecar Injector.
This is the general workflow:

1. Annotate your application to enable the injector and to configure the sidecar (see README.md)
2. Webhook intercepts and injects containers as needed

Enjoy.

```

Make sure to read the NOTES section once the chart is installed; instructions are provided on how to accept the CSR request. The CSR request **must** be approved.

## Using the Sidecar Injector

### Configuration

The sidecar injector will not inject the sidecar into pods by default. Add the `sidecar-injector.cyberark.com/inject` annotation with value `true` to the pod template spec to enable injection.

The following table lists the configurable parameters of the Sidecar Injector and their default values.

| Parameter                     | Description                                     | Default                                                    |
| -----------------------       | ---------------------------------------------   | ---------------------------------------------------------- |
| `sidecar-injector.cyberark.com/inject`| Enable the Sidecar Injector by setting to `true`            | `nil` (required) |
| `sidecar-injector.cyberark.com/secretlessConfig` | ConfigMap holding Secretless configuration               |  `nil` (required for secretless)  |  
| `sidecar-injector.cyberark.com/conjurAuthConfig` | ConfigMap holding Conjur authentication configuration            |  `nil` (required for authenticator |
| `sidecar-injector.cyberark.com/conjurConnConfig` | ConfigMap holding Conjur connection configuration               |  `nil` (required for authenticator |
| `sidecar-injector.cyberark.com/injectType` | Injected Sidecar type (`secretless` or `authenticator`)                    |  `nil` (required) |
| `sidecar-injector.cyberark.com/containerMode` | Sidecar Container mode (`init` or `sidecar`)                  |  `nil` (only applies to authenticator) |
| `sidecar-injector.cyberark.com/containerName` | Sidecar Container name                  |  `nil` (only applies to secretless)                              |  

#### sidecar-injector.cyberark.com/secretlessConfig

Expected to contain the following path:

+ secretless.yml - Secretless Configuration File

#### sidecar-injector.cyberark.com/conjurConnConfig

Expected to contain the following paths:

+ CONJUR_VERSION - the version of your Conjur instance (4 or 5)
+ CONJUR_APPLIANCE_URL - the URL of the Conjur appliance instance you are connecting to
+ CONJUR_AUTHN_URL - the URL of th authenticator service endpoint
+ CONJUR_ACCOUNT - the account name for the Conjur instance you are connecting to
+ CONJUR_SSL_CERTIFICATE - the x509 certificate that was created when Conjur was initiated

#### sidecar-injector.cyberark.com/conjurAuthConfig

Expected to contain the following path:

+ CONJUR_AUTHN_LOGIN - Host login for pod e.g. namespace/service_account/some_service_account

## Secretless Sidecar Injection Example

For this section, you'll work from a test namespace (test-namespace). Later you will label this namespace with `cyberark-sidecar-injector=enabled` so as to allow the cyberark-sidecar-injector to operate on pods created in this namespace.

1. Create test namespace
    ```bash
    ~$ kubectl create namespace test-namespace
    ```

2. Label the default namespace with `cyberark-sidecar-injector=enabled`
    ```bash
    ~$ kubectl label namespace test-namespace cyberark-sidecar-injector=enabled
    ~$ kubectl get namespace -L cyberark-sidecar-injector
    ```
    ```
    NAME                            STATUS    AGE       CYBERARK-SIDECAR-INJECTOR
    default                         Active    18h
    kube-public                     Active    18h
    kube-system                     Active    18h
    cyberark-sidecar-injector       Active    18h
    test-namespace                  Active    18h       enabled
    ```

3. Create Secretless ConfigMap
    ```bash
    ~$ cat << EOL | kubectl -n test-namespace create configmap test-secretless-config --from-file=secretless.yml=/dev/stdin
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
          id: test-secret#username
        - name: password
          provider: literal
          id: test-secret#password
    EOL
    ```

4. Deploy an app with the Secretless Sidecar, take `test-app` app as an example
    ```bash
    ~$ cat << EOF | kubectl -n test-namespace create -f -
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: test-app
    spec:
      replicas: 1
      template:
        metadata:
          annotations:
            sidecar-injector.cyberark.com/inject: "yes"
            sidecar-injector.cyberark.com/secretlessConfig: "test-secretless-config"
            sidecar-injector.cyberark.com/injectType: "secretless"
          labels:
            app: test-app
        spec:
          containers:
            - name: app
              env:
                - name: http_proxy
                  value: "http://0.0.0.0:3000"
              image: googlecontainer/echoserver:1.1
    EOF
    ```

5. Verify Secretless sidecar container injected
    ```bash
    ~$ kubectl -n test-namespace get pods
    ```
    ```
    NAME                     READY     STATUS        RESTARTS   AGE
    test-app-5c55f85f5c-tn2cs   2/2       Running       0          1m
    ```

6. Test Secretless
    ```bash
    ~$ a_test_pod=$(kubectl \
     -n test-namespace \
     get po \
     -l=app=test-app \
     -o=jsonpath="{.items[0].metadata.name}")
    
    ~$ kubectl \
      -n test-namespace \
      exec ${a_test_pod} \
      -c app \
      -i \
      -- \
      curl --silent localhost:8080 | grep authorization | sed -e s/^authorization=Basic\ // | base64 --decode; echo
   
    ```
    ```
    "test-secret#username:test-secret#password"
    ```
    

## Conjur Authenticator/Secretless Sidecar Injection Example

For this section, you'll work from a test namespace (sidecar-example-app). Later you will label this namespace with `cyberark-sidecar-injector=enabled` so as to allow the cyberark-sidecar-injector to operate on pods created in this namespace.

1. Setup a Conjur appliance running with the Kubernetes authenticator installed and enabled - e.g. run `./start` in  [kubernetes-conjur-deploy](https://github.com/cyberark/kubernetes-conjur-deploy/)

1. Load Conjur policy to create a host for the service account `$TEST_APP_SERVICE_ACCOUNT` - e.g. `test-app-secretless` and `test-app-secretless` are made available by walking through [kubernetes-conjur-demo](https://github.com/conjurdemos/kubernetes-conjur-demo) until `./3_init_conjur_cert_authority.sh`

1. Set up environment variables modify to suite your needs e.g. use the same values from [kubernetes-conjur-demo](https://github.com/conjurdemos/kubernetes-conjur-demo)
    ```bash
    # required values
    export TEST_APP_SERVICE_ACCOUNT=test-app-secretless
    export containerMode=sidecar
    export CONJUR_VERSION=4
    export CONJUR_NAMESPACE_NAME=conjur-ktanekha
    export CONJUR_ACCOUNT=my-account
    export AUTHENTICATOR_ID=sidecar-test
    export TEST_APP_NAMESPACE_NAME=sidecar-example-app


    # derived values
    CONJUR_APPLIANCE_URL="https://conjur-follower.${CONJUR_NAMESPACE_NAME}.svc.cluster.local/api"
    CONJUR_AUTHN_URL="https://conjur-follower.${CONJUR_NAMESPACE_NAME}.svc.cluster.local/api/authn-k8s/${AUTHENTICATOR_ID}"
    if [ ${CONJUR_VERSION} = '4' ]; then
      CONJUR_AUTHN_LOGIN=${TEST_APP_NAMESPACE_NAME}/service_account/${TEST_APP_SERVICE_ACCOUNT}
    else
      CONJUR_AUTHN_LOGIN=host/conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/${TEST_APP_NAMESPACE_NAME}/service_account/${TEST_APP_SERVICE_ACCOUNT}
    fi
    CONJUR_SSL_CERTIFICATE=$(follower_pod_name=$(kubectl -n ${CONJUR_NAMESPACE_NAME} get pods -l role=follower --no-headers | awk '{ print $1 }' | head -1); kubectl exec -n ${CONJUR_NAMESPACE_NAME} $follower_pod_name -- cat /opt/conjur/etc/ssl/conjur.pem)
    ```

1. Create test namespace
    ```bash
    ~$ kubectl create namespace ${TEST_APP_NAMESPACE_NAME}
    ```

1. Label the default namespace with `cyberark-sidecar-injector=enabled`
    ```bash
    ~$ kubectl label namespace ${TEST_APP_NAMESPACE_NAME} cyberark-sidecar-injector=enabled
    ~$ kubectl get namespace -L cyberark-sidecar-injector
    ```
    ```
    NAME                            STATUS    AGE       CYBERARK-SIDECAR-INJECTOR
    default                         Active    18h
    kube-public                     Active    18h
    kube-system                     Active    18h
    cyberark-sidecar-injector       Active    18h
    sidecar-example-app             Active    18h       enabled
    ```

1. Create service account
    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} create serviceaccount ${TEST_APP_SERVICE_ACCOUNT}
    ```

1. Create Conjur ConfigMap
    ```bash
    ~$ cat << EOL | kubectl -n ${TEST_APP_NAMESPACE_NAME} apply -f -
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: conjur
    data:
      CONJUR_ACCOUNT: "${CONJUR_ACCOUNT}"
      CONJUR_VERSION: "${CONJUR_VERSION}"
      CONJUR_APPLIANCE_URL: "${CONJUR_APPLIANCE_URL}"
      CONJUR_AUTHN_URL: "${CONJUR_AUTHN_URL}"
      CONJUR_SSL_CERTIFICATE: |
    $(echo "${CONJUR_SSL_CERTIFICATE}" | awk '{ print "    " $0 }')
      CONJUR_AUTHN_LOGIN: "${CONJUR_AUTHN_LOGIN}"
    EOL
    ```

### Deploy Authenticator Sidecar

1. Deploy an app with the Authenticator Sidecar, take `test-app` app as an example
    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} delete pod test-app
    ~$ cat << EOF | kubectl -n ${TEST_APP_NAMESPACE_NAME} apply -f -
    apiVersion: v1
    kind: Pod
    metadata:
      annotations:
        sidecar-injector.cyberark.com/conjurAuthConfig: conjur
        sidecar-injector.cyberark.com/conjurConnConfig: conjur
        sidecar-injector.cyberark.com/containerMode: ${containerMode}
        sidecar-injector.cyberark.com/inject: "yes"
        sidecar-injector.cyberark.com/injectType: authenticator
        sidecar-injector.cyberark.com/containerName: secretless
      labels:
        app: test-app
      name: test-app
    spec:
      containers:
      - image: googlecontainer/echoserver:1.1
        name: app
        volumeMounts:
        - mountPath: /run/conjur
          name: conjur-access-token
      serviceAccountName: ${TEST_APP_SERVICE_ACCOUNT}
    EOF
    ```

1. Verify Authenticator sidecar container injected
    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} get pods
    ```
    ```
    NAME                     READY     STATUS        RESTARTS   AGE
    test-app                 2/2       Running       0          1m
    ```

1. Test Authenticator
    ```bash
    ~$ kubectl \
      -n ${TEST_APP_NAMESPACE_NAME} \
      exec test-app \
      -c app \
      -i \
      -- \
      cat /run/conjur/access-token | jq .
    ```
    ```
    {
      "data": "host/conjur/authn-k8s/sidecar-test/apps/sidecar-example-app/service_account/test-app-secretless",
      "timestamp": "2018-09-20 16:54:04 UTC",
      "signature": "aICQDREA2S-ulOxu8yMWqT9o8h_JDKuuDKIJOFBbQsL_uKZuovManGn-q2Yr4wdT9f_kJdgCNsxh9q54w2ciptn5sAFB3YzDAmqfUzjWv9pIwel2o7N2nuzIw-h7Ho6hA2PQ8V1Iz3NSILCT2JAnWDTi_--bplxqa6g72-j0xprkuFMkDvj2cd084WtMMWXii4W_5WG6BWA9jtnd72-tzhoaU4LFSRfSK7LON8aDdzyFexkM1IbjIuiF1sASBIsvnuY2GeghNDO8VciKh6dXe-sBqNlISlYOTOaQoEMIxA8Nm2t9jeYxmDHJ0IFkTmneeC2dgaJWWoF7MtfJnyPvwn_Z-bF49hkcYDL37-xJxUHPDA4QoU_4p82oqgC3NPnI",
      "key": "11cd239ab55175a3c0f93a7376abe663"
    }
    ```


### Deploy Secretless Sidecar

1. Create Secretless ConfigMap
    ```bash
    ~$ cat << EOL | kubectl -n ${TEST_APP_NAMESPACE_NAME} create configmap secretless --from-file=secretless.yml=/dev/stdin
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
          provider: conjur
          id: test-secretless-app-db/username
        - name: password
          provider: conjur
          id: test-secretless-app-db/password
    EOL
    ```

1. Deploy an app with the Secretless Sidecar, take `test-app` app as an example
    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} delete pod test-app
    ~$ cat << EOF | kubectl -n ${TEST_APP_NAMESPACE_NAME} apply -f
    apiVersion: v1
    kind: Pod
    metadata:
      annotations:
        sidecar-injector.cyberark.com/conjurAuthConfig: conjur
        sidecar-injector.cyberark.com/conjurConnConfig: conjur
        sidecar-injector.cyberark.com/inject: "yes"
        sidecar-injector.cyberark.com/injectType: secretless
        sidecar-injector.cyberark.com/secretlessConfig: secretless
      labels:
        app: test-app
      name: test-app
    spec:
      containers:
      - env:
          - name: http_proxy
            value: "http://0.0.0.0:3000"
        image: googlecontainer/echoserver:1.1
        name: app
     
      serviceAccountName: ${TEST_APP_SERVICE_ACCOUNT}
    EOF
    ```

1. Verify Secretless sidecar container injected
    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} get pods
    ```
    ```
    NAME                     READY     STATUS        RESTARTS   AGE
    test-app                 2/2       Running       0          1m
    ```

1. Test Secretless with Conjur
    ```bash
    ~$ kubectl \
      -n ${TEST_APP_NAMESPACE_NAME} \
      exec test-app \
      -c app \
      -i \
      -- \
      curl --silent localhost:8080 | grep authorization | sed -e s/^authorization=Basic\ // | base64 --decode; echo
   
    ```
    ```
    "test_app:84674b2874a5d7c952e7fec8"
    ```
