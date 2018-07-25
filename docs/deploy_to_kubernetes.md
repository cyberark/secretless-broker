---
title: Secretless
id: deploy_to_kubernetes
layout: docs
description: Secretless Documentation
permalink: deploy_to_kubernetes
---

<p class="card-heading">Deploy to Kubernetes</p>

To get started with Secretless, try working through this tutorial, which goes through deploying Secretless with a sample application that uses a Postgres database for backend storage.

We have chosen a Postgres database as the protected resource for this tutorial, however the Secretless broker comes built-in with support for several other target services; check out our [reference page](secretless.io/references) for more info.

## Table of Contents

+ [Getting Started](#getting-started)
+ [Sample Application](#about-the-sample-application)
+ Steps for the admin user
    + [Provision Protected Resources (optional)](#provision-protected-resources-optional)
    + [Setup And Configure Protected Resources](#setup-and-configure-protected-resources)
    + [Create Application Namespace and Store Application DB-Credentials](#create-application-namespace-and-store-application-user-credentials)
    + [Create Secretless Configuration ConfigMap](#create-secretless-configuration-configmap)
+ Steps for the non-privileged user (i.e developer)
    + [Deploy Application With Secretless](#deploy-application-with-secretless)
    + [Add and Configure Secretless sidecar container](#add-and-configure-secretless-sidecar-container)
+ [Consume Application](#consume-application)
+ [Rotate Protected Resource Credentials](#rotate-protected-resource-credentials)
+ [Review Complete Tutorial With Scripts](#review-complete-sample-repository)

## Getting started

In this tutorial, we will walk through creating an application that communicates with a password-protected PostgreSQL database via the Secretless broker. _The application does not need to know anything about the credentials required to connect to the database;_ the admin super-user who provisions and configures the database will also configure Secretless to be able to communicate with it. The developer writing the application only needs to know the socket or address that Secretless is listening on to proxy the connection to the PostgreSQL backend.

If you'd rather just see the whole thing working end to end, check out [our tutorial on github](https://github.com/conjurinc/secretless/tree/master/demos/k8s-demo), complete with shell scripts for walking through the steps of the tutorial yourself and configurable to suite your needs.

We are going to do the following:

**As the admin super-user:**

1. Provision protected resources
1. Configure protected resources for usage by application and add credentials to secret store
1. Configure Secretless to broker connection using credentials from the secret store

**As the application developer:**
1. Configure application to connect to protected resource through interface exposed by Secretless
1. Deploy and run Secretless adjacent to the application

### Prerequisites

To run through this tutorial, you will need:

+ A running Kubernetes cluster (you can use [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) to run a cluster locally)
+ [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured to point to the cluster
+ [Docker CLI](https://docs.docker.com/install/)

#### Conditional Changes To Tutorial

+ Our Kubernetes deployment manifests assume that you are using minikube, so we use `NodePort` to expose the services; you may prefer to use a `LoadBalancer` for a GKE cluster.

## Sample Application

In this tutorial, we use a [simple note storage-and-retrieval application](https://github.com/conjurinc/secretless/tree/master/demos/k8s-demo/app) written in Go. The application is an **http service** that uses a **Postgres storage backend**. It exposes the following routes:

- `GET /note` to retrieve notes
- `POST /note` to add a note
  - Requires `Content-Type: application/json` header and JSON body that includes `title` and `description` data

See [Consume Application](#consume-application) for examples of consuming the routes using `curl`.

The application is configured to connect to the database by setting the `DATABASE_URL` environment variables in the application's environment (following [12-factor principles](https://12factor.net/)).

_main.go_
```go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

The application uses `DATABASE_URL` to establish a database handle, `db` (representing a pool of connections) which allows the application to communicate with the database. 

The database is **credential-protected**. With Secretless, we will be able to use a value of `DATABASE_URL` of the form:  `postgresql://localhost:5432/[db_name]`. The application will not require knowledge of credentials to connect to the database.

## Steps for the admin user

The following steps would be taken by an admin-level user, who has the ability to create and configure a database and to add secret values to a secret store.

### Provision Protected Resources (optional)

If you already have a PostgreSQL server running and want to use that in this tutorial, please continue to [Setup and Configure protected resources](#setup-and-configure-protected-resources).

In this section, you will create the Postgres storage backend in your Kubernetes cluster. 

Our Postgres storage backend is stateful. For this reason, we'll use a StatefulSet to manage the backend. Please consult [the Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) tto understand what a StatefulSet is and when it's appropriate to use it.

To deploy a Postgres StatefulSet, follow these steps:
1. To create a dedicated namespace for the storage backend, run the command:
    ```bash
    $ kubectl create namespace quick-start-db
    ```
2. To deploy the Postgres StatefulSet, run the command:

    ```
    $ cat << EOF | kubectl apply --namespace quick-start-db -f -
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
    EOF
    ```
   This StatefulSet uses the `postgres:9.6` container image available from DockerHub. When the container is started, the environment variables `POSTGRES_USER` and `POSTGRES_PASSWORD` are used to create a user with superuser power and a database with the same name.
   
   We'll treat these credentials as `admin-credentials` moving forward (i.e. `REMOTE_DB_ADMIN_USER` and `REMOTE_DB_ADMIN_PASSWORD`), to be used for administrative tasks - separate from `application-credentials`.
3. To ensure the `pg` StatefulSet pod has started and is healthy, run the command:
    ```bash
    kubectl --namespace quick-start-db get pods
    ```

Now that the `pg` StatefulSet is up and running, you will need to expose it publicly before you can consume it.

The service definition below assumes you're using minikube, where `Nodeport` is the appropriate type of service to expose the application; you may prefer to use something different e.g. a `LoadBalancer` for a GKE cluster.
 
To expose the database, run the command:

```bash
$ cat << EOF | kubectl apply --namespace quick-start-db -f -
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
```

If you used the service definition above, the application should now be available at `$(mikube ip):30001` (referred to as `$REMOTE_DB_URL`, moving forward). You can validate that this works by logging in as the admin-user:

```bash
$ docker run \
  --rm -i \
  postgres:9.6 \
  env PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
  psql -U ${REMOTE_DB_ADMIN_USER} "postgres://${REMOTE_DB_URL}"
```
 

### Setup and Configure protected resources

Before proceeding, in this section, we assume that you have a Postgres backend set up and ready to go. Concretely this means:

+ The Postgres backend is publicly available via some URL. We'll refer to this public URL by the environment variable `$REMOTE_DB_URL`
+ Admin-level database credentials exist to allow you to create the application user. We'll refer to them as the environment variables `$REMOTE_DB_ADMIN_USER` and `$REMOTE_DB_ADMIN_PASSWORD`

Please ensure the environment variable are set to reflect your environment. For example, if you followed along in the last section and are using minikube for your local K8s cluster, you can run:

``` bash
$ export REMOTE_DB_ADMIN_USER=quick_start_admin_user
$ export REMOTE_DB_ADMIN_PASSWORD=quick_start_admin_password
$ export REMOTE_DB_URL="$(minikube ip):30001"
```

In the section that follows, you will setup and configure the Postgres storage backend by carrying the following steps:
1. Create a dedicated application database (henceforth referred to by the environment variable `$APPLICATION_DB_NAME`) within the Postgres DBMS, then carry out the rest of the steps within its context
2. Create the application table (i.e. notes)
3. Create an application user with limited privileges, `SELECT` and `INSERT` on the application table 
4. Store the application user's credentials (`$APPLICATION_DB_USER` and `$APPLICATION_DB_INITIAL_PASSWORD`) in Kubernetes secrets. 

**Note:** You must set and export the value of `$APPLICATION_DB_NAME`, `$APPLICATION_DB_USER` and `$APPLICATION_DB_INITIAL_PASSWORD` before proceeding e.g. 
``` bash
export APPLICATION_DB_NAME=quick_start_db
export APPLICATION_DB_USER=app_user
export APPLICATION_DB_INITIAL_PASSWORD=app_user_password
```

To create the application database, application table, application user and grant the application user relevant privileges, run this command:

```bash
$ docker run \
  --rm -i \
  postgres:9.6 \
  env PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
  psql -U ${REMOTE_DB_ADMIN_USER} "postgres://${REMOTE_DB_URL}" - \
  << EOSQL
    /* Create Application Database */
    CREATE DATABASE ${APPLICATION_DB_NAME};
 
    /* Connect to Application Database */
    \c ${APPLICATION_DB_NAME};

    /* Create Application Table */
    CREATE TABLE notes (
        id serial primary key,
        title varchar(256),
        description varchar(1024)
    );

    /* Create Application User */
    CREATE USER ${APPLICATION_DB_USER} PASSWORD '${APPLICATION_DB_INITIAL_PASSWORD}';
 
    /* Grant Permissions */
    GRANT SELECT, INSERT ON public.notes TO ${APPLICATION_DB_USER};
EOSQL
```

### Create Application Namespace and Store Application DB-Credentials

Now that the storage backend is setup and good to go, it's time to deploy the application with secretless.

The first step is to decide on an application namespace. This is the namespace in which the application will be scoped. You can pick your own. We'll refer to this namespace as `$APPLICATION_NAMESPACE` from now on.

Run this code to create the namespace:

```yaml
$ kubectl create namespace ${APPLICATION_NAMESPACE}
```

Now that the namespace is created, you will proceed to store the application-user credentials in Kubernetes secrets. Anything but hardcoding them :)

Run this code to store application-user credentials in Kubernetes secrets:

```bash
$ cat << EOF | kubectl apply --namespace ${APPLICATION_NAMESPACE} -f -
---
apiVersion: v1
kind: Secret
metadata:
    name: quick-start-backend-credentials
type: Opaque
data:
    address: $(echo -n "${REMOTE_DB_URL}" | base64)
    username: $(echo -n "${APPLICATION_DB_USER}" | base64)
    password: $(echo -n "${APPLICATION_DB_INITIAL_PASSWORD}" | base64)
EOF
```

### Create Secretless Configuration ConfigMap

In this section you will create the configuration required by the Secretless broker to make a connection to the protected resource and expose local interfaces.

There are 3 steps to configuring the Secretless broker for usage by the application
+ Declare listeners, which provide local interfaces to protected resources
+ Declare handlers, which handle connections made to listeners
    + broker the connection to the relevant protected resource
    + use providers to retrieve credentials necessary to authenticate against the protected resource
    
In our sample, we create a `secretless.yml` file as follows:
+ Declare a `pg` listener named `pg` that listens on `localhost:5432`
+ Declare a handler named `pg` that uses the `file` provider to retrieve the `address`, `username` and `password` of the remote database. The file provider is used to access kubernetes made available to the secretless container via volume mounts. 

  _NOTE: There is a pending feature for a Kubernetes secret provider which will retrieve secrets directly from the Kubernetes API_.

_secretless.yml_
```yaml
listeners:
  - name: pg
    protocol: pg
    address: localhost:5432
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

This configuration should be saved in a file named `secretless.yml`. The configuration is shared amongst secretless sidecar containers for each application pod replica. A ConfigMap is created from the file and later will be exposed to the secretless sidecar container via a volume mount.

Run this code to create the config map:
```bash
$ kubectl create configmap quick-start-application-secretless-config \
  --namespace ${APPLICATION_NAMESPACE} \
  --from-file=secretless.yml
```

## Steps for the non-privileged user (i.e developer)

None of the steps in this section require admin or application credentials - the person deploying the application needs to know _nothing_ about the secret values required to connect to the PostgreSQL database!!

### Deploy Application With Secretless

In this section, you create the deployment manifest for the application. The manifest includes a section for specifying the pod template (`$.spec.template`), it is here that we will declare the application container and Secretless sidecar container.

Below is the base manifest that you will be building upon: 
_quick-start.yml_
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
      containers:
      # to be completed in the following steps
```

The manifest should be saved as a file named `quick-start.yml`. As you can see above, to start off the base manifest you declare a deployment of 3 replicas with associated metadata. 

The additional steps to build the manifest are as follows:
+ [Add And Configure Secretless Sidecar Container](#add-and-configure-secretless-sidecar-container)
+ [Add And Configure Application Container](#add-and-configure-application-container)

#### Add and Configure Secretless Sidecar Container

In this section you will be adding the Secretless sidecar container to the containers section under `spec` of the pod template. You will need to:

1. Add the Secretless sidecar container definition
2. Add read-only volume mounts on the Secretless sidecar container for:
    + Secretless configuration ConfigMap (`quick-start-application-secretless-config`)
    + Kubernetes secrets containing the application-user DB-credentials (`quick-start-backend-credentials`) 

_quick-start.yml_
```yaml
# add content below to path $.spec.template.spec in base manifest
containers:
  - name: quick-start-application
    # leave this section blank for now
  - name: secretless
    image: secretless:latest
    imagePullPolicy: IfNotPresent
    args: ["-f", "/etc/secretless/secretless.yml"]
    volumeMounts:
    - name: secret
      mountPath: /etc/secret
      readOnly: true
    - name: config
      mountPath: /etc/secretless
      readOnly: true
   volumes:
    - name: secret
      secret:
        secretName: quick-start-backend-credentials
    - name: config
       configMap:
        name: quick-start-application-secretless-config
```

#### Add and Configure Application Container

The secretless broker sidecar container has a shared network with the application container. This allows us to point the application to `localhost` where Secretless is listening on port `5432`, in accordance with the configuration stored in ConfigMap.

In this section you add the application container definition to the application deployment manifest. The application takes configuration from environment variables. You will set `DATABASE_URL` to point to `postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable`, so that when the application is deployed it will open the connection to the PostgreSQL backend via Secretless.

**Note:** An application must connect to Secretless without SSL, though the actual connection between Secretless and the protected resouece can leverage SSL. We include `sslmode=disable` in the connection string to prevent the Postgres driver from using SSL mode with Secretless. Secretless respects the parameters specified in the database connections string.

Ultimately, the definition for the application container looks as follows:
```yaml
# add content below to path $.spec.template.spec.containers in base manifest
containers:
  - name: quick-start-application
    image: codebykumbi/note-store-app:latest
    env:
      - name: DATABASE_URL
        value: postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
  - name: secretless
    # ... already filled in from above

```

### Deploy Application With Secretless

It is now time to deploy the application using the manifest built in the previous section.

To deploy the application, run this command:
```bash
$ kubectl apply --namespace ${APPLICATION_NAMESPACE} -f quick-start.yml
```

To ensure the application pods have started and are healthy, run this command:
```bash
$ kubectl --namespace ${APPLICATION_NAMESPACE} get po
```

#### Expose Application Publicly

Now that the application is up and running, you will need to expose it publicly before you can consume the web-service.

The service definition below assumes you're using minikube, where `Nodeport` is the appropriate type of service to expose the application; you may prefer to use something different e.g. a `LoadBalancer` for a GKE cluster.
 
To expose the application, run the command:

```bash
$ cat << EOF | kubectl apply --namespace ${APPLICATION_NAMESPACE} -f -
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
```

If you used the service definition above, the application should now be available at `$(mikube ip):30002`, (reffered to as `$APPLICATION_URL` from now on).

## Consume Application

That's it! You've configured your application to connect to PostgreSQL via Secretless, and we can try it out to validate that it's working as expected.

Run the command below to create a note:
```bash
$ curl \
 -i \
 -d '{"title":"Secretless release", "description":"Once the tutorials are uploaded, initiate the release!"}' \
 -H "Content-Type: application/json" \
 ${APPLICATION_URL}/note
```
We expect the command above to respond with HTTP status 201.

Run the command below to retrieve all the notes:
```bash
$ curl -i ${APPLICATION_URL}/note
```
We expect the command above to respond with a JSON array containing the previously created note.

There we have it. This application is communicating with a protected resource without managing any secrets.

## Rotate Protected Resource Credentials

In this section, you get to see how an application using Secretless deals with credential rotation.
These are the steps you wil take to rotate the credentials of the dabatase:
+ update the application-user DB-credentials in the vault
+ wait for the update to take effect
+ rotate the credentials in the database, and prune old connections

Typically, you'd want your rotation to work the other way - update the DB and then your vault - but we're using kubernetes secrets in this guide, which isn't built to handle secret rotation gracefully. For that, you'd want to use a better secrets management solution.

### Poll Application API
Before rotating, you will run the command below in a new terminal to poll the retrieve notes endpoint (GET `/note`). This will allow you to see the request-response cycle of the application. If something goes wrong, like a database connection failure you will see it as a > 400 HTTP status code.

```bash
$ while true
do 
    echo "Retrieving notes"
    curl $APPLICATION_URL/note
    echo ""
    sleep 1
done
```

### Rotate Credentials In Secret Store
Let us now proceed to rotation.

Run the following command to update the application-user db credentials in Kubernetes secrets:

```bash
$ cat << EOF | kubectl apply -f -
---
apiVersion: v1
kind: Secret
metadata:
    name: quick-start-backend-credentials
type: Opaque
data:
    address: $(echo -n ${REMOTE_DB_URL} | base64)
    username: $(echo -n ${APPLICATION_DB_USER} | base64)
    password: $(echo -n ${APPLICATION_DB_NEW_PASSWORD}" | base64)
EOF
```

There will be a lag for these credentials to propagate to the volume mount of the application under `/etc/secret`.

You can check that the credential have been propagate by `exec`ing into any one of the application pods and comparing the contents of `/etc/secret/password` against `APPLICATION_DB_NEW_PASSWORD`.

Run the following command to wait for Kubernetes secrets to propagate to the application pod Secretless sidecar container volume mounts:

```bash
application_first_pod = $(kubectl get --namespace ${APPLICATION_NAMESPACE} po -l=app="quick-start-application" -o=jsonpath='{$.items[0].metadata.name}')

while [[ ! "$(kubectl --namespace ${APPLICATION_NAMESPACE} exec -it ${application_first_pod} -c secretless -- cat /etc/secret/password)" == "${APPLICATION_DB_NEW_PASSWORD}" ]] ; do
    echo "Waiting for secret to be propagated"
    sleep 10
done

echo Ready!
```

### Rotate Credentils In DB
Now you can proceed to rotate the credentials in the database. 

You will also need to prune existing connections established using old credentials - this in itself has no noticeable effect on the application because most drivers keep a pool of connections and replenish them as and when needed. 

You will be using admin-credentials to carry out these steps. To prune existing connections requires higher privileges.

```bash
$ docker run \
  --rm -it \
  postgres:9.6 \
  env PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
  psql -U ${REMOTE_DB_ADMIN_USER} "postgres://${REMOTE_DB_URL}" \
  -c "
    /* Rotate Application User password */
    ALTER ROLE ${APPLICATION_DB_USER} WITH PASSWORD '${APPLICATION_DB_NEW_PASSWORD}';
    
    /* Prune Existing Connections */
    SELECT
        pg_terminate_backend(pid)
    FROM
        pg_stat_activity
    WHERE
        pid <> pg_backend_pid()
    AND
        usename='${APPLICATION_DB_USER}';
"
```

### Conclusion
Now return to the polling terminal. Observe that requests to the application API are not affected by the password rotation - we continue to be able to query the application as usual, without interruption!.

## Review Complete Tutorial With Scripts

Check out [our tutorial on github](https://github.com/conjurinc/secretless/tree/master/demos/k8s-demo), complete with shell scripts for walking through the steps of the tutorial yourself and configurable to suite your needs.
