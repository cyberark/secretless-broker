---
title: Kubernetes Tutorial
id: kubernetes_tutorial
layout: docs
description: Secretless Broker Documentation
permalink: /docs/get_started/kubernetes_tutorial.html
redirect_to: https://docs.secretless.io/Latest/en/Content/Get%20Started/kube_tut_get_started.htm
---
<<<<<<< HEAD

To get started with the Secretless Broker, try working through this tutorial, which goes through deploying the Secretless Broker with a sample application that uses a PostgreSQL database for backend storage.

We have chosen a PostgreSQL database as the target service for this tutorial, however the Secretless Broker comes built-in with support for several other target services; check out our [reference](/docs/reference.html) for more info.

If you'd rather just see **the whole thing working end to end**, check out [our tutorial on github](https://github.com/cyberark/secretless-broker/tree/master/demos/k8s-demo), complete with shell scripts for walking through the steps yourself and configurable to suit your needs.


## Table of Contents

+ [Getting Started](#getting-started)
+ [Sample Application](#sample-application)
+ Steps for the admin user
  + [Provision Target Services (optional)](#provision-target-services-optional)
  + [Setup And Configure PostgreSQL Target Service](#setup-and-configure-postgresql-target-service)
  + [Create Application Namespace and Store Application DB-Credentials](#create-application-namespace-and-store-application-db-credentials)
  + [Create Secretless Broker Configuration ConfigMap](#create-secretless-broker-configuration-configmap)
  + [Create Application Service Account and Grant Entitlements](#create-application-service-account-and-grant-entitlements)
+ Steps for the non-privileged user (i.e developer)
  + [Build Application Deployment Manifest](#build-application-deployment-manifest)
    + [Add & Configure Application Container](#add--configure-application-container)
    + [Add & Configure Secretless Broker Sidecar Container](#add--configure-secretless-broker-sidecar-container)
    + [Completed Application Deployment Manifest](#completed-application-deployment-manifest)
  + [Deploy Application With Secretless Broker](#deploy-application-with-secretless-broker)
    + [Expose Application Publicly](#expose-application-publicly)
+ [Try The Running Application](#try-the-running-application)
+ [Rotate Target Service Credentials](#rotate-target-service-credentials)
+ [Review Complete Tutorial With Scripts](#review-complete-tutorial-with-scripts)

## Getting started

In this tutorial, we will walk through creating an application that communicates with a password-protected PostgreSQL database **via the Secretless Broker**. _The application does not need to know anything about the credentials required to connect to the database;_ the admin super-user who provisions and configures the database will also configure the Secretless Broker to be able to communicate with it. The developer writing the application only needs to know the socket or address that the Secretless Broker is listening on to proxy the connection to the PostgreSQL backend.

We are going to do the following:

**As the admin super-user:**

1. Provision the PostgreSQL target service
1. Configure the PostgreSQL target service for usage by the application and add its credentials to secret store
1. Configure the Secretless Broker to broker the connection to the PostgreSQL target service using the credentials from the secret store

**As the application developer:**
1. Configure the application to connect to the PostgreSQL target service via the Secretless Broker
1. Deploy and run the Secretless Broker adjacent to the application

### Prerequisites

To run through this tutorial, you will need:

+ A running Kubernetes cluster (you can use [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) to run a cluster locally)
+ [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured to point to the cluster
+ [Docker CLI](https://docs.docker.com/install/)

Our Kubernetes deployment manifests assume that you are using minikube, so we use **NodePort** to expose the services; you may prefer to use a **LoadBalancer** for a GKE cluster.

## Sample Application

The tutorial uses an existing [pet store demo application](https://github.com/conjurdemos/pet-store-demo) that exposes the following routes:

- `GET /pets` to list all the pets in inventory
- `POST /pet` to add a pet
  - Requires **Content-Type: application/json** header and body that includes **name** data

There are additional routes that are also available, but these are the two that we will focus on for this tutorial.

Pet data is stored in a PostgreSQL database, and the application may be configured to connect to the database by setting the `DB_URL` environment variable in the application's environment (following [12-factor principles](https://12factor.net/)).

See [Try The Running Application](#try-the-running-application) for examples of consuming the routes using `curl`.

The database is **credential-protected**. With the Secretless Broker, we will be able to set the `DB_URL` environment variable to a value of the form:  `postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable`. The application will not require knowledge of credentials to connect to the database.

Note that although the local application-to-Secretless connection does not require SSL, by default the Secretless-to-PostgreSQL connection uses `sslmode=require`. For more information on PostgreSQL SSL modes please see the
[PostgreSQL documentation](https://www.postgresql.org/docs/9.6/libpq-ssl.html) or
our documentation on the [PostgreSQL handler](/docs/reference/handlers/postgres.html).

## Steps for the admin user

The following steps would be taken by an admin-level user, who has the ability to create and configure a database and to add secret values to a secret store.

### Provision Target Services (optional)

If you already have a PostgreSQL server running and want to use that in this tutorial, please continue to [Setup And Configure PostgreSQL Target Service](#setup-and-configure-postgresql-target-service).

In this section, you will create the PostgreSQL storage backend in your Kubernetes cluster.

Our PostgreSQL storage backend is stateful. For this reason, we'll use a StatefulSet to manage the backend. Please consult [the Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) to understand what a StatefulSet is and when it's appropriate to use it.

#### Deploy PostgreSQL StatefulSet

To deploy a PostgreSQL StatefulSet, follow these steps:

**1.** To create a dedicated **namespace** for the storage backend, run the command:
```bash
kubectl create namespace quick-start-backend-ns
```
<pre>
namespace "quick-start-backend-ns" created
</pre>

**2.** Create a simple self-signed certificate for the server, and upload to secret store.
To create the certificate, you can run the following commands (as described in
the [PostgreSQL documentation](https://www.postgresql.org/docs/9.6/ssl-tcp.html)):
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

Now that the certificates have been created, you will proceed to store them in Kubernetes secrets. This is better than hard-coding them - but in a real deployment you would want to store your secrets in a fully-featured vault, like Conjur.

Run this code to store the certificate files in a Kubernetes secret in the
`quick-start-backend-ns` namespace:
```bash
kubectl --namespace quick-start-backend-ns \
  create secret generic \
  quick-start-backend-certs \
  --from-file=server.crt \
  --from-file=server.key
```
<pre>secret "quick-start-backend-certs" created</pre>

**3.** To create and save the **PostgreSQL StatefulSet** manifest in a file named **pg.yml** in your current working directory, run the command:

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
            value: quick_start_admin_user
          - name: POSTGRES_PASSWORD
            value: quick_start_admin_password
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
In the manifest above, the certificate files for your database server are mounted
in a volume with permissions `0600` (as indicated by the `defaultMode: 384`) and
the pod is deployed with `999` (which is the `postgres` group) associated with any volumes mounted into the pod (as indicated by `fsGroup: 999`).

**4.** To deploy the **PostgreSQL StatefulSet**, run the command:
```bash
kubectl --namespace quick-start-backend-ns apply -f pg.yml
```
<pre>
statefulset "pg" created
</pre>

This StatefulSet uses the [**postgres:9.6**](https://hub.docker.com/r/library/postgres/) container image available from DockerHub. When the container is started, the environment variables `POSTGRES_USER` and `POSTGRES_PASSWORD` are used to create a user with superuser power.

We'll treat these credentials as **admin-credentials** moving forward (i.e. `REMOTE_DB_ADMIN_USER` and `REMOTE_DB_ADMIN_PASSWORD` environment variables), to be used for administrative tasks - separate from **application-credentials**.

**5.** To ensure the **PostgreSQL StatefulSet** pod has started and is healthy, run the command:
```bash
kubectl --namespace quick-start-backend-ns get pods
```
<pre>
NAME      READY     STATUS    RESTARTS   AGE
pg-0      1/1       Running   0          6s
</pre>

#### Expose PostgreSQL Service

Now that the **PostgreSQL StatefulSet** is up and running, you will need to expose it publicly before you can consume it.

The service definition below assumes you're using minikube, where **NodePort** is the appropriate type of service to expose the application; you may prefer to use something different e.g. a **LoadBalancer** for a GKE cluster.

To expose the database, run the command:

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

If you used the service definition above, the database server should now be available at `$(minikube ip):30001` (referred to as `REMOTE_DB_URL`, moving forward).

The database has no content at this point, however you can validate that everything works by simply logging in as the admin-user. Run this command to list the roles in this DB - `psql` will be used to make a connection to the database using admin credentials:

```bash
export REMOTE_DB_ADMIN_USER=quick_start_admin_user
export REMOTE_DB_ADMIN_PASSWORD=quick_start_admin_password
export REMOTE_DB_URL=$(minikube ip):30001

docker run \
  --rm \
  -it \
  -e PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
  postgres:9.6 \
    psql \
    -U ${REMOTE_DB_ADMIN_USER} \
    "postgres://${REMOTE_DB_URL}/postgres" \
    -c "\du"
```
<pre>
                                          List of roles
       Role name        |                         Attributes                    
     | Member of
------------------------+-------------------------------------------------------
-----+-----------
 quick_start_admin_user | Superuser, Create role, Create DB, Replication, Bypass
 RLS | {}
</pre>

### Setup And Configure PostgreSQL Target Service

In this section, we assume that you have a PostgreSQL backend set up and ready to go. This means:

+ The PostgreSQL backend is publicly available via some URL. We'll refer to this public URL by the environment variable `REMOTE_DB_URL`
+ Admin-level database credentials exist to allow you to create the application user. We'll refer to them as the environment variables `REMOTE_DB_ADMIN_USER` and `REMOTE_DB_ADMIN_PASSWORD`

Note that if you are using your own database server for this tutorial and it is
not SSL-enabled, you will want to reference the [handler documentation](/docs/reference/handlers/postgres.html) to learn how to
modify your Secretless Broker configuration to disable SSL.

Please ensure the environment variables are set to reflect your environment. For example, if you followed along in the last section and are using minikube for your local K8s cluster, you can run:

``` bash
export REMOTE_DB_ADMIN_USER=quick_start_admin_user
export REMOTE_DB_ADMIN_PASSWORD=quick_start_admin_password
export REMOTE_DB_URL="$(minikube ip):30001"
```

You will setup and configure the PostgreSQL storage backend by carrying the following steps:
1. Create a dedicated application database (henceforth referred to by the environment variable `APPLICATION_DB_NAME`) within the PostgreSQL DBMS, then carry out the rest of the steps within its context
2. Create the application table (i.e. pets)
3. Create an application user with limited privileges: `SELECT` and `INSERT` on the application table
4. Store the application user's credentials (held in the environment variables `APPLICATION_DB_USER` and `APPLICATION_DB_INITIAL_PASSWORD`) in in a secret store (for the purposes of this demo, we're using Kubernetes secrets).

**Note:** You must set the value of and export the environment variables `APPLICATION_DB_NAME`, `APPLICATION_DB_USER` and `APPLICATION_DB_INITIAL_PASSWORD` before proceeding, e.g.
``` bash
export APPLICATION_DB_NAME=quick_start_db
export APPLICATION_DB_USER=app_user
export APPLICATION_DB_INITIAL_PASSWORD=app_user_password
```

To create the application database, application table, application user and grant the application user relevant privileges, run this command:

```bash
docker run \
  --rm \
  -i \
  -e PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
  postgres:9.6 \
    psql -U ${REMOTE_DB_ADMIN_USER} "postgres://${REMOTE_DB_URL}/postgres" \
    << EOSQL
/* Create Application Database */
CREATE DATABASE ${APPLICATION_DB_NAME};

/* Connect to Application Database */
\c ${APPLICATION_DB_NAME};

/* Create Application Table */
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
You are now connected to database "quick_start_db" as user "quick_start_admin_user".
CREATE TABLE
CREATE ROLE
GRANT
GRANT
</pre>

### Create Application Namespace and Store Application DB-Credentials

Now that the storage backend is setup and good to go, it's time to set up the namespace where the application will be deployed.

The application will be scoped to the **quick-start-application-ns** namespace.

Run this code to create the namespace:

```yaml
kubectl create namespace quick-start-application-ns
```

Now that the namespace is created, you will proceed to store the application-user credentials in Kubernetes secrets. This is better than hard-coding them - but in a real deployment you would want to store your secrets in a fully-featured vault, like Conjur.

Run this code to store application-user DB-credentials in Kubernetes secrets:

```bash
kubectl --namespace quick-start-application-ns \
 create secret generic \
 quick-start-backend-credentials \
 --from-literal=address="${REMOTE_DB_URL}" \
 --from-literal=username="${APPLICATION_DB_USER}" \
 --from-literal=password="${APPLICATION_DB_INITIAL_PASSWORD}"
```
<pre>
secret "quick-start-backend-credentials" created
</pre>

### Create Secretless Broker Configuration ConfigMap

At this point, we've provisioned our database, configured it to be accessed by the application, stored the access credentials in a secret store - so we're ready to write our Secretless Broker configuration that will define how the Secretless Broker should listen for connections to this PostgreSQL database and what it should do with those connection requests when it receives them.

Once we've written that configuration, we can hand it off for the developer as they prepare to deploy their application.

The Secretless Broker configuration file has 2 sections:
+ Listeners, to define how the Secretless Broker should listen for new connection requests for a particular backend
+ Handlers, which are passed new connection requests received by the listeners, and are the part of the Secretless Broker that actually opens up a connection to the target service with credentials that it retrieves using the specified credential provider(s)

In our sample, we create a **secretless.yml** file as follows:
+ Define a `pg` type listener named **pets-pg-listener** that listens on `localhost:5432`
+ Define a handler named **pets-pg-handler** that uses the `kubernetes` secrets provider to retrieve the `address`, `username` and `password` of the remote database. The `kubernetes` secrets provider is used to access Kubernetes secrets through the Kubernetes API - authenticating with the service account credentials available from within the pod, as described in [Accessing the Kubernetes API from a pod](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/#accessing-the-api-from-a-pod) .

This configuration is shared amongst all Secretless Broker sidecar containers, each residing within an application Pod replica.

Run the command below to create a Secretless Broker configuration file named **secretless.yml** in your current working directory:

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
Note: we don't specify an `sslmode` in the Secretless Broker config, so it will
use the default `require` value.

You will now create a ConfigMap from the **secretless.yml** file. Later this ConfigMap will be made available to each Secretless Broker sidecar container via a volume mount.

Create the Secretless Broker Configuration ConfigMap by running the command:
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

In Kubernetes, a service account provides an identity for processes that run in a Pod.

Let's set up an application service account to provide identity for the application. With identity you're able to grant fine-grained entitlements for the application pod to have access to the Kubernetes secrets holding the database credentials.

Run the command below to create a *quick-start-application* ServiceAccount:

```bash
kubectl --namespace quick-start-application-ns \
  create serviceaccount \
  quick-start-application
```
<pre>
serviceaccount "quick-start-application" created
</pre>

Now you need to grant [get] access to the **quick-start-backend-credentials** secret to this ServiceAccount. This will be a 2 step process:
1. Create a **Role** with permissions to `[get]` the *quick-start-backend-credentials* secret, named *quick-start-backend-credentials-reader*
2. Create a **RoleBinding** of the Role in the previous step and application ServiceAccount, named *read-quick-start-backend-credentials*

Run this command to grant the entitlements:
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
  apply \
  -f quick-start-application-entitlements.yml
```
<pre>
role "quick-start-backend-credentials-reader" created
rolebinding "read-quick-start-backend-credentials" created
</pre>

## Steps for the non-privileged user (i.e developer)

Close the terminal you've been using to run through all of the previous steps and open a new one for these next few. That terminal window had all of the database configuration stored as environment variables - but none of the steps in this section need any credentials at all. That is, the person deploying the application needs to know _nothing_ about the secret values required to connect to the PostgreSQL database!!

The only environment variable you will need for this next set of steps is `APPLICATION_DB_NAME`, and you can re-export that as:

```bash
export APPLICATION_DB_NAME=quick_start_db
```

### Build Application Deployment Manifest

In this section, you build the deployment manifest for the application. The manifest includes a section for specifying the pod template (`$.spec.template`), it is here that we will declare the application container and Secretless Broker sidecar container.

Below is the base manifest that you will be building upon:
_quick-start-application.yml_
```yaml
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
      # to be completed in the following steps
```

The manifest should be saved as a file named **quick-start-application.yml**. As you can see above, to start off the base manifest you declare a deployment of 3 replicas with associated metadata, and assign the *quick-start-application* ServiceAccount (that we created earlier) to the pod.

The additional steps to complete building the manifest are provided to be informative:
+ [Add & Configure Application Container](#add--configure-application-container)
+ [Add & Configure Secretless Broker Sidecar Container](#add--configure-secretless-broker-sidecar-container)

A [Completed Application Deployment Manifest](#completed-application-deployment-manifest) is provided in the last section, where you'll actually create the **quick-start-application.yml**.

#### Add & Configure Application Container

The Secretless Broker sidecar container has a shared network with the application container. This allows us to point the application to `localhost` where the Secretless Broker is listening on port `5432`, in accordance with the configuration stored in ConfigMap.

_Note_:
+ An application must connect to the Secretless Broker without SSL, though the actual connection between the Secretless Broker and the target service can leverage SSL. To achieve this we append the query parameters `sslmode=disable` to the connection string, which prevents the PostgreSQL driver from using SSL mode with the Secretless Broker.
+ The Secretless Broker respects the parameters specified in the database connections string.

You will now add the application container definition to the application deployment manifest. The application takes configuration from environment variables. Set the `DB_URL` environment variable to `postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable`, so that when the application is deployed it will open the connection to the PostgreSQL backend via the Secretless Broker.

Add the application container to the base manifest:

_quick-start-application.yml_
```yaml
# update the path $.spec.template.spec in the base manifest with the content below
containers:
  - name: quick-start-application
    image: cyberark/demo-app:latest
    env:
      - name: DB_URL
        # don't forget to substitute the actual value of ${APPLICATION_DB_NAME} below !!!
        value: postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
```

#### Add & Configure Secretless Broker Sidecar Container

You will now add the Secretless Broker sidecar container to the containers section under `spec` of the pod template. You will need to:

1. Add the Secretless Broker sidecar container definition
2. Add a read-only volume mount on the Secretless Broker sidecar container for the Secretless Broker configuration ConfigMap (`quick-start-application-secretless-config`)

_quick-start-application.yml_
```yaml
# update the path $.spec.template.spec in the base manifest with the content below
containers:
  - name: quick-start-application
    # already filled in from previous section
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

#### Completed Application Deployment Manifest

You should now save the application deployment manifest in a file named **quick-start-application.yml**.
Running the command below will save a completed deployment manifest to **quick-start-application.yml** in your current working directory, using the value of the `APPLICATION_DB_NAME` environment variable in the executing terminal:

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

You now have a complete application deployment manifest from the previous section, with 2 containers (the application and the Secretless Broker sidecar) defined in the Pod template. It is time to deploy the application using this manifest.

To deploy the application, run this command:
```bash
kubectl --namespace quick-start-application-ns apply -f quick-start-application.yml
```
<pre>
deployment "quick-start-application" created
</pre>

To ensure the application pods have started and are healthy, run this command:
```bash
kubectl --namespace quick-start-application-ns get pods
```
<pre>
NAME                                       READY     STATUS        RESTARTS   AGE
quick-start-application-6bd8dbd57f-bshmf   2/2       Running       0          22s
quick-start-application-6bd8dbd57f-dr962   2/2       Running       0          26s
quick-start-application-6bd8dbd57f-fgfnh   2/2       Running       0          30s
</pre>

#### Expose Application Publicly

Now that the application is up and running, you will need to expose it publicly before you can consume the web-service.

The service definition below assumes you're using minikube, where **NodePort** is the appropriate type of service to expose the application; you may prefer to use something different e.g. a **LoadBalancer** for a GKE cluster.

To expose the application, run the command:

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
 apply \
 -f quick-start-application-service.yml
```
<pre>
service "quick-start-application" created
</pre>

If you used the service definition above, the application should now be available at `$(minikube ip):30002`, (referred to as environment variable `APPLICATION_URL` from now on).

## Try The Running Application

That's it! You've configured your application to connect to PostgreSQL via the Secretless Broker, and we can try it out to validate that it's working as expected.

The next steps rely on the presence of the `APPLICATION_URL` environment variable. For example, if you used the service definition in the previous then you would setup your environment as follows:
```bash
export APPLICATION_URL=$(minikube ip):30002
```

Run the command below to create a pet (`POST /pet`):
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
We expect the request responds with a HTTP status 201.

Run the command below to retrieve all the pets (`GET /pets`):
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
We expect the request to respond with a JSON array containing all the pets.

There we have it. This application is communicating with a target service without managing any secrets.

## Rotate Target Service Credentials

In this section, you get to see how an application using the Secretless Broker deals with credential rotation.
These are the steps you wil take to rotate the credentials of the dabatase:
+ rotate application-user DB-credentials in the database
+ update the application-user DB-credentials in the vault
+ prune existing connections established using old credentials

### Poll Application API [separate terminal]
Before rotating, **you will run the commands in this section in a new and separate terminal** to poll the retrieve pets endpoint (`GET /pets`). This will allow you to see the request-response cycle of the application. If something goes wrong, like a database connection failure you will see it as a > 400 HTTP status code.

First, setup the environment with the `APPLICATION_URL` environment variable. If you're using `minikube`:
```bash
export APPLICATION_URL=$(minikube ip):30002
```

To start polling run this command:
```bash
while true
do
  echo "Retrieving pets at $(date):"
  curl -i ${APPLICATION_URL}/pets
  echo ""
  echo ""
  sleep 3
done
```
<pre>
Retrieving pets at Thu 23 Aug 2018 19:55:33 +08:
HTTP/1.1 200
Content-Type: application/json;charset=UTF-8
Transfer-Encoding: chunked
Date: Thu, 23 Aug 2018 11:55:33 GMT

[{"id":1,"name":"Secretlessly Fluffy"}]

...
</pre>

### Rotate Credentials

You will be using **admin-credentials** to carry out these steps, pruning existing connections requires elevated privileges.

Begin by setting up environment variables (assumes the default setup with `minikube`):

```bash
export REMOTE_DB_ADMIN_USER=quick_start_admin_user
export REMOTE_DB_ADMIN_PASSWORD=quick_start_admin_password
export REMOTE_DB_URL=$(minikube ip):30001

export APPLICATION_DB_USER=app_user
# you can specify a different value for the new password below
export APPLICATION_DB_NEW_PASSWORD=new_app_user_password
```

#### Rotate Credentials In DB
Now you can proceed to rotate the credentials in the database.

Remember, you will be using **admin-credentials** to carry out this step.

To rotate the application DB-user password, run this command:
```bash
docker run \
 --rm \
 -i \
 -e PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
 postgres:9.6 \
  psql \
  -U ${REMOTE_DB_ADMIN_USER} \
  "postgres://${REMOTE_DB_URL}/postgres" \
  << EOSQL
/* Rotate Application User password */
ALTER ROLE ${APPLICATION_DB_USER} WITH PASSWORD '${APPLICATION_DB_NEW_PASSWORD}';
EOSQL
```
<pre>
ALTER ROLE
</pre>

#### Update Credentials In Secret Store

After rotation the password value held in the secret store is stale and requires updating.
Run the following command to update the application-user DB-credentials password value in Kubernetes secrets:

```bash
base64_new_password=$(echo -n "${APPLICATION_DB_NEW_PASSWORD}" | base64)
new_password_json='{"data":{"password": "'${base64_new_password}'"}}'

kubectl --namespace quick-start-application-ns \
 patch secret \
 quick-start-backend-credentials \
 -p="${new_password_json}"
```
<pre>
secret "quick-start-backend-credentials" patched
</pre>

#### Prune Existing Connections In DB

You will also need to prune existing connections established using old credentials - this in itself has no noticeable effect on the application because most drivers keep a pool of connections and replenish them as and when needed.

Note that this step takes place after updating the credentials in the secret store. This ensures immediate attempts to reconnect after this step will use the latest credentials.   

Remember, you will be using **admin-credentials** to carry out this step.

To prune existing connections, run this command:
```bash
docker run \
 --rm \
 -i \
 -e PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
 postgres:9.6 \
  psql \
  -U ${REMOTE_DB_ADMIN_USER} \
  "postgres://${REMOTE_DB_URL}/postgres" \
  << EOSQL
/* Prune Existing Connections */
SELECT
  pg_terminate_backend(pid)
FROM
  pg_stat_activity
WHERE
  pid <> pg_backend_pid()
AND
  usename='${APPLICATION_DB_USER}';
EOSQL
```
<pre>
 pg_terminate_backend
----------------------
 t
 t
 t
 .
 .
 .
(30 rows)
</pre>

### Conclusion
Now return to the polling terminal. Observe that requests to the application API are not affected by the password rotation - we continue to be able to query the application as usual, without interruption!

## Review Complete Tutorial With Scripts

Check out [our tutorial on github](https://github.com/cyberark/secretless-broker/tree/master/demos/k8s-demo), complete with shell scripts for walking through the steps of the tutorial yourself and configurable to suit your needs.
=======
>>>>>>> b909749... Add redirect links for previously deleted files
