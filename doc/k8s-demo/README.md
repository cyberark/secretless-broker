# Introduction to Secretless on Kubernetes

## Description: Secretless

Secrets are used to provide priviledged access to protected resources.
Secretless pushes the trust boundary of secrets away from application code into a priviledged process that's designed  with security best practices in mind. The Secretless broker provides a local interface for application code to transparently consume protected resources.

## Usage: Secretless Broker as Sidecar

The Secretless broker operates as a sidecar container within a kubernetes application pod. This means there is shared storage/network between the application container and the Secretless broker. It is this which allows Secretless to provide a local interface.

The following steps are generally required to get up and running with Secretless.

1. Provision protected resources
2. Configure protected resources for usage by application and add credentials to secret store
3. Configure Secretless to broker connection using credentials from the secret store, configure application to connect to protected resource through interface exposed by Secretless, deploy and run Secretless adjacent to the application.

## Quickstart

This example shows how easy it is to leverage the Secretless broker with an application that uses 12-factor principles to configure access to a database via a `DATABASE_URL` environment variable.

The database credentials of the application are those being transparently handled by secretless. The initial values of these credentials are set in `./config.sh` for convenience.

Admin-level database credentials are used to create the application user. 
These (`DB_ADMIN_USER` and `DB_ADMIN_PASSWORD`) are housed in `./config.sh` for convenience. In practice, an admin user would manage configuration of the database on their own and this config file would not be necessary.

### Prerequisites
+ Kubernetes cluster
+ kubectl pointed to cluster
+ docker-cli

#### Required modification to suite your needs

+ For this guide we've provided our own sample application - note storage-and-retrieval. Feel free to modify `quick-start.yml:35` to use your own application image, and see the power of secretless closer to home.*

+ Modify Service definition to use Service types that suit your needs.  The default just works with `minikube`.
  + Update `./etc/pg.yml` and `./etc/quick-start.yml` to suit your needs. The reference manifests use a `NodePort` type Service, which works with `minikube`. For example, a `Load Balancer` type Service might be more appropriate in a GKE cluster. 
  + Update `DB_URL` and `APPLICATION_URL` in `./config.sh` to reflect the endpoints made available by the aforementioned services.

### Set up working environment

Run through the following commands to set up an environment in which an application makes use of Secretless to access a postgres storage backend:

#### 1. Provision database

+ Provision protected resources

##### [choice 1] Postgres inside k8s

Run the following script to create a Postgres stateful-set instance in the `quick-start-db` namespace:

```
./01_create_db.sh
```

##### [choice 2] Remote Postgres

+ Ensure your Kubernetes cluster is able to access your remote DB.
+ Ensure the remote instance has a database called `quick_start_db`

#### 2. Configure database and add credentials to secret store

+ Update `DB_` env vars in `./config.sh`. For example (with Amazon RDS):

```
DB_URL=quick-start-db-example.xyzjshd3bdk3.us-east-1.rds.amazonaws.com:5432/quick_start_db
DB_ROOT_USER=quick_start_db
DB_ROOT_PASSWORD=quick_start_db
DB_USER=quick_start_user
DB_INITIAL_PASSWORD=quick_start_user
```

+ Configure protected resources for usage by application
+ Add credentials to secret store

Run:
```
./02_configure_db.sh
```

#### 3. Deploy application

+ Configure Secretless to broker connection using credentials from the secret store
+ Configure application to connect to protected resource through interface exposed by Secretless 
+ Run Secretless adjacent to the application

Run:
```
./03_deploy_app.sh
```

### Interact with working environment

#### Consume application API

GET `/note` to retrieve notes
```
APPLICATION_URL=$(. ./config.sh; echo $APPLICATION_URL)

curl $APPLICATION_URL/note
```

POST `/note` to add a note - title and description must be specified via json body.
```
APPLICATION_URL=$(. ./config.sh; echo $APPLICATION_URL)

curl \
 -d '{"title":"value1", "description":"value2"}' \
 -H "Content-Type: application/json" \
 $APPLICATION_URL/note
```

#### Rotate application database credentials

We've provided a rotator script that works by:
 + updating the password in the vault
 + waiting for the update to take effect
 + rotating the credentials with the database

Typically, you'd want your rotation to work the other way - update the DB and then your vault - but we're using kubernetes secrets in this guide, which isn't built to handle secret rotation gracefully. For that, you'd want to use a better secrets management solution.

To see graceful rotation, poll the retrieve notes endpoint (GET `/note`) in a separate terminal before rotating:

```
APPLICATION_URL=$(. ./config.sh; echo $APPLICATION_URL)

while true
do 
    echo "Retrieving notes"
    curl $APPLICATION_URL/note
    echo ""
    sleep 1
done
```

Run the following with your own value for `>new password value<`:

```
./rotate_password >new password value<
```

Observe that requests to the application API are not encumbered by rotation.
