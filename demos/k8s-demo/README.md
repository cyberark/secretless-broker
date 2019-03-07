# Advanced Introduction to Secretless in Kubernetes

This tutorial will get Secretless running quickly in Kubernetes.  

It's intended for advanced users and is light on explanations.

For a friendlier version of this tutorial, with breakdowns of every step, try:

[Our Detailed Introduction to Secretless in Kubernetes](https://secretless.io/docs/get_started/kubernetes_tutorial.html)

## Overview

Here's what we'll do:

1. Deploy a PostgreSQL database
2. Store its credentials in Kubernetes secrets
3. Setup Secretless Broker to proxy connections to it 
4. Deploy a sample application that connects to the database **without knowing
   its password**

You'll play two roles in this tutorial:

1. A **Security Admin** who handles secrets, and has sole access to those secrets
2. An **Application Developer** with no access to secrets.

**As the security admin:**

1. Create a PostgreSQL database
1. Create a DB user for the application
1. Add that user's credentials to Kubernetes Secrets
1. Configure Secretless to connect to PostgreSQL using those credentials

**As the application developer:**

1. Configure the application to connect to PostgreSQL via Secretless
1. Deploy the application and the Secretless sidecar
To accomplish this, we are going to do the following:

## Prerequisites

+ A running Kubernetes cluster (you can use
  [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) to run a
  cluster locally)
+ [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured
  to point to the cluster
+ [Docker CLI](https://docs.docker.com/install/)

## Steps for Security Admin

### 1. Create PostgreSQL Database

Deploy a PostgreSQL instance using a **StatefulSet** in the
**quick-start-backend-ns**:

```bash
./01_create_db.sh
```

<p></p>
<details>
  <summary>View expected output</summary>
  <pre>
>>--- Clean up quick-start-backend-ns namespace
namespace "quick-start-backend-ns" deleted
Waiting for quick-start-backend-ns namespace clean up
namespace "quick-start-backend-ns" created
Ready!
secret "quick-start-backend-certs" created
>>--- Create database
statefulset "pg" created
service "quick-start-backend" created
Waiting for quick-start-backend to be ready
Ready!
CREATE DATABASE
  </pre>
</details>
<p></p>

Note we upload test certificates to the PostgreSQL container using Kubernetes
Secrets. In practice, you'll have your own certificates. For more info see
[PostgreSQL documentation](https://www.postgresql.org/docs/9.6/ssl-tcp.html).

### 2. Configure Database and Kubernetes Secrets

Next we'll:

- Create the DB user and table
- Create the Kubernetes Service
- Add the application's credentials to Kubernetes Secrets

```bash
./02_configure_db.sh
```

<p></p>
<details>
  <summary>View expected output</summary>
  <pre>
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
  </pre>
</details>
<p></p>

### 3. Configure the Secretless Broker

The [secretless.yml](/demos/k8s-demo/etc/secretless.yml) config file defines a
PostgreSQL listener on port 5432 and a handler that retrieves the previous
step's credentials from Kubernetes Secrets.

This step will be performed by the script `./03_deploy_app.sh` below.

<p></p>
<details>
  <summary>View "secretless.yml"</summary>
  <pre>
    <code>
listeners:
  - name: pg
    protocol: pg
    address: localhost:5432
handlers:
  - name: pg
    listener: pg
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
    </code>
  </pre>
</details>
<p></p>

## Steps for Application Developer

**Important: The application developer never needs any credentials to connect
to the database.**

In particular, as the application developer you do not know
any of the secrets in **admin_config.sh**.

### 1. Tell application to connect to Secretless

In the application manifest, we set the `DB_URL` environment variable to
`localhost:5432`, where the Secretless Broker is listening.

The application connects to Secretless.  Secretless connects to the database.

### 2. Deploy application

To perform all remaining steps run:

```bash
./03_deploy_app.sh
```

<p></p>
<details>
  <summary>View expected output</summary>
  <pre>
>>--- Create and store Secretless configuration
configmap/quick-start-application-secretless-config created
>>--- Start application
deployment.apps/quick-start-application created
service/quick-start-application created
Waiting for quick-start-application to be ready
...
Ready!
  </pre>
</details>
<p></p>

## Try it out!

That's it!

The application is connecting to a password-protected Postgres database
**without any knowledge of the credentials**.

Let's test it...

Our sample [pet store demo
application](https://github.com/conjurdemos/pet-store-demo) has a simple API:

- `GET /pets` lists all the pets
- `POST /pet` adds a pet

To test both adding a pet and listing all pets, run:

```bash
./04_test_deployment.sh
```

<p></p>
<details>
  <summary>View expected output</summary>
  <pre>
Adding a pet...
HTTP/1.1 201 
Location: http://192.168.99.100:30002/pet/1
Content-Length: 0
Date: Thu, 07 Mar 2019 05:03:58 GMT

Checking the pets...
HTTP/1.1 200 
Content-Type: application/json;charset=UTF-8
Transfer-Encoding: chunked
Date: Thu, 07 Mar 2019 05:04:02 GMT

[{"id":1,"name":"Mr. Snuggles"}]
  </pre>
</details>
<p></p>

## Conclusion

Please let us know what you think of Secretless! You can submit [Github
issues](https://github.com/cyberark/secretless-broker/issues) for features
you'd like to see, or send a message to our [mailing
list](https://groups.google.com/forum/#!forum/secretless) with comments or
questions.
