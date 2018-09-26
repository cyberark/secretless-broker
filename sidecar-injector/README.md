# CyberArk Sidecar Injector

This document shows how to deploy and use the CyberArk Broker Sidecar Injector [MutatingAdmissionWebhook](https://kubernetes.io/docs/admin/admission-controllers/#mutatingadmissionwebhook-beta-in-19) Server which injects sidecar container(s) into a pod prior to persistence of the underlying object.

  * [Prerequisites](#prerequisites)
    + [Mandatory TLS](#mandatory-tls)
  * [Available Sidecars](#available-sidecars)
    + [Secretless](#secretless)
    + [Authenticator](#authenticator)
  * [Installation](#installation)
    + [Installing the Sidecar Injector (Manually)](#installing-the-sidecar-injector-manually)
      - [Dedicated Namespace](#dedicated-namespace)
      - [Deploy Sidecar Injector](#deploy-sidecar-injector)
      - [Verify Sidecar Injector Installation](#verify-sidecar-injector-installation)
    + [Installing the Sidecar Injector (Helm)](#installing-the-sidecar-injector-helm)
  * [Using the Sidecar Injector](#using-the-sidecar-injector)
    + [Configuration](#configuration)
      - [sidecar-injector.cyberark.com/secretlessConfig](#sidecar-injectorcyberarkcomsecretlessconfig)
      - [sidecar-injector.cyberark.com/conjurConnConfig](#sidecar-injectorcyberarkcomconjurconnconfig)
      - [sidecar-injector.cyberark.com/conjurAuthConfig](#sidecar-injectorcyberarkcomconjurauthconfig)
  * [Secretless Sidecar Injection Example](#secretless-sidecar-injection-example)
  * [Conjur Authenticator/Secretless Sidecar Injection Example](#conjur-authenticatorsecretless-sidecar-injection-example)
    + [Deploy Authenticator Sidecar](#deploy-authenticator-sidecar)
    + [Deploy Secretless Sidecar](#deploy-secretless-sidecar)

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
~$ minikube start --kubernetes-version=v1.10.0
```

### Available Sidecars

This section enumerates the available sidecars. The choice of sidecar is made via annotations - see the [Configuration](#configuration) section for details.

1. [Secretless](#secretless)
1. [Authenticator](#authenticator)

#### Secretless
See the [README](https://github.com/cyberark/secretless-broker)

Injects: 
  + a configurable **Secretless container** with
  + a **Volume Mount** at `/etc/secretless` directed to
  + a newly created **Volume** named `secretless-config` sourced from a ConfigMap

#### Authenticator
See the [README](https://github.com/cyberark/conjur-authn-k8s-client)

Injects: 
  + a configurable **Authenticator container** with
  + a **Volume Mount** at `/run/conjur` directed to
  + a newly created shareable **in-memory Volume** named `conjur-access-token` used to store the access token

### Mandatory TLS

Supporting TLS for external webhook server is required because admission is a high security operation. As part of the installation process, we need to create a TLS certificate signed by a trusted CA (shown below is the Kubernetes CA but you can use your own) to secure the communication between the webhook server and kube-apiserver. For the complete steps of creating and approving Certificate Signing Requests(CSR), please refer to [Managing TLS in a cluster](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/).


## Docker Image

The docker image for the mutating admission webhook server is publicly available on Dockerhub as [cyberark/sidecar-injector](https://hub.docker.com/r/cyberark/sidecar-injector/).

The docker image entrypoint is the server binary. The binary supports the following flags:
```bash
-tlsCertFile=/etc/webhook/certs/cert.pem
-tlsKeyFile=/etc/webhook/certs/key.pem            
-port=8080                       
```
It also supports [glog](https://github.com/golang/glog) flags e.g. `-v` for verbosity of logs

## Installation

Installation is possible either
+ [manually](#installing-the-sidecar-injector-manually) or
+ using a [Helm chart](#installing-the-sidecar-injector-helm)


### Installing the Sidecar Injector (Manually)

#### Dedicated Namespace

Create a namespace `injectors`, where you will deploy the CyberArk Sidecar Injector Webhook components.

1. Create namespace
    ```bash
    ~$ kubectl create namespace injectors
    ```

#### Deploy Sidecar Injector

1. Create a signed cert/key pair and store it in a Kubernetes `secret` that will be consumed by sidecar injector deployment
    ```bash
    ~$ ./deployment/webhook-create-signed-cert.sh \
        --service cyberark-sidecar-injector \
        --secret cyberark-sidecar-injector \
        --namespace injectors
    ```

2. Patch the `MutatingWebhookConfiguration` by setting `caBundle` with correct value from Kubernetes cluster
    ```bash
    ~$ cat deployment/mutatingwebhook.yaml | \
        deployment/webhook-patch-ca-bundle.sh \
          --service cyberark-sidecar-injector \
          --namespace injectors > \
        deployment/mutatingwebhook-ca-bundle.yaml
    ```

3. Deploy resources
    ```bash
    ~$ kubectl -n injectors apply -f deployment/deployment.yaml
    ~$ kubectl -n injectors apply -f deployment/service.yaml
    ~$ kubectl -n injectors apply -f deployment/mutatingwebhook-ca-bundle.yaml
    ```

#### Verify Sidecar Injector Installation

1. The sidecar injector webhook should be running
    ```bash
    ~$ kubectl -n injectors get pods
    ```
    ```
    NAME                                                  READY     STATUS    RESTARTS   AGE
    cyberark-sidecar-injector-bbb689d69-882dd   1/1       Running   0          5m
    ```
    ```bash
    ~$ kubectl -n injectors get deployment
    ```
    ```
    NAME                                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
    cyberark-sidecar-injector             1         1         1            1           5m
    ```

### Installing the Sidecar Injector (Helm)

+ [Helm](https://docs.helm.sh/using_helm/) is **required**

To install the sidecar injector in the `injectors` namespace run the following:

```
helm --namespace injectors \
 install \
 --set "caBundle=$(kubectl -n kube-system \
   get configmap \
   extension-apiserver-authentication \
   -o=jsonpath='{.data.client-ca-file}' \
 )" \
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
+ CONJUR_AUTHN_URL - the URL of the authenticator service endpoint
+ CONJUR_ACCOUNT - the account name for the Conjur instance you are connecting to
+ CONJUR_SSL_CERTIFICATE - the x509 certificate that was created when Conjur was initiated

#### sidecar-injector.cyberark.com/conjurAuthConfig

Expected to contain the following path:

+ CONJUR_AUTHN_LOGIN - Host login for pod e.g. namespace/service_account/some_service_account

## Secretless Sidecar Injection Example

For this section, you'll work from a test namespace `$TEST_APP_NAMESPACE_NAME` (see below). Later you will label this namespace with `cyberark-sidecar-injector=enabled` so as to allow the cyberark-sidecar-injector to operate on pods created in this namespace.

1. Set test namespace environment variable

    ```bash
    export TEST_APP_NAMESPACE_NAME=secretless-sidecar-test 
   ```
1. Create test namespace
    ```bash
    ~$ kubectl create namespace ${TEST_APP_NAMESPACE_NAME}
    ```

2. Label the default namespace with `cyberark-sidecar-injector=enabled`
    ```bash
    ~$ kubectl label \
      namespace ${TEST_APP_NAMESPACE_NAME} \
      cyberark-sidecar-injector=enabled

    ~$ kubectl get namespace \
      -L cyberark-sidecar-injector
    ```
    ```
    NAME                            STATUS    AGE       CYBERARK-SIDECAR-INJECTOR
    default                         Active    18h
    kube-public                     Active    18h
    kube-system                     Active    18h
    cyberark-sidecar-injector       Active    18h
    secretless-sidecar-test         Active    18h       enabled
    ```

3. Create Secretless ConfigMap

    This configuration sets up an `http` listener on `0.0.0.0:3000`, with a `basic_auth` handler that retrieves user and password using the `literal` secret provider. 
   
   As shown below, the username and password are the literal values `test-secret#username` and `test-secret#password`, respectively.

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
          provider: literal
          id: test-secret#username
        - name: password
          provider: literal
          id: test-secret#password
    EOL
    ```

4. Deploy an **echo server** app with the Secretless Sidecar:
   
   The app is an **echo server** listening on port **8080**, which echoes the request header of any requests sent to it. 
   
   The **Secretless Sidecar** is injected into the application pod on pod creation via the sidecar injector. The injection is configured via annotations. 
   
   + The `secretless` ConfigMap is used as a source for Secretless configuration.

    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} \
      delete pod \
      test-app --ignore-not-found

    ~$ cat << EOF | kubectl -n ${TEST_APP_NAMESPACE_NAME} create -f -
    apiVersion: v1
    kind: Pod
    metadata:
      name: test-app
      annotations:
        sidecar-injector.cyberark.com/inject: "yes"
        sidecar-injector.cyberark.com/secretlessConfig: "secretless"
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
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} get pods
    ```
    ```
    NAME                     READY     STATUS        RESTARTS   AGE
    test-app                 2/2       Running       0          1m
    ```

6. Test Secretless

    In this step, you test Secretless by `exec`ing into the application pod's main container and issuing an HTTP request against the echo server proxied by Secretless. 
    
    The pod spec for the echo server sets the environment variable `http_proxy` within the application to the Secretless `http` listener's address `http://0.0.0.0:3000`. This allows Secretless to inject HTTP Authorization headers when proxying the request, as per the Secretless configuration above.
    
    The HTTP Authorization headers are extracted and base64 decoded from the response to retrieve the username and password.

    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} \
      exec test-app \
      -c app \
      -i \
      -- \
      curl --silent localhost:8080 \
        | grep authorization \
        | sed -e s/^authorization=Basic\ // \
        | base64 --decode; echo
   
    ```
    ```
    test-secret#username:test-secret#password
    ```
    

## Conjur Authenticator/Secretless Sidecar Injection Example

For this section, you'll work from a test namespace `$TEST_APP_NAMESPACE_NAME` (see below). Later you will label this namespace with `cyberark-sidecar-injector=enabled` so as to allow the cyberark-sidecar-injector to operate on pods created in this namespace.

1. Setup a Conjur appliance running with the Kubernetes authenticator installed and enabled - e.g. run `./start` in  [kubernetes-conjur-deploy](https://github.com/cyberark/kubernetes-conjur-deploy/)

1. Load Conjur policy to create a host for the service account `$TEST_APP_SERVICE_ACCOUNT` - e.g. `test-app-secretless` is made available by walking through [kubernetes-conjur-demo](https://github.com/conjurdemos/kubernetes-conjur-demo) up to and including `./3_init_conjur_cert_authority.sh`

1. Set up environment variables, if not already set from [kubernetes-conjur-demo](https://github.com/conjurdemos/kubernetes-conjur-demo
    ```bash
    # REQUIRED values, identical to those used for
    # kubernetes-conjur-deploy and kubernetes-conjur-demo
    export CONJUR_VERSION=...
    export CONJUR_NAMESPACE_NAME=...
    export CONJUR_ACCOUNT=...
    export AUTHENTICATOR_ID=...
    export TEST_APP_NAMESPACE_NAME=...
    ```

1. Set up pod-specific environment variables - modify to match your environment
    ```bash
    # REQUIRED values
    export TEST_APP_SERVICE_ACCOUNT=test-app-secretless
    export containerMode=sidecar
    ```

1. Generate derived Conjur connection environment variables
    ```bash
    # derived values
    ## CONJUR_APPLIANCE_URL
    CONJUR_APPLIANCE_URL="https://conjur-follower.${CONJUR_NAMESPACE_NAME}.svc.cluster.local/api"
    ## CONJUR_AUTHN_URL
    CONJUR_AUTHN_URL="https://conjur-follower.${CONJUR_NAMESPACE_NAME}.svc.cluster.local/api/authn-k8s/${AUTHENTICATOR_ID}"
    ## CONJUR_AUTHN_LOGIN
    if [ ${CONJUR_VERSION} == '4' ]; then
      CONJUR_AUTHN_LOGIN=${TEST_APP_NAMESPACE_NAME}/service_account/${TEST_APP_SERVICE_ACCOUNT}
    else
      CONJUR_AUTHN_LOGIN=host/conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/${TEST_APP_NAMESPACE_NAME}/service_account/${TEST_APP_SERVICE_ACCOUNT}
    fi
    ## CONJUR_SSL_CERTIFICATE
    ### get one of the follower pod names
    follower_pod_name=$(kubectl -n ${CONJUR_NAMESPACE_NAME} \
      get pods \
      -l role=follower --no-headers \
      | awk '{ print $1 }' | head -1);
  
    CONJUR_SSL_CERTIFICATE=$(kubectl -n ${CONJUR_NAMESPACE_NAME} \
      exec \
      $follower_pod_name \
      -- cat /opt/conjur/etc/ssl/conjur.pem)
    ```

1. Create test namespace
    ```bash
    ~$ kubectl create namespace ${TEST_APP_NAMESPACE_NAME}
    ```

1. Label the default namespace with `cyberark-sidecar-injector=enabled`
    ```bash
    ~$ kubectl label \
      namespace ${TEST_APP_NAMESPACE_NAME} \
      cyberark-sidecar-injector=enabled

    ~$ kubectl get namespace -L cyberark-sidecar-injector
    ```
    ```
    NAME                            STATUS    AGE       CYBERARK-SIDECAR-INJECTOR
    default                         Active    18h
    kube-public                     Active    18h
    kube-system                     Active    18h
    cyberark-sidecar-injector       Active    18h
    conjur-sidecar-test             Active    18h       enabled
    ```

1. Create service account (might already exist from `kubernetes-conjur-demo`) to be used by the application pod
    
    This service account maps to the Conjur identity for the pod

    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} \
    create serviceaccount ${TEST_APP_SERVICE_ACCOUNT}
    ```

1. Create Conjur ConfigMap

    This ConfigMap named `conjur` stores the connection details to the Conjur appliance.
    These details are necessary for both the **Authenticator** and **Secretless** sidecars to communicate with the Conjur appliance.
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

1. You can now leverage Conjur by either
   1. [Deploying the Authenticator Sidecar](#deploy-authenticator-sidecar) or
   1. [Deploying the Secretless Sidecar](#deploy-authenticator-sidecar) 

### Deploy Authenticator Sidecar

1. Deploy an app with the Authenticator Sidecar:
    
    The **Secretless Sidecar** is injected into the application pod on pod creation via the sidecar injector. The injection is configured via annotations. 
    
    + The `conjur` ConfigMap is used for both Conjur Authentication and Connection configuration
    + The `sidecar-injector.cyberark.com/containerName` is set to "secretless" because the corresponding Conjur identity to the service account used expects the sidecar container to be named secretless.

    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} \
      delete pod \
      test-app --ignore-not-found
   
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

    In this step, you test the Authenticator by `exec`ing into the application pod's main container and read the contents of `/run/conjur/access-token`. 
    
    The `/run/conjur/access-token` file contains the access token which is injected by the **Authenticator** sidecar upon successful authentication against the Conjur appliance.

    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} \
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

1. Create Secretless ConfigMap:

   This configuration sets up an `http` listener on `0.0.0.0:3000`, with a `basic_auth` handler that retrieves user and password using the `conjur` secret provider. 
   
   The username and password are set and stored within the Conjur appliance.
   
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

1. Deploy an **echo server** app with the Secretless Sidecar:

    The app is an **echo server** listening on port **8080**, which echoes the request header of any requests sent to it. 
    
    The **Secretless Sidecar** is injected into the application pod on pod creation via the sidecar injector. The injection is configured via annotations. 
    
    + The `conjur` ConfigMap is used for both Conjur Authentication and Connection configuration
    + The `secretless` ConfigMap is used as a source for Secretless configuration.
    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} \
    delete pod \
    test-app --ignore-not-found
    
    ~$ cat << EOF | kubectl -n ${TEST_APP_NAMESPACE_NAME} apply -f -
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
    
    In this step, you test Secretless by `exec`ing into the application pod's main container and issuing an HTTP request against the echo server proxied by Secretless. 
    
    The pod spec for the echo server sets the environment variable `http_proxy` within the application to the Secretless `http` listener's address `http://0.0.0.0:3000`. This allows Secretless to inject HTTP Authorization headers when proxying the request, as per the Secretless configuration above.
    
    The HTTP Authorization headers are extracted and base64 decoded from the response to retrieve the username and password.
    
    ```bash
    ~$ kubectl -n ${TEST_APP_NAMESPACE_NAME} \
      exec test-app \
      -c app \
      -i \
      -- \
      curl --silent localhost:8080 \
        | grep authorization \
        | sed -e s/^authorization=Basic\ // \
        | base64 --decode; echo
    ```
    ```
    "test_app:84674b2874a5d7c952e7fec8"
    ```
