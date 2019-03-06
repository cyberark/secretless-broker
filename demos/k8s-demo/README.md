# Advanced Introduction to Secretless in Kubernetes

This tutorial gets Secretless running quickly in Kubernetes.  It's intended for
advanced Kubernetes users.

For a friendlier version of this tutorial, including explanations of every step, try:

[Detailed Introduction to Secretless in
Kubernetes](https://secretless.io/docs/get_started/kubernetes_tutorial.html)

## Usage: Secretless Broker as Sidecar

The Secretless Broker operates as a sidecar container within a kubernetes
application pod. This means there is shared storage/network between the
application container and the Secretless Broker. It is this which allows the
Secretless Broker to provide a local interface.

In this tutorial, we will walk through creating an application that
communicates with a password-protected PostgreSQL database via the Secretless
Broker. _The application does not need to know anything about the credentials
required to connect to the database;_ the admin super-user who provisions and
configures the database will also configure the Secretless Broker to be able to
communicate with it. The developer writing the application only needs to know
the socket or address that the Secretless Broker is listening on to proxy the
connection to the PostgreSQL backend.

To accomplish this, we are going to do the following:

**As the admin super-user:**

1. Provision protected resources
1. Configure protected resources for usage by application and add credentials to a secret store
1. Configure the Secretless Broker to broker the connection using credentials from the secret store
1. Create application identity and grant entitlements to provide access to credentials from the secret store

**As the application developer:**
1. Configure the application to connect to protected resource through the interface exposed by the Secretless Broker
1. Deploy and run the Secretless Broker adjacent to the application

## Quickstart

The tutorial uses an existing [pet store demo application](https://github.com/conjurdemos/pet-store-demo) that exposes the following routes:

- `GET /pets` to list all the pets in inventory
- `POST /pet` to add a pet
  - Requires **Content-Type: application/json** header and body that includes **name** data

There are additional routes that are also available, but these are the two that we will focus on for this tutorial.

Pet data is stored in a PostgreSQL database, and the application may be configured to connect to the database by setting the `DB_URL`, `DB_USERNAME`, and `DB_PASSWORD` environment variables in the application's environment (following [12-factor principles](https://12factor.net/)).

We are going to deploy the application with the Secretless Broker to Kubernetes, configure the Secretless Broker to be able to retrieve the credentials from a secrets store, and configure the application with the `DB_URL` environment variable pointing to the Secretless Broker _and no values set for the `DB_USERNAME` or `DB_PASSWORD` environment variables_.

### Prerequisites

To run through this tutorial, you will need:

+ A running Kubernetes cluster (you can use [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) to run a cluster locally)
+ [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured to point to the cluster
+ [Docker CLI](https://docs.docker.com/install/)

Our Kubernetes deployment manifests assume that you are using Minikube, so that for example **./etc/pg.yml** and **./etc/quick-start.yml** use **NodePort** to expose the services; you may prefer to use a **LoadBalancer** for a GKE cluster.

#### Suggested modifications for advanced demos

Once you have run through this tutorial, you may enjoy trying it with some modifications to make it more pertinent to you. Here are some suggestions for things to try:

- We've provided a sample application for you to try with the Secretless Broker - but if you're interested in exploring further, you can try out replacing it with your own app. To do this, you'll want to:
  - Modify **quick-start.yml:35** to use your own application image
  - Update **02_configure_db.sh** to appropriately configure the PostgreSQL database for your own application

- You can use your own PostgreSQL database rather than using the database we deploy in this demo; for information on how to do this, please see "Option 2" of the [provision database](#1-provision-database) step.

### Steps for the admin-level user

The following steps would be taken by an admin-level user, who has the ability to create and configure a database and to add secret values to a secret store.

These steps make use of the **admin_config.sh** file, which stores the database connection info for the PostgreSQL backend.

#### 1. Provision database

+ Provision protected resources

  **[Option 1] PostgreSQL inside k8s**

  Run the following script to deploy a PostgreSQL instance  using a **StatefulSet**
  in the **quick-start-backend-ns** namespace:

  ```bash
  ./01_create_db.sh
  ```
    ```
    >>--- Clean up quick-start-backend-ns namespace
    Error from server (NotFound): namespaces "quick-start-backend-ns" not found
    namespace "quick-start-backend-ns" created
    Ready!
    secret "quick-start-backend-certs" created
    >>--- Create database
    statefulset "pg" created
    service "quick-start-backend" created
    Waiting for quick-start-backend to be ready
    ...
    Ready!
    CREATE DATABASE
    ```
    Note that the PostgreSQL instance is deployed configured for SSL. We provide
    simple test certificates in this demo that we upload to the PostgreSQL container
    using Kubernetes Secrets. In practice, you will have your own certificates; for
    guidelines on creating the certificates for your PostgreSQL instance, check out
    the [PostgreSQL documentation](https://www.postgresql.org/docs/9.6/ssl-tcp.html).

  **[Option 2] Remote PostgreSQL server**

  + Ensure your Kubernetes cluster is able to access your remote DB.
  + Ensure your remote instance supports SSL
  + Ensure the remote instance has a database called **quick_start_db**
  + Update the `DB_` env vars in **./admin_config.sh**. For example (with Amazon RDS):

    ```bash
    DB_URL=quick-start-backend-example.xyzjshd3bdk3.us-east-1.rds.amazonaws.com:5432/quick_start_db
    DB_ADMIN_USER=quick_start_db
    DB_ADMIN_PASSWORD=quick_start_db
    DB_USER=quick_start_user
    DB_INITIAL_PASSWORD=quick_start_user
    ```

#### 2. Configure database and add credentials to secret store

In this step, we will:

+ Configure the protected resources for usage by application (i.e. create DB user, add tables, etc.)
+ Add the application's access credentials for the database to a secret store

Run:
```bash
./02_configure_db.sh
```

```
>>--- Set up database
CREATE ROLE
CREATE TABLE
GRANT
GRANT
>>--- Clean up quick-start-application-ns namespace
namespace/quick-start-application-ns created
Ready!
secret/quick-start-backend-credentials created
serviceaccount/quick-start-application created
role.rbac.authorization.k8s.io/quick-start-backend-credentials-reader created
rolebinding.rbac.authorization.k8s.io/read-quick-start-backend-credentials created
```

#### 3. Configure the Secretless Broker to broker the connection to the target service

In the last step, we added the database credentials to our secret store - so to configure the Secretless Broker to be able to retrieve these credentials and proxy the connection to the actual PostgreSQL database, we have written a [secretless.yml](/demos/k8s-demo/etc/secretless.yml) file that defines a PostgreSQL listener on port 5432 that uses the Kubernetes Secrets Provider to retrieve the credential values that we stored when we ran the last script:

```yaml
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
```
Note: we don't specify an `sslmode` in the Secretless Broker config, so it will
use the default `require` value.

### Steps for the non-privileged user (i.e. developer)

**Note:** None of these steps require the information in `admin_config.sh` - the person deploying the application needs to know _nothing_ about the secret values required to connect to the PostgreSQL database!!

#### 1. Configure application to access the database at `localhost:5432`

In the application manifest, we set the `DB_URL` environment variable to `localhost:5432`, so that when the application is deployed it will open the connection to the PostgreSQL backend via the Secretless Broker.

#### 2. Deploy application

To deploy the application with the Secretless Broker, run:
```bash
./03_deploy_app.sh
```
```
>>--- Create and store Secretless configuration
configmap/quick-start-application-secretless-config created
>>--- Start application
deployment.apps/quick-start-application created
service/quick-start-application created
Waiting for quick-start-application to be ready
...
Ready!
```

### Try it out!

That's it! You've configured your application to connect to PostgreSQL via the Secretless Broker, and we can try it out to validate that it's working as expected.

#### Use the pet store app

`POST /pet` to add a pet - the request must include `name` in the JSON body
```bash
APPLICATION_URL=$(. ./admin_config.sh; echo ${APPLICATION_URL})

curl \
  -i \
  -d '{"name": "Mr. Snuggles"}' \
  -H "Content-Type: application/json" \
  ${APPLICATION_URL}/pet
```
```bash
HTTP/1.1 201
Location: http://192.168.99.100:30002/pet/1
Content-Length: 0
Date: Thu, 23 Aug 2018 12:57:45 GMT
```

`GET /pets` to retrieve notes
```bash
APPLICATION_URL=$(. ./admin_config.sh; echo ${APPLICATION_URL})

curl -i ${APPLICATION_URL}/pets
```
```
HTTP/1.1 200
Content-Type: application/json;charset=UTF-8
Transfer-Encoding: chunked
Date: Thu, 23 Aug 2018 12:58:16 GMT

[{"id":1,"name":"Mr. Snuggles"}]
```

For convenience, we've provided the `04_test_deployment.sh` script for you to
test adding a pet and retrieving the list of pets.

#### Rotate application database credentials

In addition to the demo you've seen so far, you can also **rotate the DB credentials** and watch the app continue to perform as expected.

The rotator script:
 + Rotates the credentials in the database
 + Updates the password in the secrets store
 + Prunes previously open connections

To see graceful rotation in action, poll the endpoint to retrieve the list of pets (`GET /pets`) in a separate terminal before rotating:

```
APPLICATION_URL=$(. ./admin_config.sh; echo ${APPLICATION_URL})

while true
do
    echo "Retrieving pets"
    curl -i ${APPLICATION_URL}/pets
    echo ""
    echo ""
    echo "..."
    echo ""
    sleep 3
done
```
```
Retrieving pets
HTTP/1.1 200
Content-Type: application/json;charset=UTF-8
Transfer-Encoding: chunked
Date: Thu, 23 Aug 2018 12:58:43 GMT

[{"id":1,"name":"Mr. Snuggles"}]

...
.
.
.
```
To rotate the database password (note: you are acting as an admin user), run the following with your own value for `[new password value]`:

```bash
./rotate_password.sh [new password value]
```
```
ALTER ROLE
secret/quick-start-backend-credentials patched
 pg_terminate_backend
----------------------
 t
 t
 t
 .
 .
 .
(30 rows)

```

Observe that requests to the application API are not affected by the password rotation - we continue to be able to query the application as usual, without interruption!

## Conclusion

If you enjoyed this Secretless Broker tutorial, please try to make it your own by trying out some of the [suggested modifications](#suggested-modifications-for-advanced-demos). Please also let us know what you think of it! You can submit [Github issues](https://github.com/cyberark/secretless-broker/issues) for features you would like to see, or send a message to our [mailing list](https://groups.google.com/forum/#!forum/secretless) with comments and/or questions.
