[![pipeline status](https://gitlab.com/cyberark/secretless-broker/badges/master/pipeline.svg)](https://gitlab.com/cyberark/secretless-broker/commits/master)

# Table of Contents

- [Secretless Broker](#secretless-broker)
  - [Supported services](#currently-supported-services)
- [Quick Start](#quick-start)
- [Additional demos](#run-more-secretless-demos)
- [Using Secretless](#using-secretless)
  - [Service Connectors](#service-connectors)
  - [Credential Providers](#credential-providers)
  - [Health Checks](#health-checks)
- [Community](#community)
- [Performance](#performance)
- [Development](#development)
- [License](#license)


# Secretless Broker

The Secretless Broker is a connection broker which relieves client applications of the need to directly handle secrets to target services such as databases, web services, SSH connections, or any other TCP-based service.

![Secretless Broker Architecture](https://github.com/cyberark/secretless-broker/blob/master/docs/img/secretless_architecture.svg)

The Secretless Broker is designed to solve two problems. The first is **loss or theft of credentials from applications and services**, which can occur by:

- Accidental credential leakage (e.g. credential checked into source control, etc)
- An attack on a privileged user (e.g. phishing, developer machine compromises, etc)
- A vulnerability in an application (e.g. remote code execution, environment variable dump, etc)

The second is **downtime caused when applications or services do not respond to credential rotation** and crash or get locked out of target services as a result.

When the client connects to a target service through the Secretless Broker:

* **The client is not part of the threat surface**

  The client/app no longer has direct access to the password, and therefore cannot reveal it.

- **The client doesn’t have to know how to properly manage secrets**

  Handling secrets safely involves some complexities, and when every application needs to know how to handle secrets, accidents happen. The Secretless Broker centralizes the client-side management of secrets into one code base, making it easier for developers to focus on delivering features.

- **The client doesn’t have to manage secret rotation**

  The Secretless Broker is responsible for establishing connections to the backend, and can handle secrets rotation in a way that’s transparent to the client.

To provide the Secretless Broker access to a target service, a [Service Connector](#service-connectors) implements the protocol of the service, replacing the authentication handshake. The client does not need to know or use a real password to the service. Instead, it proxies its connection to the service through a local connection to the Secretless Broker. Secretless Broker obtains credentials to the target service from a secrets vault (such as Conjur, a keychain service, text files, or other sources) via a [Credential Provider](#credential-providers). The credentials are used to establish a connection to the actual service, and the Secretless Broker then rapidly shuttles data back and forth between the client and the service.

The Secretless Broker is currently licensed under [ASL 2.0](#license)

## Currently supported services

- MySQL (Socket and TCP)
- PostgreSQL (Socket and TCP)
- SSH / SSH-Agent (Beta)
- HTTP with Basic auth, Conjur, and AWS authorization strategies (Beta)

With many others in the planning stages!

If there is a specific target service that you would like to be included in this project, please open a [GitHub issue](https://github.com/cyberark/secretless-broker/issues) with your request.

For specific guidelines about using a particular service, please see our instructions for [using the Secretless Broker](#using-secretless).

# Quick Start

Running the quick start demo requires [Docker][get-docker].

[get-docker]: https://docs.docker.com/engine/installation

To see the Secretless Broker in action, build the quick start image:

```sh-session
$ cd demos/quick-start/
$ ./bin/build
...
Successfully built cbf747e7f548
Successfully tagged secretless-broker-quickstart:latest
```

The quick start image runs the Secretless Broker binary and a few sample target services. The Secretless Broker is configured to retrieve the credentials required to access the services from the process environment. **All services are configured to require authentication to access**, but we don't know what those credentials are. We can try to access the services, but since we don't know the password our access attempts will fail. _But when we try to connect via the Secretless Broker_, we will be granted access.

Let's try this with the PostgreSQL server running in the quick start image. We know that the server has been configured with a `quickstart` database, so let's try to access it.

You can run the Secretless Broker quick start as a Docker container:
```sh-session
$ docker run \
  --rm \
  -p 5432:5432 \
  -p 5454:5454 \
  secretless-broker-quickstart:latest
```

In a new window, if we try to connect to PostgreSQL directly via port 5432 (guessing at the `postgres` username), our attempt will fail:

```sh-session
$ psql "host=localhost port=5432 user=postgres dbname=quickstart sslmode=disable"
Password for user postgres:
psql: FATAL:  password authentication failed for user "postgres"
```

But the Secretless Broker is listening on port 5454, and will add authentication credentials (both username and password) to our connection request and proxy our connection to the PostgreSQL server:

```sh-session
$ psql "host=localhost port=5454 user=postgres dbname=quickstart sslmode=disable"
psql (10.3, server 9.6.9)
Type "help" for help.

quickstart=> \d
                List of relations
 Schema |      Name       |   Type   |   Owner    
--------+-----------------+----------+------------
 public | counties        | table    | secretless
 public | counties_id_seq | sequence | secretless
(2 rows)

quickstart=> select * from counties limit 1;
 id |   name    
----+-----------
  1 | Middlesex
(1 row)

quickstart=>
```

**Success! Smile and grab a :cookie: because it was too easy!**

You have just delegated responsibility for managing credentials to a secure process isolated from your app!

# Run more Secretless Broker demos

If the PostgreSQL quick start demo piqued your interest, please check out our [additional demos](https://github.com/cyberark/secretless-broker/tree/master/demos/quick-start#ssh-quick-start) where you can try the Secretless Broker with SSH and HTTP Basic Auth.

For an even more in-depth demo, check out our [Deploying to Kubernetes](https://github.com/cyberark/secretless-broker/tree/master/demos/k8s-demo) demo, which walks you through deploying a sample app to Kubernetes with the Secretless Broker.

# Using Secretless

For complete documentation on using Secretless, please see [our documentation](https://docs.secretless.io/Latest/en/Content/Resources/_TopNav/cc_Home.htm). The documentation includes comprehensive guides for how to get up and running with Secretless.

Secretless Broker relies on YAML configuration files to specify which target services it can connect to and how it should retrieve the access credentials to authenticate with those services.

The Secretless Broker configuration file begins with a version and a list of services:
```yaml
version: "2"
services:
  service-1:
    ...
```

Each individual service definition provides Secretless Broker with the information it needs to connect to a particular target service. In particular, Secretless needs to know:

* The protocol used by the target service
* Where to listen for new connection requests
* Where to get credentials for incoming connections
* The location of the target service (eg where to send requests)

Secretless Broker uses the protocol given in the service configuration to determine which [Service Connectors](#service-connectors) should process the connection request. Secretless retrieves any required credentials, revises the connection request to inject the valid authentication credentials, negotiates the authentication handshake with the target service, and then transparently streams data between the client and service.

A sample service configuration for PostgreSQL is below:

```yaml
version: "2"
services:
  postgres-db:
    protocol: pg
    listenOn: tcp://0.0.0.0:5432 # can be a socket as well (same name for both)
    credentials:
      host: "postgres.my-service.internal"
      password:
        from: conjur
        get: id-of-secret-in-conjur
      username:
        from: env
        get: username
    config:  # this section usually blank
      optionalStuff: foo
```

In this sample, Secretless Broker is configured to connect to a service named `postgres-db`. Clients send connection requests to this service via localhost using the default port 5432. This service uses the `pg` protocol, which indicates that the PostgreSQL service connector should process requests that come in via this port. Credentials from this connection will be retrieved from multiple sources; the `host` is given as a literal string value, the `password` will be retrieved from the Conjur variable with ID `id-of-secret-in-conjur`, and the `username` is retrieved from the environment variable named `username`.

## Service Connectors

Service connectors implement the protocol of the target service and are responsible for proxying connections and managing the authentication
handshake. When a Service Connector receives a new connection request, it retrieves the required credentials using the specified Provider(s), injects the correct authentication credentials into the connection request, and opens up a connection to the target service. From there,
the Service Connector simply transparently shuttles data between the client and service.

Secretless Broker comes with several built-in service connectors and each accepts a different set of credentials for configuration. In this section we provide some information on the credentials used by each service connector - for more complete information please see our [service connector documentation](https://docs.secretless.io/Latest/en/Content/References/connectors/overview.htm).

- MySQL (accepts connections over Unix socket or TCP)
  - Credentials:
    - `host`
    - `port`
    - `username`
    - `password`
- PostgreSQL (accepts connections over Unix socket or TCP)
  - Credentials:
    - `address`
    - `username`
    - `password`
- SSH
  - Credentials:
    - `address`
    - `privateKey`
    - `user`
      - optional; defaults to `root`
    - `hostKey`
      - optional; accepts any host key if not included
- SSH-Agent
  - Credentials:
    - `rsa`
    - `ecdsa`
    - `comment`
      - optional; free-form string
    - `lifetime`
      - optional; if not 0, the number of secs the agent will store the key for
    - `confirm`
      - optional; confirms with user before using if true
- HTTP
  - Basic Auth
    - Credentials:
      - `username`
      - `password`
      - `forceSSL` (optional)
  - Conjur
    - Credentials:
      - `accessToken`
      - `forceSSL` (optional)
  - AWS
    - Credentials:
      - `accessKeyID`
      - `secretAccessKey`
      - `accessToken`

## Credential Providers

Credential providers interact with a credential source to deliver secrets needed for authentication
to the Secretless Service Connectors. The Secretless Broker comes built-in with several different
Credential Providers, making it easy to use with your existing workflows regardless of your current
secrets management toolset.

We currently support the following secrets providers/vaults:
- [Conjur (`conjur`)](#conjur-provider)
- [HashiCorp Vault (`vault`)](#hashicorp-vault-provider)
- [Kubernetes Secrets Provider (`kubernetes`)](#kubernetes-secrets-provider)
- [File Provider (`file`)](#file-provider)
- [Environment Variable (`env`)](#environment-variable-provider )
- [Literal Value (`literal`)](#literal-value-provider)
- [Keychain (`keychain`)](#keychain-provider)

Secretless must be configured to authenticate with any credential providers referenced in its configuration. For example, if Secretless will be retrieving credentials from a vault, the Secretless application must itself be able to authenticate with the vault and must be authorized to retrieve the necessary credentials. For more information on how this works in practice, please review the [credential provider documentation](https://docs.secretless.io/Latest/en/Content/References/providers/overview.htm).

### Conjur Provider

Conjur (`conjur`) provider allows use of [CyberArk Conjur](https://www.conjur.org/) for fetching secrets.

Example:
```
...
    credentials:
      - name: accessToken
        provider: conjur
        id: path/to/the/token
...
```

### HashiCorp Vault Provider

**Note: this provider is in an _alpha_ state due to its lack of support for [native Kubernetes auth](https://www.vaultproject.io/docs/auth/kubernetes.html).**

Vault (`vault`) provider allows use of [HashiCorp Vault](https://www.vaultproject.io/) for fetching secrets.

Example:
```
...
    credentials:
      - name: accessToken
        provider: vault
        id: path/to/the/token
...
```

### Kubernetes Secrets Provider (beta)

Kubernetes Secrets (`kubernetes`) provider allows use of [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) for fetching secrets.

Example:
```
...
    credentials:
      - name: accessToken
        provider: kubernetes
        id: secret_identifier#key
...
```

### File Provider

File (`file`) provider allows you to use a file available to the Secretless Broker process and/or container as sources of
credentials.

Example:
```
...
    credentials:
      - name: rsa
        provider: file
        id: /path/to/file
...
```

### Environment Variable Provider

Environment (`env`) provider allows use of environment variables as source of credentials.

Example:
```
...
    credentials:
      - name: accessToken
        provider: env
        id: ACCESS_TOKEN
...
```

### Literal Value Provider

Literal (`literal`) provider allows use of hard-coded values as credential sources.

_Note: This type of secrets inclusion is highly likely to be much less secure versus other
available providers so please use care when choosing this as your secrets source._

Example:
```
...
    credentials:
      - name: accessToken
        provider: literal
        id: supersecretaccesstoken
...
```

### Keychain Provider (beta)

Keychain (`keychain`) provider allows use of your OS-level keychain as the credentials provider.

_Note: This provider currently only works on Mac OS at the time and only when building from source so it should
be avoided unless you are a developer working on the source code. There are plans to integrate all major OS
keychains into this provider in a future release._

Example:
```
...
    credentials:
      - name: rsa
        provider: keychain
        id: servicename#accountname
...
```

## Health Checks

Secretless broker exposes two endpoints that can be used for readiness and liveliness checks:

- `http://<host>:5335/ready` which will indicate if the broker has loaded a valid configuration.
- `http://<host>:5335/live` which will indicate if the broker has service connectors activated.

If there are failures, the service will return a `503` status or a `200` if the service indicates that
the broker is ready/live.

Note: If Secretless is not provided with a configuration (e.g. it is not listening to anything),
the live endpoint will also return 503.

You can manually check the status with these endpoints by using `curl`:
```
$ # Start Secretless Broker in another terminal on the same machine

$ curl -i localhost:5335/ready
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Thu, 27 Dec 2018 17:12:07 GMT
Content-Length: 3

{}
```

If you would like to retrieve the full informational JSON that includes details on
which checks failed and with what error, you can add the `?full=1` query parameter
to the end of either of the available endpoints:
```
$ # Start Secretless Broker in another terminal on the same machine

$ curl -i localhost:5335/ready?full=1
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Thu, 27 Dec 2018 17:13:22 GMT
Content-Length: 45

{
    "listening": "OK",
    "ready": "OK"
}
```

# Community

Our primary channel for support is through our [Secretless Broker mailing list](https://groups.google.com/forum/#!forum/secretless). More info here: [community support](https://secretless.io/community)

# Performance

Using the Secretless Broker reduces the transaction throughput by about 25% on Postgresql. Once the connection to the backend database is established, the Secretless Broker runs 2 goroutines - one reads from the client and writes to the server, the other reads from the server and writes to the client. It's as simple as this:

```
    go stream(self.Client, self.Backend)
    go stream(self.Backend, self.Client)
```

Here is some performance data created by running [pgbench](https://www.postgresql.org/docs/9.5/static/pgbench.html) with the client, the Secretless Broker and database running on a single machine (2017 MacBook Pro with 2.3 GHz Intel Core i5) where the database is running in Docker.

Directly to the database:

```
$ PGPASSWORD=mysecretpassword PGSSLMODE=disable pgbench -h localhost -U postgres -p 5432 -T 10 -c 12 -j 12 postgres
starting vacuum...end.
transaction type: <builtin: TPC-B (sort of)>
scaling factor: 1
query mode: simple
number of clients: 12
number of threads: 12
duration: 10 s
number of transactions actually processed: 2733
latency average = 44.167 ms
tps = 271.696057 (including connections establishing)
tps = 272.924176 (excluding connections establishing)
```

Through the `secretless-broker` proxy:

```
$ PGSSLMODE=disable pgbench -h localhost -U postgres -p 4321 -T 10 -c 12 -j 12 postgres
starting vacuum...end.
transaction type: <builtin: TPC-B (sort of)>
scaling factor: 1
query mode: simple
number of clients: 12
number of threads: 12
duration: 10 s
number of transactions actually processed: 2459
latency average = 49.034 ms
tps = 244.727719 (including connections establishing)
tps = 248.327635 (excluding connections establishing)
```

From the results above, you can see 9% fewer tps (even including establishing connections) via the Secretless Broker.

To run this test yourself, you can start PostgreSQL by running
```
docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -d -p 5432:5432 postgres:9.3
```

Write a `secretless.yml` file that includes:
```yaml
version: "2"
services:
  pg-db:
    protocol: pg
    listenOn: tcp://0.0.0.0:4321
    credentials:
      host: localhost
      username: postgres
      password: mysecretpassword
      sslmode: disable
```
and run Secretless:
```
$ ./bin/build_darwin # to build the OSX binary
$ ./dist/darwin/amd64/secretless-broker -f secretless.yml
```

Initialize `pgbench` by running
```
PGPASSWORD=mysecretpassword pgbench -i -h localhost -U postgres -p 5432
```
and run the tests as above.

# Development

We welcome contributions of all kinds to the Secretless Broker. For instructions on
how to get started and descriptions of our development workflows, please see our
[contributing guide](CONTRIBUTING.md). This document includes guidelines for
writing plugins to extend the functionality of Secretless Broker.

# License

The Secretless Broker is licensed under Apache License 2.0 - see [`LICENSE.md`](LICENSE.md) for more details.
