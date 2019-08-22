---
title: Using Secretless in Kubernetes
id: kubernetes_tutorial
layout: tutorials
description: Secretless Broker Documentation
section-header: Steps for Security Admin
time-complete: 5
products-used: Kubernetes Secrets, PostgreSQL Service Connector
back-btn: /tutorials/kubernetes/overview.html
continue-btn: /tutorials/kubernetes/app-dev.html
up-next: As an Application Developer, you no longer need to worry about all the passwords and database connections! You will deploy an application and leave it up to the Secretless Broker to make the desired connection to the database.
permalink: /tutorials/kubernetes/sec-admin.html
---
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

1. Create and save the **PostgreSQL StatefulSet manifest** in a file named
   **pg.yml** in your current working directory:

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

The database is now available at `$(minikube ip):30001`, which we'll compose with
`REMOTE_DB_HOST` and `REMOTE_DB_PORT` variables.

The database has no data yet, but we can verify it works by logging in as the
security admin and listing the users:

```bash
export SECURITY_ADMIN_USER=security_admin_user
export SECURITY_ADMIN_PASSWORD=security_admin_password
export REMOTE_DB_HOST=$(minikube ip)
export REMOTE_DB_PORT=30001

docker run --rm -it -e PGPASSWORD=${SECURITY_ADMIN_PASSWORD} postgres:9.6 \
  psql -U ${SECURITY_ADMIN_USER} "postgres://${REMOTE_DB_HOST}:${REMOTE_DB_PORT}/postgres" -c "\du"
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
- It's publicly available at `$REMOTE_DB_HOST:$REMOTE_DB_PORT`
- You have admin-level database credentials
- The `SECURITY_ADMIN_USER` and `SECURITY_ADMIN_PASSWORD` environment variables
  hold those credentials

<div class="note">
  If you're using your own database server and it's not SSL-enabled, please see
  the <a
  href="https://docs.secretless.io/Latest/en/Content/References/connectors/postgres.htm">service connector
  documentation</a> for how to disable SSL in your Secretless configuration.
</div>

If you followed along in the last section and are using minikube, you can run:

``` bash
export SECURITY_ADMIN_USER=security_admin_user
export SECURITY_ADMIN_PASSWORD=security_admin_password
export REMOTE_DB_HOST="$(minikube ip)"
export REMOTE_DB_PORT="30001"
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
    psql -U ${SECURITY_ADMIN_USER} "postgres://${REMOTE_DB_HOST}:${REMOTE_DB_PORT}/postgres" \
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
  --from-literal=host="${REMOTE_DB_HOST}" \
  --from-literal=port="${REMOTE_DB_PORT}" \
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

A Secretless Broker configuration file defines the services that Secretless with authenticate to on behalf of your application.

To create **secretless.yml** in your current directory, run:

```bash
cat << EOF > secretless.yml
version: "2"
services:
  pets-pg:
    connector: pg
    listenOn: tcp://localhost:5432
    credentials:
      host:
        from: kubernetes
        get: quick-start-backend-credentials#host
      port:
        from: kubernetes
        get: quick-start-backend-credentials#port
      username:
        from: kubernetes
        get: quick-start-backend-credentials#username
      password:
        from: kubernetes
        get: quick-start-backend-credentials#password
EOF
```

Here's what this does:

- Defines a service called `pets-pg` that listens for PostgreSQL connections
  on `localhost:5432`
- Says that the database `host`, `port`, `username` and `password` are stored in
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
