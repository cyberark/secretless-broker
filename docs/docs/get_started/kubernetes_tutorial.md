---
title: Using Secretless in Kubernetes
id: kubernetes_tutorial
layout: subpages
description: Secretless Broker Documentation
permalink: /docs/get_started/kubernetes_tutorial.html
---

This is a detailed, step-by-step tutorial. 

You will:

1. Deploy a PostgreSQL database
2. Store its credentials in Kubernetes secrets
3. Setup Secretless Broker to proxy connections to it 
4. Deploy an application that connects to the database **without knowing its password**

Already a Kubernetes expert? You may prefer our:

<div style="text-align: center">
  <a href="https://github.com/cyberark/secretless-broker/tree/master/demos/k8s-demo" class="button btn-primary gradient">Advanced Github Tutorial</a>
</div>

complete with shell scripts to get **the whole thing working end to end fast**.

## Table of Contents

+ [Overview](#overview)
+ Steps for Security Admin
  + [Create PostgreSQL Service in Kubernetes](#create-postgresql-service-in-kubernetes)
  + [Create Application Database](#create-application-database)
  + [Create Application Namespace and Store Credentials](#create-application-namespace-and-store-credentials)
  + [Create Secretless Broker Configuration ConfigMap](#create-secretless-broker-configuration-configmap)
  + [Create Application Service Account and Grant Entitlements](#create-application-service-account-and-grant-entitlements)
+ Steps for Application Developer
  + [Sample Application Overview](#sample-application-overview)
  + [Create Application Deployment Manifest](#create-application-deployment-manifest)
  + [Deploy Application With Secretless Broker](#deploy-application-with-secretless-broker)
  + [Expose Application Publicly](#expose-application-publicly)
+ [Test the Application](#test-the-application)
+ [Appendix - Secretless Deployment Manifest Explained](#appendix---secretless-deployment-manifest-explained)
  + [Networking](#networking)
  + [SSL](#ssl)
  + [Credential Access](#credential-access)
  + [Configuration Access](#configuration-access)

## Overview

Applications and application developers should be **incapable of leaking secrets**.

To achieve that goal, you'll play two roles in this tutorial:

1. A **Security Admin** who handles secrets, and has sole access to those secrets
2. An **Application Developer** with no access to secrets.

The situation looks like this:

![Image](/img/secretless_overview.jpg)

Specifically, we will:

**As the security admin:**

1. Create a PostgreSQL database
1. Create a DB user for the application
1. Add that user's credentials to Kubernetes Secrets
1. Configure Secretless to connect to PostgreSQL using those credentials

**As the application developer:**

1. Configure the application to connect to PostgreSQL via Secretless
1. Deploy the application and the Secretless sidecar

### Prerequisites

To run through this tutorial, you will need:

+ A running Kubernetes cluster (you can use
  [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) to run a
  cluster locally)
+ [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured
  to point to the cluster
+ [Docker CLI](https://docs.docker.com/install/)

## Steps for Security Admin

<div class="change-role">
  <div class="character-icon"><img src="/img/security_admin.jpg" alt="Security Admin"/></div>
  <div class="content">
    <div class="change-announcement">
      You are now the Security Admin
    </div>
    <div class="message">
      The Security Admin sets up PostgreSQL, configures Secretless, and has sole
      access to the credentials.
    </div>
  </div>
</div>


### Create PostgreSQL Service in Kubernetes

PostgreSQL is stateful, so we'll use a
[StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
to manage it.

#### Deploy PostgreSQL StatefulSet

To deploy a PostgreSQL StatefulSet:

1. Create a dedicated **namespace** for the storage backend:
    ```bash
    kubectl create namespace quick-start-backend-ns
    ```
    <pre>
    namespace "quick-start-backend-ns" created
    </pre>

1. Create a self-signed certificate (see [PostgreSQL documentation for
   more info](https://www.postgresql.org/docs/9.6/ssl-tcp.html)):

    ```bash
    openssl req -new -x509 -days 365 -nodes -text -out server.crt \
      -keyout server.key -subj "/CN=pg"
    chmod og-rwx server.key
    ```
    <pre>Generating a 2048 bit RSA private key
    ....................................................................................+++++
    .......+++++
    writing new private key to 'server.key'
    -----</pre>

1. Store the certificate files as Kubernetes secrets in the
   `quick-start-backend-ns` namespace:
    ```bash
    kubectl --namespace quick-start-backend-ns \
      create secret generic \
      quick-start-backend-certs \
      --from-file=server.crt \
      --from-file=server.key
    ```
    <pre>secret "quick-start-backend-certs" created</pre>

    <div class="note">
      While Kubernetes Secrets are more secure than hard-coded ones, in
      a real deployment you should secure secrets in a fully-featured vault, like
      Conjur.
    </div>

1. Create and save the **PostgreSQL StatefulSet manifest** in a file named **pg.yml** in your current working directory:

    ```bash
cat << EOF > pg.yml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: pg
  labels:
    app: quick-start-backend
spec:
  serviceName: quick-start-backend
  selector:
    matchLabels:
      app: quick-start-backend
  template:
    metadata:
      labels:
        app: quick-start-backend
    spec:
      securityContext:
        fsGroup: 999
      containers:
      - name: quick-start-backend
        image: postgres:9.6
        imagePullPolicy: IfNotPresent
        ports:
          - containerPort: 5432
        env:
          - name: POSTGRES_DB
            value: postgres
          - name: POSTGRES_USER
            value: security_admin_user
          - name: POSTGRES_PASSWORD
            value: security_admin_password
        volumeMounts:
        - name: backend-certs
          mountPath: "/etc/certs/"
          readOnly: true
        args: ["-c", "ssl=on", "-c", "ssl_cert_file=/etc/certs/server.crt", "-c", "ssl_key_file=/etc/certs/server.key"]
      volumes:
      - name: backend-certs
        secret:
          secretName: quick-start-backend-certs
          defaultMode: 384
EOF
    ```
    <div class="note">
      In the manifest above, the certificate files for your database server are
      mounted in a volume with <code class="highlighter-rouge">defaultMode:
      384</code> giving it permissions <code
      class="highlighter-rouge">0600</code> (Why?  Because <code
      class="highlighter-rouge">600</code> in base 8 = <code
      class="highlighter-rouge">384</code> in base 10).
    </div>
    
    <div class="note">
      The pod is deployed with <code class="highlighter-rouge">999</code> as
      the group associated with any mounted volumes, as indicated by <code
      class="highlighter-rouge">fsGroup: 999</code>.  <code
      class="highlighter-rouge">999</code> is a the static postgres gid,
      defined in the <a
      href="https://github.com/docker-library/postgres/blob/master/9.6/Dockerfile#L16">postgres
      Docker image</a>
    </div>

1. Deploy the **PostgreSQL StatefulSet**:
    ```bash
kubectl --namespace quick-start-backend-ns apply -f pg.yml
    ```
    <pre>
    statefulset "pg" created
    </pre>

    This StatefulSet uses the DockerHub
    [**postgres:9.6**](https://hub.docker.com/r/library/postgres/) container.

    On startup, the container creates a superuser from the environment
    variables `POSTGRES_USER` and `POSTGRES_PASSWORD`, which 
    we set to the values `security_admin_user` and `security_admin_password`,
    respectively.

    Going forward, we'll call these values the **admin-credentials**, to distinguish
    them from the **application-credentials** our application will use.

    In the scripts below, we'll refer to the admin-credentials by the
    environment variables `SECURITY_ADMIN_USER` and `SECURITY_ADMIN_PASSWORD`.

1. To ensure the **PostgreSQL StatefulSet** pod has started and is healthy
   (this may take a minute or so), run:
    ```bash
kubectl --namespace quick-start-backend-ns get pods
    ```
    <pre>
    NAME      READY     STATUS    RESTARTS   AGE
    pg-0      1/1       Running   0          6s
    </pre>



#### Expose PostgreSQL Service

Our **PostgreSQL StatefulSet** is running, but we still need to expose it
publicly as a Kubernetes service.

To expose the database, run:

```bash
cat << EOF > pg-service.yml
kind: Service
apiVersion: v1
metadata:
  name: quick-start-backend
spec:
  selector:
    app: quick-start-backend
  ports:
    - port: 5432
      targetPort: 5432
      nodePort: 30001
  type: NodePort

EOF
kubectl --namespace quick-start-backend-ns  apply -f pg-service.yml
```
<pre>
service "quick-start-backend" created
</pre>

<div class="note">
  The service manifest above assumes you're using minikube, where <b>NodePort</b>
  is the correct service type; for a GKE cluser, you may prefer a different
  service type, such as a <b>LoadBalancer</b>.
</div>

The database is now available at `$(minikube ip):30001`, which we'll call the
`REMOTE_DB_URL`.

The database has no data yet, but we can verify it works by logging in as the
security admin and listing the users:

```bash
export SECURITY_ADMIN_USER=security_admin_user
export SECURITY_ADMIN_PASSWORD=security_admin_password
export REMOTE_DB_URL=$(minikube ip):30001

docker run --rm -it -e PGPASSWORD=${SECURITY_ADMIN_PASSWORD} postgres:9.6 \
  psql -U ${SECURITY_ADMIN_USER} "postgres://${REMOTE_DB_URL}/postgres" -c "\du"
```

<pre>
                                          List of roles
       Role name        |                         Attributes                    
     | Member of
------------------------+-------------------------------------------------------
-----+-----------
 security_admin_user    | Superuser, Create role, Create DB, Replication, Bypass
 RLS | {}
</pre>

### Create Application Database

In this section, we assume the following:

- You already have a PostgreSQL database exposed as a Kubernetes service.
- It's publicly available via the URL in `REMOTE_DB_URL`
- You have admin-level database credentials
- The `SECURITY_ADMIN_USER` and `SECURITY_ADMIN_PASSWORD` environment variables
  hold those credentials

<div class="note">
  If you're using your own database server and it's not SSL-enabled, please see
  the <a
  href="https://docs.secretless.io/Latest/en/Content/References/handlers/postgres.htm">handler
  documentation</a> for how to disable SSL in your Secretless configuration.
</div>

If you followed along in the last section and are using minikube, you can run:

``` bash
export SECURITY_ADMIN_USER=security_admin_user
export SECURITY_ADMIN_PASSWORD=security_admin_password
export REMOTE_DB_URL="$(minikube ip):30001"
```

Next, we'll create the application database and user, and securely store the
user's credentials:

1. Create the application database
1. Create the `pets` table in that database
1. Create an application user with limited privileges: `SELECT` and `INSERT` on
   the `pets` table
1. Store these database **application-credentials** in Kubernetes secrets.

So we can refer to them later, export the database name and
application-credentials as environment variables:

``` bash
export APPLICATION_DB_NAME=quick_start_db

export APPLICATION_DB_USER=app_user
export APPLICATION_DB_INITIAL_PASSWORD=app_user_password
```

Finally, to perform the 4 steps listed above, run:

```bash
docker run --rm -i -e PGPASSWORD=${SECURITY_ADMIN_PASSWORD} postgres:9.6 \
    psql -U ${SECURITY_ADMIN_USER} "postgres://${REMOTE_DB_URL}/postgres" \
    << EOSQL

CREATE DATABASE ${APPLICATION_DB_NAME};

/* connect to it */

\c ${APPLICATION_DB_NAME};

CREATE TABLE pets (
  id serial primary key,
  name varchar(256)
);

/* Create Application User */

CREATE USER ${APPLICATION_DB_USER} PASSWORD '${APPLICATION_DB_INITIAL_PASSWORD}';

'${APPLICATION_DB_INITIAL_PASSWORD}';
/* Grant Permissions */

GRANT SELECT, INSERT ON public.pets TO ${APPLICATION_DB_USER};
GRANT USAGE, SELECT ON SEQUENCE public.pets_id_seq TO ${APPLICATION_DB_USER};
EOSQL
```
<pre>
CREATE DATABASE
You are now connected to database "quick_start_db" as user "security_admin_user".
CREATE TABLE
CREATE ROLE
GRANT
GRANT
</pre>

### Create Application Namespace and Store Credentials

The application will be scoped to the **quick-start-application-ns** namespace.

To create the namespace run:

```yaml
kubectl create namespace quick-start-application-ns
```
<pre>
namespace "quick-start-application-ns" created
</pre>

Next we'll store the application-credentials in Kubernetes Secrets:

```bash
kubectl --namespace quick-start-application-ns \
  create secret generic quick-start-backend-credentials \
  --from-literal=address="${REMOTE_DB_URL}" \
  --from-literal=username="${APPLICATION_DB_USER}" \
  --from-literal=password="${APPLICATION_DB_INITIAL_PASSWORD}"
```
<pre>
secret "quick-start-backend-credentials" created
</pre>

<div class="note">
  While Kubernetes Secrets are more secure than hard-coded ones, in a real
  deployment you should secure secrets in a fully-featured vault, like Conjur.
</div>

### Create Secretless Broker Configuration ConfigMap

With our database ready and our credentials safely stored, we can now configure
the Secretless Broker.  We'll tell it where to listen for connections and how
to proxy them.

After that, the developer's application can access the database **without ever
knowing the application-credentials**.

A Secretless Broker configuration file has 2 sections:
  - **Listeners:** Define how and where to listen for connections
  - **Handlers:** Define where to get credentials and how to connect to the
    target service (a Postgres database, in our example)

To create **secretless.yml** in your current directory, run:

```bash
cat << EOF > secretless.yml
listeners:
  - name: pets-pg-listener
    protocol: pg
    address: localhost:5432

handlers:
  - name: pets-pg-handler
    listener: pets-pg-listener
    credentials:
      - name: address
        provider: kubernetes
        id: quick-start-backend-credentials#address
      - name: username
        provider: kubernetes
        id: quick-start-backend-credentials#username
      - name: password
        provider: kubernetes
        id: quick-start-backend-credentials#password
EOF
```

Here's what this does:

- Defines a listener named `pets-pg-listener` that listens for Postgres
  connections on `localhost:5432`
- Defines a handler `pets-pg-handler` connected to that listener
- Says that the database `address`, `username` and `password` are stored in
  Kubernetes Secrets 
- Lists the ids of those credentials within Kubernetes Secrets

<div class="note">
  This configuration is shared by all Secretless Broker sidecar containers.
  There is one Secretless sidecar in every application Pod replica.
</div>

<div class="note">
  Since we don't specify an <code class="highlighter-rouge">sslmode</code> in
  the Secretless Broker config, it will use the default <code
  class="highlighter-rouge">require</code> value.
</div>


Next we create a Kubernetes `ConfigMap` from this **secretless.yml**:

```bash
kubectl --namespace quick-start-application-ns \
  create configmap \
  quick-start-application-secretless-config \
  --from-file=secretless.yml
```
<pre>
configmap "quick-start-application-secretless-config" created
</pre>

### Create Application Service Account and Grant Entitlements

To grant our application access to the credentials in Kubernetes Secrets, 
we'll need a ServiceAccount:

```bash
kubectl --namespace quick-start-application-ns \
  create serviceaccount \
  quick-start-application
```
<pre>
serviceaccount "quick-start-application" created
</pre>

Next we grant this ServiceAccount "get" access to the
**quick-start-backend-credentials**.  This is a 2 step process:

1. Create a **Role** with permissions to `get` the
   `quick-start-backend-credentials` secret
2. Create a **RoleBinding** so our ServiceAccount has this Role

Run this command to perform both steps:

```bash
cat << EOF > quick-start-application-entitlements.yml
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: quick-start-backend-credentials-reader
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["quick-start-backend-credentials"]
  verbs: ["get"]

---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: read-quick-start-backend-credentials
subjects:
- kind: ServiceAccount
  name: quick-start-application
roleRef:
  kind: Role
  name: quick-start-backend-credentials-reader
  apiGroup: rbac.authorization.k8s.io
EOF

kubectl --namespace quick-start-application-ns \
  apply -f quick-start-application-entitlements.yml
```
<pre>
role "quick-start-backend-credentials-reader" created
rolebinding "read-quick-start-backend-credentials" created
</pre>


## Steps for the Application Developer

<div class="change-role">
  <div class="character-icon"><img src="/img/application_developer.jpg" alt="Application Developer"/></div>
  <div class="content">
    <div class="change-announcement">
      You are now the application developer.  
    </div>
    <div class="message">
      You can no longer access the secrets we stored above in environment
      variables.  Open a new terminal so that all those variables are gone.
    </div>
  </div>
</div>

You know only one thing -- the name of the database:

```bash
export APPLICATION_DB_NAME=quick_start_db
```

### Sample Application Overview

The application we'll be deploying is a [pet store demo
application](https://github.com/conjurdemos/pet-store-demo) with a simple API:

- `GET /pets` lists all the pets
- `POST /pet` adds a pet

Its PostgreSQL backend is configured using a `DB_URL` environment variable: 

<pre>
postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
</pre>

Again, the application has no knowledge of the database credentials it's using.

For usage examples, please see [Test the Application](#test-the-application).

### Create Application Deployment Manifest

We're ready to deploy our application.

A detailed explanation of the manifest below is in the [Appendix - Secretless
Deployment Manifest
Explained](#appendix---secretless-deployment-manifest-explained), but isn't
needed to complete the tutorial.

To create the **quick-start-application.yml** manifest using the
`APPLICATION_DB_NAME` above, run:

```bash
cat << EOF > quick-start-application.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quick-start-application
  labels:
    app: quick-start-application
spec:
  replicas: 3
  selector:
    matchLabels:
      app: quick-start-application
  template:
    metadata:
      labels:
        app: quick-start-application
    spec:
      serviceAccountName: quick-start-application
      automountServiceAccountToken: true
      containers:
        - name: quick-start-application
          image: cyberark/demo-app:latest
          env:
            - name: DB_URL
              value: postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
        - name: secretless-broker
          image: cyberark/secretless-broker:latest
          imagePullPolicy: IfNotPresent
          args: ["-f", "/etc/secretless/secretless.yml"]
          volumeMounts:
            - name: config
              mountPath: /etc/secretless
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: quick-start-application-secretless-config
EOF
```

### Deploy Application With Secretless Broker

To deploy the application, run:

```bash
kubectl --namespace quick-start-application-ns apply -f quick-start-application.yml
```
<pre>
deployment "quick-start-application" created
</pre>

Before moving on, verify that the pods are healthy:

```bash
kubectl --namespace quick-start-application-ns get pods
```
<pre>
NAME                                       READY     STATUS        RESTARTS   AGE
quick-start-application-6bd8dbd57f-bshmf   2/2       Running       0          22s
quick-start-application-6bd8dbd57f-dr962   2/2       Running       0          26s
quick-start-application-6bd8dbd57f-fgfnh   2/2       Running       0          30s
</pre>

### Expose Application Publicly

The application is running, but not yet publicly available.

To expose it publicly as a Kubernetes Service, run:

```bash
cat << EOF > quick-start-application-service.yml
kind: Service
apiVersion: v1
metadata:
  name: quick-start-application
spec:
  selector:
    app: quick-start-application
  ports:
  - port: 8080
    targetPort: 8080
    nodePort: 30002
  type: NodePort
EOF
kubectl --namespace quick-start-application-ns \
 apply -f quick-start-application-service.yml
```
<pre>
service "quick-start-application" created
</pre>

Congratulations!

The application is now available at `$(minikube ip):30002`.  We'll call
this the `APPLICATION_URL` going forward.

## Test the Application

Let's verify everything works as expected.

First, make sure the `APPLICATION_URL` is correctly set:

```bash
export APPLICATION_URL=$(minikube ip):30002
```

Now let's create a pet (`POST /pet`):

```bash
curl -i -d @- \
 -H "Content-Type: application/json" \
 ${APPLICATION_URL}/pet \
 << EOF
{
   "name": "Secretlessly Fluffy"
}
EOF
```
<pre>
HTTP/1.1 201
Location: http://192.168.99.100:30002/pet/2
Content-Length: 0
Date: Thu, 23 Aug 2018 11:56:27 GMT
</pre>

We should get a 201 response status.

Now let's retrieve all the pets (`GET /pets`):

```bash
curl -i ${APPLICATION_URL}/pets
```
<pre>
HTTP/1.1 200
Content-Type: application/json;charset=UTF-8
Transfer-Encoding: chunked
Date: Thu, 23 Aug 2018 11:57:17 GMT

[{"id":1,"name":"Secretlessly Fluffy"}]
</pre>

We should get a 200 response with a JSON array of the pets.

That's it! 

<div class="the-big-finish">
  <p>
  The application is connecting to a password-protected Postgres database
  <b>without any knowledge of the credentials</b>.
  </p>

  <img src="/img/its_magic.jpg" alt="It's Magic"/>
</div>

For more info on configuring Secretless for your own use case, see the:

<div style="text-align: center">
  <a href="https://docs.secretless.io/Latest/en/Content/Overview/how_it_works.htm" class="button btn-primary gradient">Full Secretless Documentation</a>
</div>

## Appendix - Secretless Deployment Manifest Explained

Here we'll walk through the application deployment manifest, to better
understand how Secretless works.

We'll focus on the Pod's template, which is where the magic happens:

```yaml
  # top part elided...
  template:
    metadata:
      labels:
        app: quick-start-application
    spec:
      serviceAccountName: quick-start-application
      automountServiceAccountToken: true
      containers:
        - name: quick-start-application
          image: cyberark/demo-app:latest
          env:
            - name: DB_URL
              value: postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
        - name: secretless-broker
          image: cyberark/secretless-broker:latest
          imagePullPolicy: IfNotPresent
          args: ["-f", "/etc/secretless/secretless.yml"]
          volumeMounts:
            - name: config
              mountPath: /etc/secretless
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: quick-start-application-secretless-config
```

### Networking

Since it resides in the same pod, the application can access the Secretless
sidecar container over localhost.

As specified in the ConfigMap we created, Secretless listens on port
`5432`, and hence this:

```yaml
          env:
            - name: DB_URL
              value: postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
```

is all our application needs to locate Secretless.

### SSL

Notice the `?sslmode=disable` at the end of our `DB_URL`.

This means that **the application connects to Secretless without SSL**, which
is safe because it is intra-Pod communication over localhost.

However, the **connection between Secretless and Postgres is secure, and does
use SSL**.  

The situation looks like this:

```
                 No SSL                       SSL
Application   <---------->   Secretless   <---------->   Postgres
```

For more information on PostgreSQL SSL modes see:

- [PostgreSQL SSL documentation](https://www.postgresql.org/docs/9.6/libpq-ssl.html)
- [PostgreSQL Secretless Handler documentation](https://docs.secretless.io/Latest/en/Content/References/handlers/postgres.htm).

### Credential Access

Notice we add the **quick-start-application** ServiceAccount to the pod:

```yaml
    spec:
      serviceAccountName: quick-start-application
```

That's the ServiceAccount we created earlier, the one with access to the
credentials in Kubernetes Secrets.  This is what gives Secretless access
to those credentials. 

### Configuration Access

Finally, notice the sections defining the volumes and the volume mount in the
Secretless container:

```yaml
          # ... elided
          volumeMounts:
            - name: config
              mountPath: /etc/secretless
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: quick-start-application-secretless-config
```

Here we create a volume base on the ConfigMap we created earlier, which stores
our **secretless.yml** configuration file.

Thus Secretless gets its configuration file via a volume mount.
