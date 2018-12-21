---
title: How to Deploy
id: how_to_deploy
layout: docs
description: Secretless Broker Documentation
permalink: docs/get_started/how-to-deploy/using-kubernetes-secrets.html
---

## Using Kubernetes Secrets

### 1. Adding the Secretless Broker Sidecar Container
To begin, we start by adding the Secretless Broker sidecar to an existing
service definition. This includes adding the Secretless Broker container, a
Kubernetes Secrets volume and a ConfigMap for the Secretless configuration. In
this example, the Secretless Broker will be configured to authenticate
connections to a PostgreSQL database. For documentation on the other handlers
available, visit [Handlers](/docs/reference/handlers/overview.html).
``` yaml
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: my-service
  namespace: demo
  labels:
    app: my-service
  spec:
    containers:
    - name: secretless-broker
      image: cyberark/secretless-broker:latest
      args: ["-f", "/etc/secretless/secretless.yml"]
      ports:
      - containerPort: 5432
      volumeMounts:
      - name: secret
        mountPath: "/etc/secret"
        readOnly: true
      - name: config
        mountPath: "/etc/secretless"
        readOnly: true

    # <-- Add your own container definition here -->
    # - name: my-service
    #   image: my-service:latest

    volumes:
    - name: secret
      secret:
        secretName: my-service-postgres
        items:
        - key: address
          path: address
        - key: username
          path: username
        - key: password
          path: password

    - name: config
      configMap:
        name: my-service-secretless-config
```
### 2. Configuring the Secretless Broker
Next, we'll define a Secretless Broker configuration. Write the following YAML
to a file named `secretless.yml`.
``` yaml
listeners:
  - name: pg
    protocol: pg
    address: 0.0.0.0:5432

handlers:
  - name: pg
    listener: pg
    credentials:
      - name: address
        provider: file
        id: /etc/secret/address
      - name: username
        provider: file
        id: /etc/secret/username
      - name: password
        provider: file
        id: /etc/secret/password
```
Note: by default Secretless Broker will connect to PostgreSQL using
`sslmode=require`. For information on additional `sslmode` values available,
please see the [handler documentation](/docs/reference/handlers/overview.html).
---
Create a new ConfigMap in Kubernetes using the newly created `secretless.yml`.
``` bash
kubectl create configmap my-service-secretless-config --from-file=secretless.yml
```
----
Our `secretless.yml` uses the
[file provider](/docs/reference/providers/file.html) to resolve credentials
required to connect to PostgreSQL. Here we create a Kubernetes Secret to store
our credentials.
``` bash
kubectl create secret generic my-service-postgres \
  --from-literal=address=$POSTGRES_ADDRESS \
  --from-literal=username=$POSTGRES_USERNAME \
  --from-literal=password=$POSTGRES_PASSWORD
```

### 3. Running

Apply the manifest. Once running, PostgreSQL will be available within the Pod at
`localhost:5432`. You may need to make a change to your applications
configuration to update the address of the database. References to username or
password can be safely removed.
``` bash
kubectl apply -f my-service.yml
```

### 4. Next
We've just completed a quick deployment of the Secretless Broker to an existing
application using Kubernetes Secrets.
- Learn how to [deploy Secretless Broker with Conjur](/docs/get_started/how-to-deploy/using-conjur.html)
- Learn about the different [Credential Providers](/docs/reference/providers/overview.html)
- Learn about other [Handlers](/docs/reference/handlers/overview.html)
