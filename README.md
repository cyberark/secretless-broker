[![pipeline status](https://gitlab.com/cyberark/secretless-broker/badges/master/pipeline.svg)](https://gitlab.com/cyberark/secretless-broker/commits/master)

***

**Status**: Beta

The Secretless Broker is currently in beta.

Naming and API's are still subject to *breaking* changes.

***

# Table of Contents

- [Secretless Broker](#secretless-broker)
  - [Supported services](#currently-supported-services)
- [Quick Start](#quick-start)
- [Additional demos](#run-more-secretless-demos)
- [Using Secretless](#using-secretless)
  - [Listeners](#listeners)
  - [Handlers](#handlers)
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

To provide the Secretless Broker access to a target service, a [Handler](#handlers) implements the protocol of the service, replacing the authentication handshake. The client does not need to know or use a real password to the service. Instead, it proxies its connection to the service through a local connection to the Secretless Broker via a [Listener](#listeners). The Secretless Broker obtains credentials to the target service from a secrets vault (such as Conjur, a keychain service, text files, or other sources) via a [Credential Provider](#credential-providers). The credentials are used to establish a connection to the actual service, and the Secretless Broker then rapidly shuttles data back and forth between the client and the service.

The Secretless Broker is currently licensed under [ASL 2.0](#license)

## Currently supported services

- MySQL (Socket and TCP)
- PostgreSQL (Socket and TCP)
- SSH
- SSH-Agent
- HTTP (Basic auth, Conjur, and AWS authorization strategies)

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

The quick start image runs the Secretless Broker binary and a few sample target services. The Secretless Broker is configured to retrieve the credentials required to access the services from the process environment. **All services are configured to require authorization to access**, but we don't know what those credentials are. We can try to access the services, but since we don't know the password our access attempts will fail. _But when we try to connect via the Secretless Broker_, we will be granted access.

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
$ psql -h localhost -p 5432 -U postgres -d quickstart
Password for user postgres:
psql: FATAL:  password authentication failed for user "postgres"
```

But the Secretless Broker is listening on port 5454, and will add authentication credentials (both username and password) to our connection request and proxy our connection to the PostgreSQL server:

```sh-session
$ psql \
  -h localhost \
  -p 5454 \
  --set=sslmode=disable \
  -d quickstart
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

The Secretless Broker relies on YAML configuration files to specify which target services it can connect to and how it should retrieve the access credentials to authenticate with those services.

Each Secretless Broker configuration file is composed of two sections:

* `listeners`: A list of protocol Listeners, each one on a Unix socket or TCP port.
* `handlers`: A list of Handlers to process the requests received by each Listener. Handlers implement the protocol for the target services and are configured to obtain the backend connection credentials from one or more Providers.

## Listeners

You can configure the following kinds of Secretless Broker *Listeners*:

1) `unix` The Secretless Broker serves the backend protocol on a Unix domain socket.
2) `tcp` The Secretless Broker serves the backend protocol on a TCP socket.

For example, PostgreSQL clients can connect to the PostgreSQL server either via Unix domain socket or over a TCP connection. If you are setting up the Secretless Broker to facilitate a connection to a PostgreSQL server, you can either configure it:

- to listen on a Unix socket as usual (default: `/var/run/postgresql/.s.PGSQL.5432`)
  ```
  listeners:
  - name: pg_socket
    protocol: pg
    socket: /sock/.s.PGSQL.5432
  ```
  In this case, the client would be configured to connect to the database URL `/sock`.

- to listen on a given port, which may be the PostgreSQL default 5432 or may be a different port to avoid conflicts with the actual PostgreSQL server
  ```
  listeners:
  - name: pg_tcp
    protocol: pg
    address: 0.0.0.0:5432
  ```
  In this case, the client would be configured to connect to the database URL `localhost:5432`

Note that in each case, **the client is not required to specify the username and password to connect to the target service**. It just needs to know where the Secretless Broker is listening, and it connects to the Secretless Broker directly via a local, unsecured connection.

In general, there are currently two strategies to redirect your client to connect to the target service via the Secretless Broker:

1) **Connection URL**

    Connections to the backend service are established by a connection URL. For example, PostgreSQL supports connection URLs such as `postgres://user@password:hostname:port/database`. `hostname:port` can also be a path to a Unix socket, and it can be omitted to use the default PostgreSQL socket `/var/run/postgresql/.s.PGSQL.5432`.

2) **Proxy**

    HTTP services support an environment variable or configuration setting `http_proxy=<url>` which will cause outbound traffic to route through the proxy URL on its way to the destination. The Secretless Broker can operate as an HTTP forward proxy, in which case it will place the proper authorization header on the outbound request. It can also optionally forward the connection using HTTPS. The client should always use plain `http://` URLs, otherwise the Secretless Broker cannot read the network traffic because it will encrypted.

Regardless of the connection strategy, the operating system provides security between the client and the Secretless Broker. It's important to configure the OS properly so that unauthorized processes and clients can't connect to the Secretless Broker. With Unix domain sockets, operating system file permissions protect the socket. With TCP connections, the Secretless Broker should be listening only on localhost.

The Listener configuration governs the _client to Secretless Broker_ connection. The connection from the Secretless Broker to the PostgreSQL server is defined in the Handler configuration, where the actual address and credential information for the connection to the PostgreSQL server is defined.

At this time, the Secretless-Broker-to-target-service connection always happens over TCP by default.

## Handlers

When the Secretless Broker receives a new request on a defined Listener, it automatically passes the request on to the Handler defined in the Secretless Broker configuration for processing. Each Listener in the Secretless Broker configuration should therefore have a corresponding Handler.

The Handler configuration specifies the Listener that the Handler is handling connections for and any credentials that will be needed for that connection. Several credential sources are currently supported; see the [Credential Providers](#credential-providers) section for more information.

In this example, I am setting up a Handler to process connection requests from the `pg_socket` Listener, and it has three credentials: `address`, `username`, and `password`. The `address` and `username` are literally specified in this case, and the `password` is taken from the environment of the running Secretless Broker process.
```
handlers:
  - name: pg_via_socket
    listener: pg_socket
    credentials:
      - name: address
        provider: literal
        id: pg:5432
      - name: username
        provider: literal
        id: myuser
      - name: password
        provider: env
        id: PG_PASSWORD
```

In production you would want your credential information to be pulled from a vault, and the Secretless Broker currently supports multiple vault Credential Providers.

When a Handler receives a new connection requests, it retrieves any required credentials using the specified Provider(s), injects the correct authentication credentials into the connection request, and opens up a connection to the target service. From there, the Handler simply transparently shuttles data between the client and service.

_Please note: Handler API interface signatures are currently under heavy development due to needing to deal with non-overlapping types of communications protocols (as expressed by the interface definitions) so they will be likely to change in the near future._

### Handler Credentials

The Secretless Broker comes with several built-in Handlers, and each accepts a different set of credentials for configuration. In this section we provide information on the credentials used by each Handler.

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
to the Secretless Broker Listeners and Handlers. The Secretless Broker comes built-in with several different
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

### Kubernetes Secrets Provider

Kubernetes Secrets (`kubernetes`) provider allows use of [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) for fetching secrets.

Example:
```
...
    credentials:
      - name: accessToken
        provider: vault
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

### Keychain Provider

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
- `http://<host>:5335/live` which will indicate if the broker has listeners activated.

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

Here is some performance data created by running [pgbench](https://www.postgresql.org/docs/9.5/static/pgbench.html) in a Dockerized environment with the client, the Secretless Broker and database running on a single machine (2017 MacBook Pro with 4-core Intel Core i7 @ 2.9GHz).

Directly to the database:

```
root@566b7c06abcf:/go/src/github.com/cyberark/secretless-broker# PGPASSWORD=test PGSSLMODE=disable pgbench -h pg -U test -T 10 -c 12 -j 12 postgres
starting vacuum...end.
transaction type: TPC-B (sort of)
scaling factor: 1
query mode: simple
number of clients: 12
number of threads: 12
duration: 10 s
number of transactions actually processed: 14434
latency average: 8.327 ms
tps = 1441.077559 (including connections establishing)
tps = 1443.230144 (excluding connections establishing)
```

Through the `secretless-broker` proxy:

```
root@566b7c06abcf:/go/src/github.com/cyberark/secretless-broker# PGSSLMODE=disable pgbench -h 172.18.0.9 -T 10 -c 12 -j 12 postgres
starting vacuum...end.
transaction type: TPC-B (sort of)
scaling factor: 1
query mode: simple
number of clients: 12
number of threads: 12
duration: 10 s
number of transactions actually processed: 10695
latency average: 11.237 ms
tps = 1067.933129 (including connections establishing)
tps = 1075.661082 (excluding connections establishing)
```

Here is a set of test results running directly against an RDS Postgresql:

```
root@2a33637a9cb5:/work# pgbench -h demo1.cb5uzm0ycqol.us-east-1.rds.amazonaws.com -p 5432 -U alice -T 10 -c 12 -j 12 postgres
Password:
starting vacuum...end.
transaction type: TPC-B (sort of)
scaling factor: 1
query mode: simple
number of clients: 12
number of threads: 12
duration: 10 s
number of transactions actually processed: 197
latency average: 657.775 ms
tps = 18.243307 (including connections establishing)
tps = 18.542609 (excluding connections establishing)
```

And to RDS through the Secretless Broker:

```
root@2a33637a9cb5:/work# pgbench -U alice -T 10 -c 12 -j 12 postgres
starting vacuum...end.
transaction type: TPC-B (sort of)
scaling factor: 1
query mode: simple
number of clients: 12
number of threads: 12
duration: 10 s
number of transactions actually processed: 153
latency average: 824.491 ms
tps = 14.554441 (including connections establishing)
tps = 15.822442 (excluding connections establishing)
```

14% fewer tps (excluding establishing connections) via the Secretless Broker.

Changing the `-c` (number of clients) and `-j` (number of threads) didn't have much effect on the relative throughput, though increasing these from 1 to 12 does approximately double the tps in both direct and proxied scenarios.

# Development

We welcome contributions of all kinds to the Secretless Broker. For instructions on
how to get started and descriptions of our development workflows, please see our
[contributing guide](CONTRIBUTING.md). This document includes guidelines for
writing plugins to extend the functionality of Secretless Broker.

# License

The Secretless Broker is licensed under Apache License 2.0 - see [`LICENSE.md`](LICENSE.md) for more details.
