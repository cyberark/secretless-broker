# CyberArk Conjur

This document describes how to connect to a database using Secretless with credentials stored in CyberArk Conjur. This method avoids handlings the credentials directly and therefore mitigates an important vector for secret leakage. Your app connects to Secretless, and Secretless connects to the database. To make this possible, Secretless runs as a sidecar container next to your app, and your app connects to it over a local TCP or Unix socket.

NOTE: Code snippets provided in this document can be executed in a BASH shell or equivalent. Environment variables are used to represent required values that you'll need to provide. These required values will be used to generate other values via environment variable substitution e.g. `CONJUR_APPLIANCE_URL=https://${OSS_CONJUR_SERVICE_NAME}.${OSS_CONJUR_NAMESPACE}.svc.cluster.local`.  There'll also be code snippets that can be used to generate Conjur policy, Kubernetes resource manifests and carry out actions.

## Prerequisites

### Assumptions

-   You have deployed your CyberArk Conjur instance on your Kubernetes cluster and it is configured to use the Kubernetes authenticator
-   You have an application that requires a  [supported database](https://docs.secretless.io/Latest/en/Content/SystemReq.htm)
-   Your database is running and accessible to apps in your Kubernetes environment, it supports SSL, and the credentials for it are stored in Conjur

### Required Information

To deploy Secretless, you need the following information about your Conjur configuration:


| Required Info|Description|How we refer to it|
|--|--|--|
|Policy branch with database credentials|The fully qualified ID of the policy branch in Conjur that contains your database secrets. You will need `read`  access to the secrets in this branch so that you can see the IDs of the secrets you will need.|`${APP_SECRETS_POLICY_BRANCH}`|
|Layer/group with access to secrets|The fully qualified Conjur ID of a layer or group whose members have access to the database secrets. We will add the application host identity to this layer/group and the Secretless sidecar will authenticate to Conjur using this host identity to retrieve secrets. Note that our examples refer to a layer; if you are provided with a group, replace all references to  `!layer "/${APP_SECRETS_READER_LAYER}"`  with  `!group "/${APP_SECRETS_READER_LAYER}"`  instead.|`${APP_SECRETS_READER_LAYER}`|
|Kubernetes  [authenticator name](https://docs.conjur.org/Latest/en/Content/Integrations/ConjurDeployFollowers.htm#ConfigureConjurforautoenrollmentoffollowers)| The name of the Kubernetes authenticator configured in your Conjur instance.|`${AUTHENTICATOR_ID}`|
|Conjur instance Kubernetes namespace|The Kubernetes namespace where the Conjur instance is deployed.|`${OSS_CONJUR_NAMESPACE}`|
|Conjur instance Kubernetes service account|The Kubernetes service account associated with the Conjur instance.|`${CONJUR_SERVICE_ACCOUNT_NAME}`|
|Conjur URL|The URL of the Conjur instance, for example: `https://${OSS_CONJUR_SERVICE_NAME}.${OSS_CONJUR_NAMESPACE}.svc.cluster.local`| `${CONJUR_APPLIANCE_URL}` |
|Conjur Account|The Conjur account where the database credentials are stored and the Kubernetes authenticator is configured.|`#{CONJUR_ACCOUNT}`|
|Conjur Admin Login|The Conjur username for loading application-specific Conjur policy.|`${CONJUR_ADMIN_AUTHN_LOGIN}`|
|Conjur Admin API Key|The Conjur API Key of the Conjur username used for loading application-specific Conjur policy.|`${CONJUR_ADMIN_API_KEY}`|
|App Name|The DNS valid name of the application.|`${APP_NAME}`|
|App Namespace|The Kubernetes namespace where the application pods reside.|`${APP_NAMESPACE}`|
|App Service Account Name|The Kubernetes service account assigned to the application pods.|`${APP_SERVICE_ACCOUNT_NAME}`|


Let's capture all these values in a file called `./env.sh`. It'll be sourced by other code snippets in this document.

```bash
#!/usr/bin/env bash

APP_NAME=my-app
APP_NAMESPACE=kumbi-app-example
APP_SERVICE_ACCOUNT_NAME=my-app-sa

AUTHENTICATOR_ID="example"

APP_SECRETS_POLICY_BRANCH="apps/secrets/test"
APP_SECRETS_READER_LAYER="apps/layers/myapp"

CONJUR_ACCOUNT="example_acc"
CONJUR_APPLIANCE_URL="https://sealing-whale-conjur-oss.kumbi-conjur.svc.cluster.local"
CONJUR_ADMIN_AUTHN_LOGIN="admin"
CONJUR_ADMIN_API_KEY="33dk09k3bcgda6x9edvb22k9n861k5kv1cc5sndp9mge4sbq2ek"

OSS_CONJUR_SERVICE_ACCOUNT_NAME="conjur-sa"
OSS_CONJUR_NAMESPACE="kumbi-conjur"
```
## Add your application to Conjur policy

You can define your host using a variety of Kubernetes resources; see the  [Conjur documentation](https://docs.conjur.org/Latest/en/Content/Integrations/Kubernetes_MachineIdentity.htm)  for the available options.

Here we will use the service account-based host identity. Our host ID will look like this:

`${APP_NAMESPACE}/service_account/${APP_SERVICE_ACCOUNT_NAME}`

where  `${APP_NAMESPACE}`  is your app's Kubernetes namespace and  `${APP_SERVICE_ACCOUNT_NAME}`  is the service account assigned to the application pod.

We'll add the host identity to the  `conjur/authn-k8s/${AUTHENTICATOR_NAME}/apps`  policy branch. The host will belong to a layer with the same name as the branch. This allows the host to authenticate to Conjur with the Kubernetes authenticator.

Finally, to give the host access to the database credentials, we'll add it to the  `${APP_SECRETS_READER_LAYER}`  layer.

The bash script snippet below generates the Conjur policy that does all the above.

```bash
#!/usr/bin/env bash
. ./env.sh

cat << EOL
---
# Policy enabling the Kubernetes authenticator for your application
- !policy
  id: conjur/authn-k8s/${AUTHENTICATOR_ID}/apps
  body:
    - &hosts
      - !host
        id: ${APP_NAMESPACE}/service_account/${APP_SERVICE_ACCOUNT_NAME}
        annotations:
          kubernetes/authentication-container-name: ${APP_AUTHENTICATION_CONTAINER_NAME}
          kubernetes: "true"
    - !grant
      role: !layer
      members: *hosts

# Grant application's authn identity membership to the application secrets reader layer so authn identity inherits read privileges on application secrets
- !grant
  role: !layer ${APP_SECRETS_READER_LAYER}
  members:
  - !host /conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/${APP_NAMESPACE}/service_account/${APP_SERVICE_ACCOUNT_NAME}

EOL
```

Apply the policy:

``` bash
#!/usr/bin/env bash

conjur policy load root app-policy.yml
```

## Grant Conjur instance access to pods in the application namespace

The bash script snippet below generates Kubernetes manifest for a Role with the relevant permissions and a Role Binding of the application service account to the aforementioned Role. This Role and RoleBinding combination grant the service account assigned to the Conjur instance pod access to pods in the application namespace.

```bash
#!/usr/bin/env bash

. ./env.sh

cat << EOL > conjur-authenticator-role.yml
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: conjur-authenticator
  namespace: ${APP_NAMESPACE}
rules:
- apiGroups: [""] # "" indicates the core API group
  resources: ["pods", "serviceaccounts"]
  verbs: ["get", "list"]
- apiGroups: ["extensions"]
  resources: [ "deployments", "replicasets"]
  verbs: ["get", "list"]
- apiGroups: ["apps"]  # needed on OpenShift 3.7+
  resources: [ "deployments", "statefulsets", "replicasets"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create", "get"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: conjur-authenticator-role-binding
  namespace: ${APP_NAMESPACE}
subjects:
  - kind: ServiceAccount
    name: ${OSS_CONJUR_SERVICE_ACCOUNT_NAME}
    namespace: ${OSS_CONJUR_NAMESPACE}
roleRef:
  kind: Role
  name: conjur-authenticator
  apiGroup: rbac.authorization.k8s.io
EOL
```

After generating the manifest containing the Role and RoleBinding, create them by running:
 
```bash
#!/usr/bin/env bash

kubectl create -f conjur-authenticator-role.yml
```

## Store the Conjur SSL certificate in a ConfigMap

Fetch Conjur SSL Certificate
```bash
#!/usr/bin/env bash
. ./env.sh

openssl s_client -showcerts \
  -connect "${CONJUR_APPLIANCE_URL}" </dev/null 2>/dev/null \
  | sed -n '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > conjur.pem
```

Store Conjur SSL Certificate
```bash
#!/usr/bin/env bash
. ./env.sh

kubectl \
  --namespace "${APP_NAMESPACE}" \
  create configmap \
  conjur-cert \
  --from-file=ssl-certificate="conjur.pem"
```

## Store the Secretless configuration in a ConfigMap

The bash script snippet below generates Secretless configuration. This configuration tells Secretless how to setup the service connector. Modify to suit your needs.

```bash
#!/usr/bin/env bash
. ./env.sh

cat << EOL > secretless.yml
version: "2"
services:
  app_db:
    protocol: mysql
    listenOn: tcp://0.0.0.0:3000
    credentials:
      host:
        from: conjur
        get: ${APP_SECRETS_POLICY_BRANCH}/host
      port:
        from: conjur
        get: ${APP_SECRETS_POLICY_BRANCH}/port
      username:
        from: conjur
        get: ${APP_SECRETS_POLICY_BRANCH}/username
      password:
        from: conjur
        get: ${APP_SECRETS_POLICY_BRANCH}/password
      sslmode: disable
EOL
```

After generating the Secretless configuration, store it in a ConfigMap manifest by running the following:
 
```bash
#!/usr/bin/env bash
. ./env.sh

kubectl \
  --namespace "${APP_NAMESPACE}" \
  create configmap \
  secretless-config \
  --from-file=secretless.yml
```

## Deploy your application with Secretless

The bash script snippet below generates a Kubernetes Deployment manifest with an application + Secretless. Modify to suit your needs.

```bash
#!/usr/bin/env bash

. ./env.sh

cat << EOL > app-manifest.yml
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    app: "${APP_NAME}"
  name: "${APP_NAME}"
  namespace: "${APP_NAMESPACE}"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: "${APP_NAME}"
  template:
    metadata:
      labels:
        app: "${APP_NAME}"
    spec:
      serviceAccountName: "${APP_SERVICE_ACCOUNT_NAME}"
      containers:
      - name: app
        image: mysql/mysql-server:5.7
        command: [ "sleep", "infinity" ]
        imagePullPolicy: Always
      - name: "${APP_AUTHENTICATION_CONTAINER_NAME}"
        image: cyberark/secretless-broker:latest
        imagePullPolicy: Always
        args: ["-f", "/etc/secretless/secretless.yml"]
        env:
          - name: MY_POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: MY_POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: MY_POD_IP
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
          - name: CONJUR_VERSION
            value: 5
          - name: CONJUR_APPLIANCE_URL
            value: "${CONJUR_APPLIANCE_URL}"
          - name: CONJUR_AUTHN_URL
            value: "${CONJUR_APPLIANCE_URL}/authn-k8s/${AUTHENTICATOR_ID}"
          - name: CONJUR_ACCOUNT
            value: "${CONJUR_ACCOUNT}"
          - name: CONJUR_AUTHN_LOGIN
            value: "host/conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/${APP_NAMESPACE}/service_account/${APP_SERVICE_ACCOUNT_NAME}"
          - name: CONJUR_SSL_CERTIFICATE
            valueFrom:
              configMapKeyRef:
                name: conjur-cert
                key: ssl-certificate
        readinessProbe:
          httpGet:
            path: /ready
            port: 5335
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 2
          failureThreshold: 60
        volumeMounts:
          - mountPath: /etc/secretless
            name: config
            readOnly: true
      volumes:
        - name: config
          configMap:
            name: secretless-config
            defaultMode: 420
EOL
```

After generating the application manifest, deploy the application by running:
 
```bash
#!/usr/bin/env bash

kubectl create -f app-manifest.yml
```
