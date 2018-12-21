---
title: How to Deploy
id: how_to_deploy
layout: docs
description: Secretless Broker Documentation
permalink: docs/get_started/how-to-deploy/using-conjur.html
---

## Using CyberArk Conjur

### 1. Running Conjur
To begin, make sure you have an instance of Conjur available to your Kubernetes
cluster (either internally or externally). A quick start guide is available at
the [Conjur website](https://www.conjur.org/get-started/). As you're getting set
up, take note of the hostname used for your Conjur service, as well as the
account name you're using. These will be needed for the next step.

### 2. Adding the Secretless Broker Sidecar Container
Next, we start by adding the Secretless Broker sidecar to an existing service
definition. This includes adding the Secretless Broker container and a ConfigMap
for the Secretless Broker configuration. In this example, the Secretless Broker
will be configured to authenticate local connections to a remote PostgreSQL
database. For documentation on the other handlers available, visit
[Handlers](/docs/reference/handlers/overview.html). Be sure to change the
`CONJUR_APPLIANCE_URL` and `CONJUR_ACCOUNT` environment variables to match your
own Conjur configuration.
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
      environment:
        CONJUR_APPLIANCE_URL: http://conjur.internal
        CONJUR_ACCOUNT: demo
        CONJUR_AUTHN_LOGIN: host/my-service
        CONJUR_AUTHN_API_KEY: ${CONJUR_AUTHN_API_KEY}
      ports:
      - containerPort: 5432
      volumeMounts:
      - name: config
        mountPath: "/etc/secretless"
        readOnly: true

    # <-- Add your own container definition here -->
    # - name: my-service
    #   image: my-service:latest

    - name: config
      configMap:
        name: my-service-secretless-config
```

### 3. Configuring the Secretless Broker
The next step is to define a Secretless Broker configuration. Write the
following YAML to a file named `secretless.yml`.
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
        provider: conjur
        id: my-service/address
      - name: username
        provider: conjur
        id: my-service/username
      - name: password
        provider: conjur
        id: my-service/password
```
Note: by default Secretless Broker will connect to PostgreSQL using
`sslmode=require`. For information on additional `sslmode` values available,
please see the [handler documentation](/docs/reference/handlers/overview.html).
---
Create a new ConfigMap in Kubernetes using the newly created `secretless.yml`.
``` bash

kubectl create configmap my-service-secretless-config --from-file=secretless.yml
```

### 3. Preparing Conjur
Our `secretless.yml` uses the
[conjur provider](/docs/reference/providers/file.html) to resolve credentials
required to connect to PostgreSQL. Here we create a
[Conjur policy](https://www.conjur.org/get-started/key-concepts/intro-to-conjur-policy.html)
which defines our application, its credentials and permissions.

``` yaml
---
- !policy
  id: my-service
  body:
  - &secrets
    - !variable address
    - !variable username
    - !variable password

  - !host
  - !layer

  - !grant
    role: !layer
    member: !host

  - !permit
    role: !layer
    privileges: [ read, execute ]
    resources: *secrets
```

Save the above policy as `my-service.yml` and load it into Conjur using the
following command, then export `CONJUR_AUTHN_API_KEY` to the value of `api_key`
returned in JSON.
``` bash
$ conjur policy load root my-service.yml
Loaded policy 'root'
{
  "created_roles": {
    "demo:host:my-service": {
      "id": "demo:host:my-service",
      "api_key": "393rne9kpn5gy1xf6wa63jd17emkztvmt9xf2yq2ecphwa1c60cg2"
    }
  },
  "version": 1
}
$ export CONJUR_AUTHN_API_KEY=393rne9kpn5gy1xf6wa63jd17emkztvmt9xf2yq2ecphwa1c60cg2
```

### 4. Running

Apply the manifest. Once running, PostgreSQL will be available within the Pod at
`localhost:5432`. You may need to make a change to your applications
configuration to update the address of the database. References to username or
password can be safely removed.
``` bash
sed '/s\${CONJUR_AUTHN_API_KEY}/$CONJUR_API_KEY/g' my-service.yml \
  | kubectl apply -f -
```

### 5. Next
We've just completed a quick deployment of the Secretless Broker to an existing
application using Conjur.
- Learn about the different [Credential Providers](/docs/reference/providers/overview.html)
- Learn about other [Handlers](/docs/reference/handlers/overview.html)
