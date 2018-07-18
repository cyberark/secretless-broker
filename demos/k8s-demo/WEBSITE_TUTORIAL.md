# Introduction to Secretless on Kubernetes (Tutorial)

To get started with Secretless, try working through this tutorial, which goes through deploying Secretless with a sample application that uses a Postgres database for backend storage.

We've chosen a Postgres database as the protected resource for this sample, but you have many more options for protected resources such as other databases, web services, SSH connections, or any other TCP-based service.

## Introduction to Secretless on Kubernetes

The following sections of this tutorial detail the steps required to configure and deploy the sample. If you'd rather just run the code and inspect it, jump to [Review complete sample repository](#review-complete-sample-repository).

This tutorial assumes you have a running Kubernetes cluster and access to it via `kubectl`.

We also make use of `docker` to avoid asking you to install utility software on your local machine such as `psql`.

### Sample application

For this sample, we use a [simple note storage-and-retrieval application written in Go](https://github.com/conjurinc/secretless/tree/master/demos/k8s-demo/app). The application is an `http service` that uses a `Postgres storage backend`. It follows the [12-factor principles](https://12factor.net/config) to store configuration in the environment. Specifically, the `DATABASE_URL` environment variable is used:

_main.go_
```go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

`DATABASE_URL` is used by the application to establish a database handle, `db` (representing a pool of connections) which allows the application to communicate with the database.

#### Application server routes

For simplicity, the application only exposes the `C` and `R` of the well-known `CRUD` operations. 

+ GET `/note` to retrieve notes
+ POST `/note` to add a note - title and description must be specified via JSON body

For a deep dive into how the routes are handled please consult [the source code](https://github.com/conjurinc/secretless/blob/master/demos/k8s-demo/app/note.go).

### Provision protected resources

In this section, you will create the Postgres storage backend in your Kubernetes cluster. 

NOTE: It is possible to skip this step and leverage a remote Postgres storage backend. Continue to [Configure protected resources](#configure-protected-resources)

Our Postgres storage backend is stateful. For this reason, we'll use a StatefulSet to manage the backend. Please consult [the kubernetes docs](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) to better understand why a StatefulSets are appropriate for stateful applications.

To deploy a Postgres StatefulSet, follow these steps:
1. Run this command to create a dedicated namespace for the storage backend:
    ```bash
    $ kubectl create namespace quick-start-db
    ```
2. Run this command to deploy the Postgres StatefulSet:

    ```bash
    $ cat <<EOF | kubectl apply --namespace quick-start-db -f -
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
              - name: POSTGRES_USER
                value: postgres
              - name: POSTGRES_PASSWORD
                value: postgres
    EOF
    ```
   This StatefulSet uses the `postgres:9.6` container image available from DockerHub. When the container is started, the environment variables `POSTGRES_USER` and `POSTGRES_PASSWORD` are used to create a user with superuser power and a database with the same name.
   
   We'll treat these credentials as `admin-credentials` moving forward (i.e. `REMOTE_DB_ADMIN_USER` and `REMOTE_DB_ADMIN_PASSWORD`), to be used for administrative tasks - separate from `application-credentials`.
3. Run this command to deploy a `Service` that exposes the Postgres backend on node port 30001:
    ```bash
    $ cat <<EOF | kubectl apply --namespace quick-start-db -f -
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
    
    In this way, the Postgres backend becomes available via `$NODE_IP:30001` (referred to as `REMOTE_DB_URL`, moving forward).
 
### Setup and Configure protected resources

Before proceeding, in this section, we assume that you have a Postgres backend set up and ready to go. Concretely this means:

+ The Postgres backend is publicly available via some URL. We'll refer to this public URL by the environment variable `REMOTE_DB_URL`
+ Admin-level database credentials exist to allow you to create the application user. We'll refer to them as the environment variables `REMOTE_DB_ADMIN_USER` and `REMOTE_DB_ADMIN_PASSWORD`

In the section that follow, you will setup and configure the Postgres storage backend by carrying the following steps:
+ Create a dedicated application database within the Postgres DBMS, then carry out the rest of the steps within its context
+ Create the application table (i.e. notes)
+ Create an application user with limited privileges, `SELECT` and `INSERT` on the application table 
+ Store the application user's credentials (`APPLICATION_DB_USER` and `APPLICATION_DB_INITIAL_PASSWORD`) in Kubernetes secrets. You must choose the value of these environment variable, specify and export them before proceeding e.g. `export APPLICATION_DB_USER=app_user`

Run this `psql` command to create the application database, application table, application user and grant the application user relevant privileges:

```bash
$ docker run --rm -it postgres:9.6 env \
 PGPASSWORD=${REMOTE_DB_ADMIN_PASSWORD} psql -U ${REMOTE_DB_ADMIN_USER} "postgres://${REMOTE_DB_URL}" -c "
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
"
```

### Create application namespace and store application-user credentials

Now that the storage backend is setup and good to go, it's time to deploy the application with secretless.

The first step is to decide on an application namespace. This is the namespace in which the application will be scoped. You can pick your own. We'll refer to this namespace as `APPLICATION_NAMESPACE` from now on.

Run this code to create the namespace:

```yaml
$ kubectl create namespace ${APPLICATION_NAMESPACE}
```

Now that the namespace is created, you will proceed to store the application-user credentials in Kubernetes secrets. Anything but hardcoding them :)

Run this code to store application-user credentials in Kubernetes secrets:

```bash
$ cat <<EOF | kubectl apply --namespace ${APPLICATION_NAMESPACE} -f -
---
apiVersion: v1
kind: Secret
metadata:
    name: quick-start-backend-credentials
type: Opaque
data:
    address: $(echo -n ${REMOTE_DB_URL} | base64)
    username: $(echo -n ${APPLICATION_DB_USER} | base64)
    password: $(echo -n ${APPLICATION_DB_INITIAL_PASSWORD}" | base64)
EOF
```


### Deploy and run Application + Secretless

In this section, you create the deployment manifest for deploying your application and Secretless by building on this base manifest:

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
      ...
```

In this base manifest, you declare a deployment of 3 replicas with associated metadata. The next steps are to:

+ Add and configure the Application
+ Create and store secretless configuration as a ConfigMap
+ Add and configure the Secretless sidecar container

#### Add and configure application container

The sample application receives it's configuration via environment variables. For this reason we add `DATABASE_URL=postgresql://localhost:5432/quick_start_db?sslmode=disable` to the application container spec in the application deployment manifest at `$.spec.template.spec.containers`.

The secretless broker sidecar container has a shared network with the application container. This allows us to point the application to `localhost` where Secretless is listening on port `5432`.

Application must connect to Secretless without SSL, though the actual connection between Secretless and the database can leverage SSL. We include `sslmode=disable` in the connection string to prevent the Go Postgres driver from using SSL mode with Secretless. Note that Secretless respects the parameters specified in the database connections string e.g. `db_url/:db_name?param=param_value`.

Ultimately, the container definition for the application looks as follows:

```yaml
$.spec.template.spec.containers...
    - name: quick-start-application
      image: codebykumbi/note-store-app:latest
      env:
        - name: DATABASE_URL
          value: postgresql://localhost:5432/quick_start_db?sslmode=disable
```

#### Create and store Secretless configuration as ConfigMap

There are 3 steps to configuring Secretless for usage by the application
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

This configuration is shared amongst secretless sidecar containers for each application pod replica. A ConfigMap is created to house this configuration and expose it to the secretless sidecar container via a volume mount.

Run this code to create the config map:
```bash
$ kubectl create configmap quick-start-application-secretless-config \
  --namespace ${APPLICATION_NAMESPACE} \
  --from-file=secretless.yml
```

### Add and Configure Secretless sidecar container

Below is the template spec for the application and the secretless sidecar. It includes:

1. Secretless sidecar container
2. Read-only volume mounts on the secretless sidecar container for:
    + secretless configuration ConfigMap (`quick-start-application-secretless-config`)
    + the Kubernetes secrets containing the application-user DB credentials (`quick-start-backend-credentials`) 

_quick-start.yml_
```yaml
$.spec.template.spec...
    containers:
      - name: quick-start-application
        ...
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
          items:
          - key: address
            path: address
          - key: username
            path: username
          - key: password
            path: password
      - name: config
        configMap:
          name: quick-start-application-secretless-config
```

#### Deploy and run Application + Secretless

Run this command to deploy the application:
```bash
$ kubectl apply --namespace ${APPLICATION_NAMESPACE} -f quick-start.yml
```

Run this command to ensure the application pods have started and are healthy:
```bash
$ kubectl get po --namespace ${APPLICATION_NAMESPACE}
```
#### Expose application publicly

Run this command to expose the application on node port 30002:

```bash
$ cat <<EOF | kubectl apply --namespace ${APPLICATION_NAMESPACE} -f -
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

### Consume application

Now that the application is up and running, let's test it out.

Run the command below to create a note:
```bash
$ curl \
 -v \
 -d '{"title":"Secretless release", "description":"Once the tutorials are uploaded, initiate the release!"}' \
 -H "Content-Type: application/json" \
 $APPLICATION_URL/note
```
We expect the command above to respond with HTTP status 201.

Run the command below to retrieve all the notes:
```bash
$ curl $APPLICATION_URL/note
```
We expect the command above to respond with a JSON array containing the previously created note.

There we have it. This application is communicating with a protected resource without managing any secrets.

#TODO:
### Rotate protected resource credentials
## Review complete sample repository
## Next steps

