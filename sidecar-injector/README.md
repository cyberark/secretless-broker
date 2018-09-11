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

In addition, the `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` admission controllers should be added and listed in the correct order in the admission-control flag of kube-apiserver.

With `minikube`, start your cluster as follows:
```bash
~$ minikube start --memory=8192 --kubernetes-version=v1.10.0 \
    --extra-config=controller-manager.cluster-signing-cert-file="/var/lib/localkube/certs/ca.crt" \
    --extra-config=controller-manager.cluster-signing-key-file="/var/lib/localkube/certs/ca.key"
```

## Docker image

The docker image for the mutating webhook admission controller is publicly available on Dockerhub as [cyberark/sidecar-injector](https://hub.docker.com/r/cyberark/sidecar-injector/) .

## Secretless Namespace

Create a namespace "secretless", where you will deploy the Secretless Broker Sidecar Injector Webhook components.

1. Create namespace
    ```bash
    ~$ kubectl create namespace secretless
    ```

## Deploy

1. Create a signed cert/key pair and store it in a Kubernetes `secret` that will be consumed by sidecar deployment
    ```bash
    ~$ ./deployment/webhook-create-signed-cert.sh \
        --service secretless-sidecar-injector-webhook-svc \
        --secret secretless-sidecar-injector-webhook-certs \
        --namespace secretless
    ```

2. Patch the `MutatingWebhookConfiguration` by set `caBundle` with correct value from Kubernetes cluster
    ```bash
    ~$ cat deployment/mutatingwebhook.yaml | \
        deployment/webhook-patch-ca-bundle.sh \
          --service secretless-sidecar-injector-webhook-svc \
          --namespace secretless > \
        deployment/mutatingwebhook-ca-bundle.yaml
    ```

3. Deploy resources
    ```bash
    ~$ kubectl -n secretless apply -f deployment/deployment.yaml
    ~$ kubectl -n secretless apply -f deployment/service.yaml
    ~$ kubectl -n secretless apply -f deployment/mutatingwebhook-ca-bundle.yaml
    ```

## Verify

1. The sidecar injector webhook should be running
    ```bash
    ~$ kubectl -n secretless get pods
    ```
    ```
    NAME                                                  READY     STATUS    RESTARTS   AGE
    secretless-sidecar-injector-webhook-deployment-bbb689d69-882dd   1/1       Running   0          5m
    ```
    ```bash
    ~$ kubectl -n secretless get deployment
    ```
    ```
    NAME                                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
    secretless-sidecar-injector-webhook-deployment   1         1         1            1           5m
    ```

## Deploy test application

For this section, you'll work from a test namespace (test-namespace). Later you will label this namespace with `secretless-sidecar-injector=enabled` so as to allow the secretless-sidecar-injector to operate on pods created in this namespace.

1. Create test namespace
    ```bash
    ~$ kubectl create namespace test-namespace
    ```

2. Label the default namespace with `secretless-sidecar-injector=enabled`
    ```bash
    ~$ kubectl label namespace test-namespace secretless-sidecar-injector=enabled
    ~$ kubectl get namespace -L secretless-sidecar-injector
    ```
    ```
    NAME          STATUS    AGE       SECRETLESS-SIDECAR-INJECTOR
    default       Active    18h       enabled
    kube-public   Active    18h
    kube-system   Active    18h
    ```

3. Create Secretless ConfigMap
    ```bash
    ~$ cat << EOL | kubectl -n test-namespace create configmap sleep-secretless-config --from-file=secretless.yml=/dev/stdin
    listeners:
    - name: http_good_basic_auth
      debug: true
      protocol: http
      address: 0.0.0.0:8080
    
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

4. Deploy an app in Kubernetes cluster, take `sleep` app as an example
    ```bash
    ~$ cat << EOF | kubectl -n test-namespace create -f -
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: sleep
    spec:
      replicas: 1
      template:
        metadata:
          annotations:
            sidecar-injector.cyberark.com/inject: "yes"
            sidecar-injector.cyberark.com/config: "sleep-secretless-config"
            sidecar-injector.cyberark.com/injectType: "secretless"
          labels:
            app: sleep
        spec:
          containers:
          - name: sleep
            image: everpeace/curl-jq
            command: ["/bin/sleep","infinity"]
    EOF
    ```

5. Verify sidecar container injected
    ```bash
    ~$ kubectl -n test-namespace get pods
    ```
    ```
    NAME                     READY     STATUS        RESTARTS   AGE
    sleep-5c55f85f5c-tn2cs   2/2       Running       0          1m
    ```

6. Test Secretless
    ```bash
    ~$ a_sleep_pod=$(kubectl \
     -n test-namespace \
     get po \
     -l=app=sleep \
     -o=jsonpath="{.items[0].metadata.name}")
    
    ~$ kubectl \
      -n test-namespace \
      exec ${a_sleep_pod} \
      -c sleep \
      -i \
      -- \
      bash << 'EOL'
    export http_proxy=localhost:8080
    
    response=$(curl --request GET --url http://scooterlabs.com/echo.json)
    pretty_resp=$(echo "${response}" | jq -r .headers.Authorization)
    echo '"'"$(echo "${pretty_resp##* }" | base64 --decode)"'"' | jq .
    EOL
    ```
    ```
    "test-secret#username:test-secret#password"
    ```
