# Introduction to Secretless on K8S

## Description: Secretless

Secrets are used to provide priviledged access to protected resources.
Secretless pushes the trust boundary of secrets away from application code into a priviledged process that's designed  with security best practices in mind. The Secretless broker provides a local interface for application code to transparently consume protected resources.

## Usage: Secretless Broker as Sidecar

The Secretless broker operates as a sidecar container within a kubernetes application pod. This means there is shared storage/network between the application container and the Secretless broker. It is this which allows Secretless to provide a local interface.

The following steps are generally required to get up and running with Secretless.

1. Provision protected resources
2. Setup protected resources for usage by application
4. Add credentials to secret store
3. Configure Secretless to broker connection using credentials from the secret store
5. Configure application to connect to protected resource through interface exposed by Secretless 
6. Run Secretless adjacent to the application

## Quickstart

This example shows how easy it is to leverage the Secretless broker with an application that uses 12-factor principles to configure access to a database via a DATABASE_URL environment variable.

### Prerequisites
+ Kubernetes cluster
+ kubectl pointed to cluster
+ docker-cli

### Set up working environment

Run through the following commands to set up an environment in which a simple note storage-and-retrieval application makes use of Secretless to access a postgres storage backend:

#### Provision database

1. Provision protected resources


#### [choice 1] Postgres inside k8s

Note: If not running minikube, modify IP address component of `DB_URL` to match the public IP of your cluster in `./config.sh`.

Run the following script to create a pg stateful state in the `quick-start-db` namespace:

```
./01_create_db.sh
```

#### [choice 2] Remote Postgres

+ Ensure your Kubernetes cluster is able to access your remote db.
+ Ensure the remote instance has a database called `quick_start_db`
+ Update `DB_` env vars in `./config.sh`. For example (with Amazon RDS):

```
DB_URL=quick-start-db-example.xyzjshd3bdk3.us-east-1.rds.amazonaws.com:5432/quick_start_db
DB_ROOT_USER=quick_start_db
DB_ROOT_PASSWORD=quick_start_db
DB_USER=quick_start_user
DB_INITIAL_PASSWORD=quick_start_user
```


#### Setup database and add credentials to secret store

2. Setup protected resources for usage by application
3. Add credentials to secret store

Run:
```
./02_setup_db.sh
```

#### Build and deploy application

4. Configure Secretless to broker connection using credentials from the secret store
5. Configure application to connect to protected resource through interface exposed by Secretless 
6. Run Secretless adjacent to the application

*Feel free to modify `quick-start.yml:35` to use your own application image.*

Run: 
```
./03_start_app.sh
```

### Interact with working environment

#### Port-forward application service to localhost

Port-forwarding allows local access to the application service. This is to avoid the complexities of publicly exposing an application outside of the cluster.
 
```
kubectl port-forward -n quick-start svc/quick-start-application 8080:8080
```

The application service should then be available at `localhost:8080`.

#### Consume application API

GET `/note` to retrieve notes
```
curl localhost:8080/note
```
POST `/note` to add a note - title and description must be specified via json body.
```
curl \
 -d '{"title":"value1", "description":"value2"}' \
 -H "Content-Type: application/json" \
 -X POST \
 localhost:8080/note
```

#### Rotate application database credentials

Rotator script rotates database credentials in k8s secrets, waits for secrets to be propagated to application pod volumes, then rotates database credentials. k8s secrets updates have an associated lag. Using a secret management solution, such as Conjur, would avoid the need to wait for propagation.

Run the following with your own value for `>new password value<`:

```
./rotate_password >new password value<
```

Observe that requests to the application API are not encumbered by rotation.
 