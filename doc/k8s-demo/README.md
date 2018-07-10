# Introduction to Secretless on K8S

## Description: Secretless

Secrets are used to provide priviledged access to protected resources.
Secretless pushes the trust boundary of secrets away from application code into a priviledged process that's designed  with security best practices in mind. The secretless broker provides a local interface for application code to transparently consume protected resources.

## Usage: Secretless Broker as Sidecar

The secretless broker operates as a sidecar container within a kubernetes application pod. This means there is shared storage/network between the application container and the secretless broker. It is this which allows secretless to provide a local interface.

The following steps are generally required to get up and running with secretless.

1. Provision protected resources
2. Add credentials to secret store
3. Configure Secretless to broker connection using credentials from the secret store
4. Configure application to connect to protected resource through interface exposed by secretless 
5. Run secretless adjacent to the application

## Quickstart

This example shows how easy it is to leverage the secretless broker with an application that uses 12-factor principles to configure access to a database via a DATABASE_URL environment variable.

### Prerequisites
+ kubernetes cluster running in minikube
+ kubectl pointed to minikube cluster
+ docker-cli pointed to daemon inside minikube cluster

### Set up working environment

Run through the following commands to set up an environment in which a simple note storage-and-retrieval application makes use of secretless to access a postgres storage backend:

```
./start_db.sh
```

1. Provision protected resources
2. Add credentials to secret store

```
./start_app.sh
```

3. Configure Secretless to broker connection using credentials from the secret store
4. Configure application to connect to protected resource through interface exposed by secretless 
5. Run secretless adjacent to the application

### Interact with working environment

#### Consume application API
GET `/note` to retrieve notes
```
curl $(minikube service -n quick-start quick-start-application --url)/note
```
POST `/note` to add a note - title and description must be specified via json body.
```
curl \
 -d '{"title":"value1", "description":"value2"}' \
 -H "Content-Type: application/json" \
 -X POST \
 $(minikube service -n quick-start quick-start-application --url)/note
```

#### Rotate application database credentials

Run the following with your own value for `>new password value<`
```
./rotate_password >new password value<
```

## TODO: 
+ add flexibility to `./start_db.sh` and `./rotate_password.sh` to work with remote database
