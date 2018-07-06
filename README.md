***

**Status**: Preview

Secretless is currently a technology preview, suitable for demo and evaluation purposes.

***

- [Secretless](#secretless)
- [Why Secretless?](#why-secretless)
- [Quick Start](#quick-start)
- [Longer Example](#longer-example-with-docker-images)
  - [Walkthrough](#walkthrough)
- [Configuring Secretless](#configuring-secretless)
  - [Listeners](#listeners)
  - [Handlers](#handlers)
- [Plugins](#plugins)
- [Client Application Configuration](#client-application-configuration)
- [Testing](#testing)
- [Performance](#performance)

# Secretless

Secretless is a connection broker which relieves client applications of the need to directly handle secrets to backend services such as databases, web services, SSH connections, or any other TCP-based service. 

To provide Secretless access to a backend service, a "provider" implements the protocol of the backend service, replacing the authentication handshake. The client does not need to know or use a real password to the backend. Instead, it proxies its connection to the backend through Secretless. Secretless obtains credentials to the backend service from a secrets vault such as Conjur, a keychain service, text files, or other sources. The credentials are used to establish a connection to the actual backend, and the Secretless server then rapidly shuttles data back and forth between the client and the backend.

# Currently supported

- MySQL (Socket)
- PostgreSQL (Socket and TCP)
- SSH
- SSHAgent (OSX and Linux keyrings)
- HTTP (Conjur and AWS auth providers)

With many others in plans!

# Why Secretless?

Exposing plaintext secrets to clients (both users and machines) is hazardous from both a security and operational standpoint. First, by providing a secret to a client, the client becomes part of the threat surface. If the client is compromised, then the attacker has a good chance of obtaining the plaintext secrets and being able to establish direct connections to backend resources. To mitigate the severity of this problem, important secrets are (or should be) rotated (changed) on a regular basis. However, rotation introduces the operational problem of keeping applications up to date with changing passwords. This is a significant problem as many applications only read secrets on startup and are not prepared to handle changing passwords thus requiring restarts when they change.

When the client connects to a backend resource through Secretless:

* **The client is not part of the threat surface** The client does not have direct access to the password, and therefore cannot reveal it.
* **The client does not have to know how to properly manage secrets** Handling secrets safely is very difficult. When every application needs to know how to handle secrets, accidents happen. Secretless centralizes the client-side management of secrets into one code base.
* **The client does not have to handle changing secrets** Secretless is responsible for establishing connections to the backend, and can handle secrets rotation in a way that's transparent to the client.

# Quick Start

**Prerequisites**

* **Docker** You need Docker to run the examples and the tests. Most code has been confirmed to work on Docker version 18.03.1-ce and up.
* **PostgreSQL** You need PostgreSQL client (`psql`) on your host to fully run this example
* **Linux** This example assumes a Linux amd64 environment

## Download Secretless

```sh-session
$ wget https://github.com/conjurinc/secretless/releases/download/v<version>/secretless-linux-amd64.tar.gz
$ tar -xzf ./secretless-linux-amd64.tar.gz
$ ls
secretless  secretless-linux-amd64.tar.gz  summon2
```

## Start PostgreSQL container on localhost:6543
```sh-session
$ sudo docker run --rm \
                  --name pgsql \
                  -p 6543:5432 \
                  -e POSTGRES_PASSWORD=test \
                  -d postgres:alpine
```

## Create a minimal configuration for Secretless

```sh-session
$ # We will listen on 15432 and forward requests to our
$ # PostgreSQL container on localhost:6543
$ tee secretless.yml <<EOF
listeners:
  - name: pg
    protocol: pg
    address: 0.0.0.0:15432

handlers:
  - name: pg
    listener: pg
    credentials:
      - name: address
        provider: literal
        id: localhost:6543
      - name: username
        provider: literal
        id: postgres
      - name: password
        provider: literal
        id: test
EOF
```
Normally, you would pull the credentials from a real vault with one of the included providers but for the quick start, we're just adding the secrets directly.

## (Optional) Verify that PostgreSQL requires credentials

```sh-session
$ # This should only allow you to log in if you type in `test` as the password
$ psql -h localhost -p 6543 -U postgres -d postgres
Password for user postgres: 
...
postgres-# \q
```

## Start Secretless
```sh-session
$ ./secretless
```

## Check that you can connect to PostgreSQL over Secretless without a password
In a different terminal, connect to Secretless listening port with `psql`:
```sh-session
$ # You may need to install psql client tools on your host to do this step (`sudo apt install postgresql-client-10`).
$ # Note that there was _no_ request for password to be able to log in!
$ psql -h localhost -p 15432 -d postgres
psql (10.4 (Ubuntu 10.4-0ubuntu0.18.04))
Type "help" for help.

postgres=# select current_user;
 current_user 
--------------
 postgres
(1 row)

```

## Smile and grab a :cookie: because it was too easy!

You have just delegated responsibilty for keeping credentials secure away from your app!

# Longer example with Docker images

**Prerequisites**

* **Docker** You need Docker to run the examples and the tests.

## Walkthrough
--------------------------------
Navigate to the directory `doc/quick-demo`:

```sh-session
$ # From Secretless repository root 
$ cd doc/quick-demo
```

You will use Secretless to connect a client to a Postgresql database, without the client knowing the database password.

Start by running a Postgresql server using `docker-compose`:

```sh-session
$ docker-compose up -d pg
Creating network "quick_default" with the default driver
Creating quick_pg_1 ...
Creating quick_pg_1 ... done
```

Verify that PostgreSQL is running and accepting connections on port 5432:

```
$ docker-compose ps
   Name                 Command              State            Ports
----------------------------------------------------------------------------
quick_pg_1   docker-entrypoint.sdh postgres   Up      5432/tcp
```

Now you can test a normal connection to Postgresql in which the client knows the password. Start a `psql` container:

```sh-session
$ docker-compose run --rm psql
Starting quick_psql_1 ... done
root@f6683931b82c:/#
```

Now connect to PostgreSQL using the username "test" and password "test" (type `\q` to quit):

```sh-session
root@f6683931b82c:/# PGPASSWORD=test PGUSER=test PGPORT=5432 PGHOST=pg PGDATABASE=postgres psql
psql (9.5.10, server 9.3.20)
Type "help" for help.

postgres=> \q
```

Fine, this is the normal way of connecting to Postgresql. Now let's see how to connect a client to the database without knowing the password. 

You'll use a YAML file to tell Secretless the following information:

* Listen on Unix socket `/run/postgresql/s.PGSQL.5432` for client connections.
* Route client connections on that socket to the `pg` handler.
* The `pg` handler should obtain the database address, username and password from environment variables.

Here's what this [secretless.yml](doc/quick/secretless.yml) looks like:

```yaml
listeners:
  - name: pg
    socket: ./run/postgresql/.s.PGSQL.5432

handlers:
  - name: pg
    listener: pg
    credentials:
      - name: address
        provider: env
        id: PG_ADDRESS
      - name: username
        provider: env
        id: PG_USER
      - name: password
        provider: env
        id: PG_PASSWORD
```

In real world scenarios, the credentials (secrets) can be obtained from a secrets vault or operating system keychain.

Run `secretless` using `docker-compose`:

```sh-session
$ docker-compose up -d secretless
quick_pg_1 is up-to-date
Creating quick_secretless_1 ...
Creating quick_secretless_1 ... done
```

Verify that Secretless is up and listening:

```sh-session
$ docker-compose logs secretless
Attaching to quick_secretless_1
secretless_1  | 2018/01/09 20:37:15 Secretless starting up...
secretless_1  | 2018/01/09 20:37:15 Loaded configuration : {[] [{pg   ./run/postgresql/.s.PGSQL.5432 []}] [{pg  pg {false  map[]} false [] [] []}]}
secretless_1  | 2018/01/09 20:37:15  listener 'pg' listening at: ./run/postgresql/.s.PGSQL.5432
```

Now start another `psql` container:

```sh-session
$ docker-compose run --rm psql
Starting quick_pg_1 ... done
root@2fdd8fa01ef2:/#
```

In the directory `/run/postgresql/` you'll see a socket file where Secretless is listening:

```sh-session
root@2fdd8fa01ef2:/# ls -al /run/postgresql/
total 4
drwxr-xr-x 4 root root  136 Jan  9 20:37 .
drwxr-xr-x 1 root root 4096 Jan  9 20:38 ..
-rw-r--r-- 1 root root    0 Jan  9 20:19 .keep
srwxrwxrwx 1 root root    0 Jan  9 20:37 .s.PGSQL.5432
```

This is the default location of the Postgresql server socket. Keep in mind, it's not actually Postgresql that's listening on this socket, it's Secretless.

You can now establish a secretless connection to Postgresql:

```sh-session
root@ae57550f7e95:/# psql postgres
psql (9.5.10, server 9.3.20)
Type "help" for help.

postgres=> 
```

Issue the SQL command `select * from test` to list the rows in the `test` table:

```sh-session
postgres=> select * from test;
 id
----
  1
  2
(2 rows)
```

That's it! You connected a client through Secretless to Postgresql. 

Note that a real-world deployment would differ from this setup in the following ways:

* The backend service (e.g. Postgresql) would be running remotely on the network.
* The backend service credentials would be stored in a secrets vault.
* `secretless.yml` would configure the authentication credentials to the vault.
* `secretless.yml` might contain listeners and handlers for other backend services, such as SSH and/or HTTP web services.

# Configuring Secretless

The Secretless configuration file is composed of two sections:

* `listeners` A list of protocol listeners, each one on a Unix socket or TCP port.
* `handlers` When a new connection is received by a Listener, it's routed to a Handler for processing. The Handler is configured to obtain the backend connection credentials from one or more Providers. 

## Listeners

You can configure the following kinds of Secretless *Listeners*:

1) `unix` Secretless serves the backend protocol on a Unix domain socket.
2) `tcp` Secretless serves the backend protocol on a TCP socket.

When Secretless is managing a backend service that supports Unix domain socket connections, it's best to have the client establish the connection directly to the Unix socket.

For example, Postgresql clients can connect to the Postgresql server on a Unix domain socket (default: `/var/run/postgresql/.s.PGSQL.5432`). Configure Secretless to listen on this socket, and configure the client with the database URL `/var/run/postgresql`.

Alternatively, Secretless can listen on a TCP port, and the client can connect to that port. 

To use the Postgresql example again, the Postgresql server listens by default on port 5432. Configure Secretless to listen on port 5432, and configure the client with the database URL `localhost:5432`.

To configure Secretless to broker web service connections, configure Secretless with a TCP listener on a well-known port such as `1080`. 

Then set the environment variable `http_proxy=localhost:1080` in the client environment. Ensure that the client sends HTTP and not HTTPS requests (TLS can be added by Secretless). 

## Handlers

Handlers are objects that get instantiated on each connection to a listener that provide connectivity:
- Downstream to the proxy server
- Proxy server to upstream server

As part of this functionality, they also modify traffic to inject credentials for the connection to the upstream server - but the majority of their functionality is in simple shuttling data between downstream and upstream in a transparent manner.

_Please note: Handler API interface signatures are currently under heavy development due to needing to deal with non-overlapping types of communications protocols (as expressed by the interface definitions) so they will be likely to change in the near future._

# Plugins

Plugins can be used to extend the functionality of Secretless via a shared library in `/usr/local/lib/secretless` by providing a way to add additional:
- Listener plugins
- Handler plugins
- Connection management plugins

Currently, these API definitions reside [here](pkg/secretless/plugin/v1) and an example plugin can be found in the [`test/plugin`](test/plugin) directory.

You can read more about how to make plugins and the underlying architecture in the [API directory](pkg/secretless/plugin).

_Please note: Plugin API interface signatures and supported plugin API version(s) are currently under heavy development so they will be likely to change in the near future._


# Client Application Configuration

You need to ensure that when your client code connects to a backend service, the connection is routed through Secretless. The way that you do this depends on what kind of backend the client is connecting to: Postgresql database, HTTP web service, etc. Generally, there are two strategies:

1) **Connection URL** Connections to the backend service are established by a connection URL. For example, Postgresql supports connection URLs such as `postgres://user@password:hostname:port/database`. `host:port` can also be a path to a Unix socket, and it can be omitted to use the default Postgresql socket `/var/run/postgresql/.s.PGSQL.5432`.
2) **Proxy** HTTP services support an environment variable or configuration setting `http_proxy=<url>` which will cause outbound traffic to route through the proxy URL on its way to the destination. Secretless can operate as an HTTP forward proxy, in which case it will place the proper authorization header on the outbound request. It can also optionally forward the connection using HTTPS. The client should always use plain `http://` URLs, otherwise Secretless cannot read the network traffic because it will encrypted.  

In all cases, the operating system provides security between the client and Secretless. It's important to configure the OS properly so that unauthorized processes and clients can't connect to Secretless. With Unix domain sockets, operating system file permissions protect the socket. With TCP connections, Secretless should be listening only on localhost.


# Building

First, clone `https://github.com/conjurinc/secretless`. If you're new to Go, be aware that Go is very selective about
where the files are placed on the filesystem. There is an environment variable called `GOPATH`, whose default value
is `~/go`. Projects should be checked out to `$GOPATH/src`. This is required by Go in order for dependencies to resolve
properly. So after you clone, the source code should be located in `$GOPATH/src/github.com/conjurinc/secretless`.

Now you can build Secretless. First fetch all the dependencies:

```sh-session
~ $ cd $GOPATH/src/github.com/conjurinc/secretless
secretless $ dep ensure
```

## Docker containers

```sh-session
$ # From Secretless repository root 
$ ./bin/build
```

This should create a Docker container with tag `secretless:latest` in your local registry.

## Binaries
### Linux
```sh-session
$ # From Secretless repository root 
$ go build -o ./secretless ./cmd/secretless
```

### OSX

```sh-session
$ # From Secretless repository root 
$ ./bin/build_darwin
```

# Testing

**Prerequisites**

* **Docker** You need Docker to run the tests.

Build the project by running:

```sh-session
$ ./bin/build
```

Then run the test cases:

```sh-session
$ ./bin/test
```

# Performance

Using Secretless reduces the transaction throughput by about 25% on Postgresql. Once the connection to the backend database is established, Secretless runs 2 goroutines - one reads from the client and writes to the server, the other reads from the server and writes to the client. It's as simple as this:

```
    go stream(self.Client, self.Backend)
    go stream(self.Backend, self.Client)
```

Here is some performance data created by running [pgbench](https://www.postgresql.org/docs/9.5/static/pgbench.html) in a Dockerized environment with the client, Secretless and database running on a single machine (2017 MacBook Pro with 4-core Intel Core i7 @ 2.9GHz).

Directly to the database:

```
root@566b7c06abcf:/go/src/github.com/conjurinc/secretless# PGPASSWORD=test PGSSLMODE=disable pgbench -h pg -U test -T 10 -c 12 -j 12 postgres
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

Through the `secretless` proxy:

```
root@566b7c06abcf:/go/src/github.com/conjurinc/secretless# PGSSLMODE=disable pgbench -h 172.18.0.9 -T 10 -c 12 -j 12 postgres
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

And to RDS through Secretless:

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

14% fewer tps (excluding establishing connections) via Secretless.

Changing the `-c` (number of clients) and `-j` (number of threads) didn't have much effect on the relative throughput, though increasing these from 1 to 12 does approximately double the tps in both direct and proxied scenarios. 
