# demo-app-secretless

[The Secretless Broker](https://github.com/cyberark/secretless-broker) is a connection broker which relieves client applications of the need to directly handle secrets to target services such as databases, web services, SSH connections, or any other TCP-based service.

## TL;DR;

```bash
$ helm install -f values.yaml .
```

## Introduction

This chart bootstraps a deployment of a configurable demo-application, which uses [PostgreSQL](https://github.com/docker-library/postgres) as its backend, with Secretless on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager. Secretless seamlessly manages the database connection for this application.

## Prerequisites

- Kubernetes 1.4+ with Beta APIs enabled

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
$ helm install --name my-release \
 --set applicationDBAddress="${DB_URL}" \
 --set applicationDBUsername="${DB_USER}" \
 --set applicationDBPassword="${DB_INITIAL_PASSWORD}"
 .
```

The command deploys PostgreSQL on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the PostgreSQL chart and their default values.

| Parameter                     | Description                                     | Default                                                    |
| -----------------------       | ---------------------------------------------   | ---------------------------------------------------------- |
| `applicationDBAddress`        | `postgres` database URL, e.g. db.svc.cluster.local:5432/table-name | `nil` (required)                                           |
| `applicationDBUsername`       | Application database username                   | `nil` (required)                                           |
| `applicationDBPassword`       | Application database password                   | `nil` (required)                                           |
| `applicationPort`             | Image pull secrets                              | `nil` (required)                                           |
| `withSecretless`              | Inject and leverage Secretless broker           | `true`                                                     |
| `applicationImage.repository` | Application image repository                    | `cyberark/demo-app`                                        |
| `applicationImage.tag`        | Application image tag                           | `latest`                                                   |
| `applicationImage.pullPolicy` | Application image pull policy                   | `IfNotPresent`                                             |
| `service.port`                | Application service TCP port                    | `80`                                                       |
| `service.type`                | Application service type exposing ports, e.g. `NodePort`| `ClusterIP`                                                |

The above application db-credentials map to Secrets to the env variables defined in [postgres](http://github.com/docker-library/postgres). For more information please refer to the [postgres](http://github.com/docker-library/postgres) image documentation.

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
$ helm install --name my-release \
   --set applicationDBAddress="db.svc.cluster.local:5432/my-database" \
   --set applicationDBUsername="admin" \
   --set applicationDBPassword="password123"
   .
```

The above command creates a deployment of application + Secretless pods, stores the application db-credentials in the secret store for Secretless to use in brokering a connection to the remote database at `db.svc.cluster.local:5432`  named `my-database`.

Setting `withSecretless` to `false` will only deploy the application, hard-coding the credentials as environment variables. This is useful for comparison purposes.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
$ helm install --name my-release -f values.yaml .
```

### End to End with a PostgresSQL Helm Chart

1. Setup shared environment variables:
```bash
BACKEND_NAMESPACE=quick-start-backend
BACKEND_RELEASE_NAME=quick-start-backend
DB_NAME=quick_start_db
DB_URL=${BACKEND_RELEASE_NAME}-postgresql.${BACKEND_NAMESPACE}.svc.cluster.local:5432 # CHANGE to reflect endpoint exposed by db service

# admin-user credentials
DB_ADMIN_USER=postgres
DB_ADMIN_PASSWORD=admin_password

# application-user credentials
DB_USER=quick_start_application
DB_INITIAL_PASSWORD=quick_start_application_password
```

2. Install a PostgresSQL release using Helm chart
```bash
$ helm install \
  --name ${BACKEND_RELEASE_NAME} \
  --namespace ${BACKEND_NAMESPACE} \
  --set postgresUser=${DB_ADMIN_USER} \
  --set postgresPassword=${DB_ADMIN_PASSWORD} \
  --set postgresDatabase=postgres \
  stable/postgresql
```

3. Configure PostgresSQL release for application 
```bash
$ kubectl run --rm -i configure-quick-start-backend --env PGPASSWORD=${DB_ADMIN_PASSWORD} --image=postgres:9.6 --restart=Never --command -- psql \
  -U ${DB_ADMIN_USER} \
  "postgres://$DB_URL" \
  << EOL
/* Create Application Database */
CREATE DATABASE "${DB_NAME}";

/* Create Application User */
CREATE USER ${DB_USER} PASSWORD '${DB_INITIAL_PASSWORD}';
/* Create Table */
CREATE TABLE pets (
    id serial primary key,
    name varchar(256)
);
/* Grant Permissions */
GRANT SELECT, INSERT ON public.pets TO ${DB_USER};
GRANT USAGE, SELECT ON SEQUENCE public.pets_id_seq TO ${DB_USER};
EOL
```

4. Install application release (with Secretless) using this Helm chart
```bash
$ helm install \
  --set applicationDBAddress="${DB_URL}/${DB_NAME}" \
  --set applicationDBUsername="${DB_USER}" \
  --set applicationDBPassword="${DB_INITIAL_PASSWORD}" \
  .
```

5. Install application release (without Secretless) using this Helm chart
```bash
$ helm install \
  --set applicationDBAddress="${DB_URL}/${DB_NAME}" \
  --set applicationDBUsername="${DB_USER}" \
  --set applicationDBPassword="${DB_INITIAL_PASSWORD}" \
  --set withSecretless=false \
  .
```
