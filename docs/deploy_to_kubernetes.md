---
title: Secretless
id: deploy_to_kubernetes
layout: docs
description: Secretless Documentation
permalink: deploy_to_kubernetes
---

<p class="card-heading">Deploy to Kubernetes</p>

To get started with Secretless, try working through this tutorial, which goes through deploying Secretless with a sample application that uses a PostgresSQL database for backend storage.

We have chosen a PostgresSQL database as the target service for this tutorial, however the Secretless broker comes built-in with support for several other target services; check out our [reference](/reference.html) for more info.

## Table of Contents

+ [Getting Started](#getting-started)
+ [Sample Application](#sample-application)
+ Steps for the admin user
  + [Provision Target Services (optional)](#provision-target-services-optional)
  + [Setup And Configure Target Service](#setup-and-configure-target-service)
  + [Create Application Namespace and Store Application DB-Credentials](#create-application-namespace-and-store-application-db-credentials)
  + [Create Secretless Configuration ConfigMap](#create-secretless-configuration-configmap)
+ Steps for the non-privileged user (i.e developer)
  + [Build Application Deployment Manifest](#build-application-deployment-manifest)
    + [Add & Configure Application Container](#add--configure-application-container)
    + [Add & Configure Secretless Sidecar Container](#add--configure-secretless-sidecar-container)
    + [Completed Application Deployment Manifest](#completed-application-deployment-manifest)
  + [Deploy Application With Secretless](#deploy-application-with-secretless)
    + [Expose Application Publicly](#expose-application-publicly)
+ [Try The Running Application](#try-the-running-application)
+ [Rotate Target Service Credentials](#rotate-target-service-credentials)
+ [Review Complete Tutorial With Scripts](#review-complete-tutorial-with-scripts)

## Getting started

In this tutorial, we will walk through creating an application that communicates with a password-protected PostgreSQL database via the Secretless broker. _The application does not need to know anything about the credentials required to connect to the database;_ the admin super-user who provisions and configures the database will also configure Secretless to be able to communicate with it. The developer writing the application only needs to know the socket or address that Secretless is listening on to proxy the connection to the PostgreSQL backend.

If you'd rather just see the whole thing working end to end, check out [our tutorial on github](https://github.com/conjurinc/secretless/tree/master/demos/k8s-demo), complete with shell scripts for walking through the steps of the tutorial yourself and configurable to suit your needs.

We are going to do the following:

**As the admin super-user:**

1. Provision target services
1. Configure target services for usage by application and add credentials to secret store
1. Configure Secretless to broker the connection to the target service using credentials from the secret store

**As the application developer:**
1. Configure application to connect to target service through interface exposed by Secretless
1. Deploy and run Secretless adjacent to the application

### Prerequisites

To run through this tutorial, you will need:

+ A running Kubernetes cluster (you can use [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) to run a cluster locally)
+ [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured to point to the cluster
+ [Docker CLI](https://docs.docker.com/install/)

Our Kubernetes deployment manifests assume that you are using minikube, so we use `NodePort` to expose the services; you may prefer to use a `LoadBalancer` for a GKE cluster.

## Sample Application

The tutorial uses an existing [pet store demo application](https://github.com/conjurinc/pet-store-demo) that exposes the following routes:

- `GET /pets` to list all the pets in inventory
- `POST /pet` to add a pet
  - Requires `Content-Type: application/json` header and body that includes `name` data

There are additional routes that are also available, but these are the two that we will focus on for this tutorial.

Pet data is stored in a PostgreSQL database, and the application may be configured to connect to the database by setting the `DB_URL` environment variables in the application's environment (following [12-factor principles](https://12factor.net/)).

See [Try The Running Application](#try-the-running-application) for examples of consuming the routes using `curl`.

_main.go_
```go
db, err := sql.Open("postgres", os.Getenv("DB_URL"))
```

The application uses `DB_URL` to establish a database handle, `db` (representing a pool of connections) which allows the application to communicate with the database. 

The database is **credential-protected**. With Secretless, we will be able to use a value of `DB_URL` of the form:  `postgresql://x@localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable`. The application will not require knowledge of credentials to connect to the database.

## Steps for the admin user

The following steps would be taken by an admin-level user, who has the ability to create and configure a database and to add secret values to a secret store.

### Provision Target Services (optional)

If you already have a PostgreSQL server running and want to use that in this tutorial, please continue to [Setup And Configure Target Service](#setup-and-configure-target-service).

In this section, you will create the PostgresSQL storage backend in your Kubernetes cluster. 

Our PostgresSQL storage backend is stateful. For this reason, we'll use a StatefulSet to manage the backend. Please consult [the Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) to understand what a StatefulSet is and when it's appropriate to use it.

#### Deploy PostgresSQL StatefulSet

To deploy a PostgresSQL StatefulSet, follow these steps:

**1.** To create a dedicated namespace for the storage backend, run the command:
```
kubectl create namespace quick-start-db
```

**2.** To create and save the pg StatefulSet manifest in a file named `pg.yml` in your current working directory, run the command:

```
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
**3.** To deploy the PostgresSQL StatefulSet, run the command:
```bash
kubectl --namespace quick-start-db apply -f pg.yml
```

This StatefulSet uses the `postgres:9.6` container image available from DockerHub. When the container is started, the environment variables `POSTGRES_USER` and `POSTGRES_PASSWORD` are used to create a user with superuser power.

We'll treat these credentials as `admin-credentials` moving forward (i.e. `$REMOTE_DB_ADMIN_USER` and `$REMOTE_DB_ADMIN_PASSWORD` environment variables), to be used for administrative tasks - separate from `application-credentials`.

**3.** To ensure the `pg` StatefulSet pod has started and is healthy, run the command:
```bash
kubectl --namespace quick-start-db get pods
```

#### Expose PostgresSQL Service

Now that the `pg` StatefulSet is up and running, you will need to expose it publicly before you can consume it.

The service definition below assumes you're using minikube, where `NodePort` is the appropriate type of service to expose the application; you may prefer to use something different e.g. a `LoadBalancer` for a GKE cluster.
 
To expose the database, run the command:

```bash
cat << EOF | kubectl --namespace quick-start-db  apply -f -
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

If you used the service definition above, the application should now be available at `$(minikube ip):30001` (referred to as `$REMOTE_DB_URL`, moving forward).
 
The DB has no content at this point, however you can validate that everything works by simply logging in as the admin-user:

```bash
export REMOTE_DB_ADMIN_USER=quick_start_admin_user
export REMOTE_DB_ADMIN_PASSWORD=quick_start_admin_password
export REMOTE_DB_URL=$(minikube ip):30001

docker run \
  --rm -it \
  postgres:9.6 \
  env PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
  psql \
  -U ${REMOTE_DB_ADMIN_USER} \
  "postgres://${REMOTE_DB_URL}/postgres"
```

Note that the DB has no content at this point.

### Setup And Configure Target Service

In this section, we assume that you have a PostgresSQL backend set up and ready to go. Concretely this means:

+ The PostgresSQL backend is publicly available via some URL. We'll refer to this public URL by the environment variable `$REMOTE_DB_URL`
+ Admin-level database credentials exist to allow you to create the application user. We'll refer to them as the environment variables `$REMOTE_DB_ADMIN_USER` and `$REMOTE_DB_ADMIN_PASSWORD`

Please ensure the environment variable are set to reflect your environment. For example, if you followed along in the last section and are using minikube for your local K8s cluster, you can run:

``` bash
export REMOTE_DB_ADMIN_USER=quick_start_admin_user
export REMOTE_DB_ADMIN_PASSWORD=quick_start_admin_password
export REMOTE_DB_URL="$(minikube ip):30001"
```

You will setup and configure the PostgresSQL storage backend by carrying the following steps:
1. Create a dedicated application database (henceforth referred to by the environment variable `$APPLICATION_DB_NAME`) within the PostgresSQL DBMS, then carry out the rest of the steps within its context
2. Create the application table (i.e. pets)
3. Create an application user with limited privileges, `SELECT` and `INSERT` on the application table 
4. Store the application user's credentials (`$APPLICATION_DB_USER` and `$APPLICATION_DB_INITIAL_PASSWORD`) in in a secret store (for the purposes of this demo, we're using Kubernetes secrets). 

**Note:** You must set and export the value of `$APPLICATION_DB_NAME`, `$APPLICATION_DB_USER` and `$APPLICATION_DB_INITIAL_PASSWORD` before proceeding, e.g. 
``` bash
export APPLICATION_DB_NAME=quick_start_db
export APPLICATION_DB_USER=app_user
export APPLICATION_DB_INITIAL_PASSWORD=app_user_password
```

To create the application database, application table, application user and grant the application user relevant privileges, run this command:

```bash
docker run \
  --rm -i \
  postgres:9.6 \
  env PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
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

### Create Application Namespace and Store Application DB-Credentials

Now that the storage backend is setup and good to go, it's time to set up the namespace where the application will be deployed.

The first step is to decide on an application namespace. This is the namespace in which the application will be scoped. You can pick your own by setting the environment variable `APPLICATION_NAMESPACE`:
```bash
export APPLICATION_NAMESPACE=quick-start-app
```

Run this code to create the namespace:

```yaml
kubectl create namespace ${APPLICATION_NAMESPACE}
```

Now that the namespace is created, you will proceed to store the application-user credentials in Kubernetes secrets. This is better than hard-coding them - but in a real deployment you would want to store your secrets in an actual vault, like Conjur.

Run this code to store application-user DB-credentials in Kubernetes secrets:

```bash
cat << EOF | kubectl --namespace ${APPLICATION_NAMESPACE} apply -f -
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

At this point, we've provisioned our database, configured it to be accessed by the application, stored the access credentials in a secret store - so we're ready to write our Secretless configuration that will define how Secretless should listen for connections to this PostgreSQL database and what it should do with those connection requests when it receives them.

Once we've written that configuration, we can hand it off for the developer as they prepare to deploy their application.

The Secretless configuration file has 2 sections:
+ Listeners, to define how Secretless should listen for new connection requests for a particular backend
+ Handlers, which are passed new connection requests received by the listeners, and are the part of Secretless that actually opens up a connection to the target service with credentials that it retrieves using the specified credential provider(s)

In our sample, we create a `secretless.yml` file as follows:
+ Define a `pg` listener named `pets-pg-listener` that listens on `localhost:5432`
+ Define a handler named `pets-pg-handler` that uses the `file` provider to retrieve the `address`, `username` and `password` of the remote database. The file provider is used to access Kubernetes secrets made available to the Secretless container via volume mounts. 

  _NOTE: There is a pending feature for a Kubernetes secret provider which will retrieve secrets directly from the Kubernetes API_.
  
This configuration is shared amongst all Secretless sidecar containers, each residing within an application Pod replica. 

Run the command below to create a Secretless configuration file named `secretless.yml` in your current working directory:

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
        provider: file
        id: /etc/secret/address
      - name: username
        provider: file
        id: /etc/secret/username
      - name: password
        provider: file
        id: /etc/secret/password
EOF
```

You will now create a ConfigMap from the `secretless.yml` file. Later this ConfigMap will be made available to each Secretless sidecar container via a volume mount.

Create the Secretless Configuration ConfigMap by running the command:
```bash
kubectl --namespace ${APPLICATION_NAMESPACE} \
  create configmap \
  quick-start-application-secretless-config \
  --from-file=secretless.yml
```

## Steps for the non-privileged user (i.e developer)

Close the terminal you've been using to run through all of the previous steps and open a new one for these next few. That terminal window had all of the database configuration stored as environment variables - but none of the steps in this section need any credentials at all. That is, the person deploying the application needs to know _nothing_ about the secret values required to connect to the PostgreSQL database!!

The only environment variables you will need for this next set of steps are the `$APPLICATION_NAMESPACE` and `$APPLICATION_DB_NAME`, and you can re-export them as:

```bash
export APPLICATION_NAMESPACE=quick-start-app
export APPLICATION_DB_NAME=quick_start_db
```

### Build Application Deployment Manifest

In this section, you build the deployment manifest for the application. The manifest includes a section for specifying the pod template (`$.spec.template`), it is here that we will declare the application container and Secretless sidecar container.

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

The additional steps to complete building the manifest are:
+ [Add & Configure Application Container](#add--configure-application-container)
+ [Add & Configure Secretless Sidecar Container](#add--configure-secretless-sidecar-container)

#### Add & Configure Application Container

The secretless broker sidecar container has a shared network with the application container. This allows us to point the application to `localhost` where Secretless is listening on port `5432`, in accordance with the configuration stored in ConfigMap.

_Note_: 
+ An application must connect to Secretless without SSL, though the actual connection between Secretless and the target service can leverage SSL. To achieve this we append the query parameters `sslmode=disable` to the connection string, which prevents the PostgresSQL driver from using SSL mode with Secretless.
+ Some database drivers require the connection string to explicitly specify a user. This is the reason for `..x@localhost...` in `$DB_URL` below.
+ Secretless respects the parameters specified in the database connections string.

You will now add the application container definition to the application deployment manifest. The application takes configuration from environment variables. Set the `$DB_URL` environment variable to `postgresql://x@localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable`, so that when the application is deployed it will open the connection to the PostgreSQL backend via Secretless.

Add the application container to the base manifest:

_quick-start.yml_
```yaml
# update the path $.spec.template.spec in the base manifest with the content below
containers:
  - name: quick-start-application
    image: codebykumbi/pet-store:latest
    env:
      - name: DB_URL
        # don't forget to substitute the actual value of ${APPLICATION_DB_NAME} below !!!
        value: postgresql://x@localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
```

#### Add & Configure Secretless Sidecar Container

You will now add the Secretless sidecar container to the containers section under `spec` of the pod template. You will need to:

1. Add the Secretless sidecar container definition
2. Add read-only volume mounts on the Secretless sidecar container for:
  + Secretless configuration ConfigMap (`quick-start-application-secretless-config`)
  + Kubernetes secrets containing the application-user DB-credentials (`quick-start-backend-credentials`) 

_quick-start.yml_
```yaml
# update the path $.spec.template.spec in the base manifest with the content below
containers:
  - name: quick-start-application
    # already filled in from previous section
  - name: secretless
    image: cyberark/secretless:latest
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

#### Completed Application Deployment Manifest

You should now save the application deployment manifest in a file named `quick-start.yml`.
Running the command below will save a completed deployment manifest to `quick-start.yml` in your current working directory, using the value of the `$APPLICATION_DB_NAME` environment variable in the executing terminal:

```bash
cat << EOF > quick-start.yml
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
        - name: quick-start-application
          image: codebykumbi/pet-store:latest
          env:
            - name: DB_URL
              value: postgresql://x@localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
        - name: secretless
          image: cyberark/secretless:latest
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
EOF
```

### Deploy Application With Secretless

You now have a complete application deployment manifest from the previous section, with 2 containers (the application and the Secretless sidecar) defined in the Pod template. It is time to deploy the application using this manifest.

To deploy the application, run this command:
```bash
kubectl --namespace ${APPLICATION_NAMESPACE} apply -f quick-start.yml
```

To ensure the application pods have started and are healthy, run this command:
```bash
kubectl --namespace ${APPLICATION_NAMESPACE} get pods
```

#### Expose Application Publicly

Now that the application is up and running, you will need to expose it publicly before you can consume the web-service.

The service definition below assumes you're using minikube, where `NodePort` is the appropriate type of service to expose the application; you may prefer to use something different e.g. a `LoadBalancer` for a GKE cluster.
 
To expose the application, run the command:

```bash
cat << EOF | kubectl --namespace ${APPLICATION_NAMESPACE} apply -f -
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

If you used the service definition above, the application should now be available at `$(minikube ip):30002`, (referred to as `$APPLICATION_URL` from now on).

## Try The Running Application

That's it! You've configured your application to connect to PostgreSQL via Secretless, and we can try it out to validate that it's working as expected.

The next steps rely on the presence of `$APPLICATION_URL` in your environment. For example, if you used the service definition in the previous then you would setup your environment as follows:
```bash
export APPLICATION_URL=$(minikube ip):30002
```

Run the command below to create a pet:
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
We expect the command above to respond with HTTP status 201.

Run the command below to retrieve all the pets:
```bash
curl -i ${APPLICATION_URL}/pets
```
We expect the command above to respond with a JSON array containing all the pets.

There we have it. This application is communicating with a target service without managing any secrets.

## Rotate Target Service Credentials

In this section, you get to see how an application using Secretless deals with credential rotation.
These are the steps you wil take to rotate the credentials of the dabatase:
+ update the application-user DB-credentials in the vault
+ wait for the update to take effect
+ rotate the credentials in the database, and prune old connections

Typically, you'd want your rotation to work the other way - update the DB and then your vault - but we're using kubernetes secrets in this guide, which isn't built to handle secret rotation gracefully. For that, you'd want to use a better secrets management solution.

### Poll Application API [separate terminal]
Before rotating, **you will run the commands in this section in a new and separate terminal** to poll the retrieve pets endpoint (GET `/pets`). This will allow you to see the request-response cycle of the application. If something goes wrong, like a database connection failure you will see it as a > 400 HTTP status code.

First, setup the environment with `$APPLICATION_URL`. If you're using `minikube`:
```bash
export APPLICATION_URL=$(minikube ip):30002
```

To start polling run this command:
```bash
while true
do 
  echo "Retrieving pets at $(date):"
  curl $APPLICATION_URL/pets
  echo ""
  echo ""
  sleep 3
done
```

### Rotate Credentials

You will be using admin-credentials to carry out these steps, pruning existing connections requires elevated privileges.

Begin by setting up environment variables (assumes the default setup with `minikube`):

```bash
export REMOTE_DB_ADMIN_USER=quick_start_admin_user
export REMOTE_DB_ADMIN_PASSWORD=quick_start_admin_password
export REMOTE_DB_URL=$(minikube ip):30001

export APPLICATION_NAMESPACE=quick-start-app
export APPLICATION_DB_USER=app_user
# you can specify a different value for the new password below
export APPLICATION_DB_NEW_PASSWORD=new_app_user_password
```

#### Rotate Credentials In Secret Store
Let us now proceed to rotation.

Run the following command to update the application-user DB-credentials in Kubernetes secrets:

```bash
cat << EOF | kubectl --namespace ${APPLICATION_NAMESPACE} apply -f -
---
apiVersion: v1
kind: Secret
metadata:
  name: quick-start-backend-credentials
type: Opaque
data:
  address: $(echo -n "${REMOTE_DB_URL}" | base64)
  username: $(echo -n "${APPLICATION_DB_USER}" | base64)
  password: $(echo -n "${APPLICATION_DB_NEW_PASSWORD}" | base64)
EOF
```

There will be a lag for these credentials to propagate to the volume mount of the application under `/etc/secret`.

You can check that the credential have been propagate by `exec`ing into any one of the application pods and comparing the contents of `/etc/secret/password` against `APPLICATION_DB_NEW_PASSWORD`.

Run the following command to wait for Kubernetes secrets to propagate to the application pod Secretless sidecar container volume mounts:

```bash
function first_pod {
  kubectl --namespace ${APPLICATION_NAMESPACE} \
    get pods \
    -l=app="quick-start-application" \
    -o=jsonpath='{$.items[0].metadata.name}'
}

function first_pod_password {
  kubectl --namespace ${APPLICATION_NAMESPACE} \
    exec \
    -it \
    -c secretless \
    $(first_pod) \
    -- \
    cat /etc/secret/password
}

function wait_for_secret_propagation {
  while [[ ! "$(first_pod_password)" == "${APPLICATION_DB_NEW_PASSWORD}" ]]
  do
    echo "Waiting for secret to be propagated"
    sleep 10
  done

  echo "Ready!"
}

wait_for_secret_propagation
```

#### Rotate Credentials In DB
Now you can proceed to rotate the credentials in the database. 

You will also need to prune existing connections established using old credentials - this in itself has no noticeable effect on the application because most drivers keep a pool of connections and replenish them as and when needed. 

Remember, you will be using admin-credentials to carry out these steps, pruning existing connections requires elevated privileges.

To rotate the application DB-user password and prune existing connections, run this command:
```bash
docker run \
  --rm -it \
  postgres:9.6 \
  env PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} \
  psql -U ${REMOTE_DB_ADMIN_USER} "postgres://${REMOTE_DB_URL}/postgres" \
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

Check out [our tutorial on github](https://github.com/conjurinc/secretless/tree/master/demos/k8s-demo), complete with shell scripts for walking through the steps of the tutorial yourself and configurable to suit your needs.
