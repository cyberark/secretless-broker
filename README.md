# Table of Contents

- [Secretless Broker](#secretless-broker)
  - [Supported services](#currently-supported-services)
- [Quick Start](#quick-start)
- [Additional demos](#run-more-secretless-broker-demos)
- [Using Secretless](#using-secretless)
  - [Using This Project With Conjur-OSS](#using-secretless-broker-with-conjur-open-source)
  - [About our releases](#about-our-releases)
- [Community](#community)
- [Performance](#performance)
- [Development](#development)
- [License](#license)


# Secretless Broker&trade;

Secretless Broker is a connection broker which relieves client applications of the need to directly handle secrets to target services such as databases, web services, SSH connections, or any other TCP-based service.

![Secretless Broker Architecture](https://github.com/cyberark/secretless-broker/blob/main/docs/img/secretless_architecture.svg)

Secretless is designed to solve two problems. The first is **loss or theft of credentials from applications and services**, which can occur by:

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

  Secretless is responsible for establishing connections to the backend, and can handle secret rotation in a way that’s transparent to the client.

To provide Secretless access to a target service, a [Service Connector](#service-connectors) implements the protocol of the service, replacing the authentication handshake. The client does not need to know or use a real password to the service. Instead, it proxies its connection to the service through a local connection to Secretless. Secretless obtains credentials to the target service from a secrets vault (such as Conjur, a keychain service, text files, or other sources) via a [Credential Provider](#credential-providers). The credentials are used to establish a connection to the actual service, and Secretless then rapidly shuttles data back and forth between the client and the service.

Secretless Broker is currently licensed under [ASL 2.0](#license)

## Currently supported services

Secretless supports several target services out of the box, and these include:

- MySQL (Socket and TCP)
- PostgreSQL (Socket and TCP)
- SSH / SSH-Agent (Beta)
- HTTP with Basic auth, Conjur, and AWS authorization strategies (Beta)

Support for these services is provided via internal plugins (also referred to as "Service Connectors") that are part
of the Secretless binary distribution.

If you want to use Secretless with a target service that is not currently supported, you can use the [Secretless Plugin
Interface](pkg/secretless/plugin) to create [Connector Plugins](pkg/secretless/plugin/connector) to extend Secretless
to support virtually any target service. These external plugins can be integrated in environments using a standard
Secretless Broker implementation.

For more information on building a Secretless Connector Plugin please see our [documentation](https://docs.secretless.io/Latest/en/Content/References/Secretless%20Plugin%20Interface/scl_Secretless_Plugin_Interface_Intro.htm),
which will walk you through creating a new Connector Plugin using our templates.

Are we missing an internal service connector that you think is important? Are you curious if anyone else has thought of
building the service connector you're interested in? Please open a [GitHub issue](https://github.com/cyberark/secretless-broker/issues) with more information on the connector you'd like, and start the conversation.

For specific guidelines about using a particular service, please see our [documentation](https://docs.secretless.io/Latest/en/Content/References/connectors/scl_connectors_overview.htm).

# Quick Start

Running the quick start demo requires [Docker][get-docker].

[get-docker]: https://docs.docker.com/engine/installation

To see the Secretless Broker in action, pull the quick start image:

```sh-session
docker pull cyberark/secretless-broker-quickstart
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

If the PostgreSQL quick start demo piqued your interest, please check out our [additional demos](https://github.com/cyberark/secretless-broker/tree/main/demos/quick-start#ssh-quick-start) where you can try the Secretless Broker with SSH and HTTP Basic Auth.

For an even more in-depth demo, check out our [Deploying to Kubernetes](https://github.com/cyberark/secretless-broker/tree/main/demos/k8s-demo) demo, which walks you through deploying a sample app to Kubernetes with the Secretless Broker.

# Using Secretless

For complete documentation on using Secretless, please see [our documentation](https://docs.secretless.io/Latest/en/Content/Resources/_TopNav/cc_Home.htm). The documentation includes comprehensive guides for how to get up and running with Secretless.

## Using secretless-broker with Conjur Open Source

Are you using this project with
[Conjur Open Source](https://github.com/cyberark/conjur)?
Then we **strongly** recommend choosing the version of this project to use from
the latest [Conjur OSS suite release](https://docs.conjur.org/Latest/en/Content/Overview/Conjur-OSS-Suite-Overview.html).
Conjur maintainers perform additional testing on the suite release versions to ensure
compatibility. When possible, upgrade your Conjur version to match the
[latest suite release](https://docs.conjur.org/Latest/en/Content/ReleaseNotes/ConjurOSS-suite-RN.htm);
when using integrations, choose the latest suite release that matches your Conjur version. 
For any questions, please contact us on [Discourse](https://discuss.cyberarkcommons.org/c/conjur/5).

## About our releases

### Docker images
The primary source of Secretless Broker releases is our [DockerHub](https://hub.docker.com/r/cyberark/secretless-broker/tags/).

Every new GitHub tag (`x.y.z`) added to the project produces a set of Docker images in DockerHub: (`x.y.z`, `x.y`, `x`, `latest`). The `latest` image tag therefore always corresponds to the latest GitHub tag, and the `x.y` and `x` tags alway correspond to the latest matching `x.y.z` image.

### GitHub releases
Post-1.0, GitHub releases are created for GitHub tagged versions that have undergone additional quality activities to ensure that Secretless continues to meet its [stability requirements](#stable-release-definition). When a GitHub release is created (indicating a new stable Secretless version), an additional Docker image is pushed to DockerHub with the `stable` tag.

At any given time there is only one Docker image with the `stable` tag - that of the latest `stable` release. Older stable versions (v1.0+) can be found on the official [GitHub releases page](https://github.com/cyberark/secretless-broker/releases).

**Note:** the [GitHub releases page](https://github.com/cyberark/secretless-broker/releases) will also show `pre-releases`, which have not yet been promoted to stable.

### GitHub repository
The code on `main` in the project's GitHub repository represents the development work done since the previous GitHub tag. It is possible to build Secretless from source (see [our contributing guidelines](#development) for more info), but for regular use we recommend using the `stable` Docker image from DockerHub or an official GitHub release.

### Stable release definition
Stable components (eg service connectors, credential providers, etc.) of Secretless meet the core acceptance criteria:
- The component should perform its functionality transparently to the underlying application
- The component must guard against threats from all parts of the [STRIDE threat model](https://en.wikipedia.org/wiki/STRIDE_(security))
- Documentation exists that clearly explains how to set up and use the component as well as providing troubleshooting information for anticipated common failure cases
-  A suite of automated tests that exercise the component exists and provides excellent code coverage

In addition, the following must be true for any stable release:
- Secretless has had security review (including static code analysis), and all known high and critical issues have been addressed. Any low or medium issues that have not been addressed have been logged in the [GitHub issue backlog](https://github.com/cyberark/secretless-broker/issues) with a label of the form `security/X`
- Secretless has undergone [STRIDE threat modeling](https://en.wikipedia.org/wiki/STRIDE_(security))
- For use cases involving stable components of Secretless (eg service connectors, credential providers, etc.):
  -	Secretless is stable while running
    -	It does not drop connections while running
    -	It can run without failure (eg it can consistently open connections, it maintains existing open connections, etc) for an extended period of time (~4 days)
    -	When the Secretless process dies in Kubernetes / OpenShift, the pod is destroyed and rescheduled
  -	Secretless has minimal impact on connection speeds
    -	Request throughput is within 20% of direct-to-DB speed
  -	Secretless performs under load in a realistic environment (eg deployed to the same pod as an application in Kubernetes or OpenShift)
  -	Secretless handles connections transparently, so that requests from the client are appropriately transmitted to the server and messages from the server are propagated back to the client
-	Secretless is easy to set up, and adding the configured Secretless Broker sidecar to your application takes less than 30 minutes
-	Secretless is clear about known limitations

# Community

Our primary channel for support is through our [Secretless Broker Discourse](https://discuss.cyberarkcommons.org/c/secretless-broker). More info here: [community support](https://secretless.io/community)

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
    connector: pg
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

We welcome contributions of all kinds to Secretless Broker. For instructions on
how to get started and descriptions of our development workflows, please see our
[contributing guide](CONTRIBUTING.md). This document includes guidelines for
writing plugins to extend the functionality of Secretless Broker.

# License

The Secretless Broker is licensed under Apache License 2.0 - see [`LICENSE`](LICENSE) for more details.
