# Secretless

Secretless is a connection broker which relieves client applications of the need to directly handle secrets to backend services such as databases, web services, SSH connections, or any other TCP-based service. 

To provide Secretless access to a backend service, a "provider" implements the protocol of the backend service, replacing the authentication handshake. The client does not need to know or use a real password to the backend. Instead, it proxies its connection to the backend through Secretless. Secretless obtains credentials to the Backend service from a secrets vault such as Conjur, a keychain service, or text files. The credentials are used to establish a connection to the actual backend, and the Secretless server then rapidly shuttles data back and forth between the client and the backend. 

# Why Secretless?

Exposing plaintext secrets to clients (both users and machines) is hazardous from both a security and operational standpoint. First, by providing a secret to a client, the client becomes part of the threat surface. If the client is compromised, then the attacker has a good chance of obtaining the plaintext secrets and being able to establish direct connections to backend resources. To mitigate the severity of this problem, important secrets are (or should be) rotated (changed) on a regular basis. However, rotation introduces the operational problem of keeping applications up to date with changing passwords. This is a significant problem as many applications only read secrets on startup and are not prepared to handle changing passwords.

When the client connects to a backend resource through Secretless:

* **The client is not part of the threat surface** The client does not have direct access to the password, and therefore cannot reveal it.
* **The client does not have to know how to properly manage secrets** Handling secrets safely is very difficult. When every application needs to know how to handle secrets, accidents happen. Secretless centralizes the client-side management of secrets into one code base.
* **The client does not have to handle changing secrets** Secretless is responsible for establishing connections to the backend, and can handle secrets rotation in a way that's transparent to the client.

# Getting Started

For now, you'll need to build Secretless in order to use it. 

Clone `https://github.com/conjurinc/secretless` and then build it:

```sh-session
$ dep ensure
$ go build ./cmd/secretless
```

Next, create a configuration file written in YAML which tells Secretless:

* Which ports and sockets to listen on.
* Which secrets providers to use and how to configure them.
* Which handlers to use for inbound requests.

For example [demo/secretless.myapp.yml](demo/secretless.myapp.yml).

Now run `secretless`:

```sh-session
$ ./secretless -config demo/secretless.myapp.yml
```

Secretless is listening on the ports and sockets that you requested. 

Now you need to ensure that when your client code connects to a backend service, the connection is routed through Secretless. The client connects to Secretless in one of the following ways:

1) Secretless serves the backend protocol on a Unix domain socket.
2) Secretless serves the backend protocol on a TCP socket.
3) Secretless serves as an HTTP forward proxy.

In all cases, the operating system provides security between the client and Secretless. It's important to configure the OS properly so that unauthorized processes and clients can't connect to Secretless. With Unix domain sockets, operating system file permissions protect the socket. With TCP connections, Secretless should be listening only on localhost.

## Unix Domain Socket

When Secretless is managing a backend service that supports Unix domain socket connections, it's best to have the client establish the connection directly to the Unix socket.

For example, Postgresql clients can connect to the Postgresql server on a Unix domain socket (default: `/var/run/postgresql/.s.PGSQL.5432`). Configure Secretless to listen on this socket, and configure the client with the database URL `/var/run/postgresql`.

## TCP Protocol

Alternatively, Secretless can listen on a TCP port, and the client can connect to that port. 

To use the Postgresql example again, the Postgresql server listens by default on port 5432. Configure Secretless to listen on port 5432, and configure the client with the database URL `localhost:5432`.

## HTTP Forward Proxy

To configure Secretless to broker web service connections, configure Secretless with a TCP listener on a well-known port such as `1080`. 

Then set the environment variable `http_proxy=localhost:1080` in the client environment. Ensure that the client sends HTTP and not HTTPS requests (TLS can be added by Secretless). 

# Detailed Example

Here's how to setup a client to connect to a database (Postgresql) through Secretless. 

First, store database connection parameters in a vault (we'll use Conjur):

```
- !host myapp

- !policy
  id: pg
  body:
  # Connection parameters to the pg backend
  - &variables
    - !variable url
    - !variable username
    - !variable password

  - !group secrets-users

  # Permit the proxy to read the connection parameters
  - !permit
    role: !group secrets-users
    privilege: [ read, execute ]
    resources: *variables

# Permit the application to read the pg variables
- !grant
  role: !group pg/secrets-users
  member: !host myapp
```

Load this policy in the normal manner, e.g. `conjur policy load --replace root conjur.yml`.

Next, load the connection parameters:

```
$ conjur variable values add pg/username conjur
$ conjur variable values add pg/password conjur
$ conjur variable values add pg/url pg:5432
```

Now run Postgres on port 5432 in a container called `pg`, and run the Secretless in a container called `secretless`, also on port 5432. The `secretless` container needs environment variables `CONJUR_AUTHN_LOGIN=host/myapp` and `CONJUR_AUTHN_API_KEY=<api key of host/myapp>`.

The client container will establish a `psql` connection to Postgresql, connecting through Secretless. Secretless will listen on the standard Postgresql Unix domain socket `/var/run/postgresql/.s.PGSQL.5432`. The client container and Secretless share the socket via Docker volume share:

```
$ psql
psql (9.4.13, server 9.3.19)
Type "help" for help.

postgres=>
```

In the `secretless` log, you can see Secretless establishing the connection to `pg:5432` using the username and password obtained from Conjur Variables. 

# Testing

You'll need Docker to run the test cases.

Build the project by running:

```sh-session
$ ./build/build.sh
```

Or on OS X:

```sh-session
$ ./build/build_darwin.sh
```

Then run the test cases:

```sh-session
$ ./build/test.sh
```

# Performance

Using Secretless reduces the transaction throughput by 28-30% on Postgresql. Once the connection to the backend database is established, Secretless runs 2 goroutines - one reads from the client and writes to the server, the other reads from the server and writes to the client. It's as simple as this:

```
    go stream(self.Client, self.Backend)
    go stream(self.Backend, self.Client)
```

Here is some performance data created by running [pgbench](https://www.postgresql.org/docs/9.5/static/pgbench.html) in a Dockerized environment with the client, Secretless and database running on a single machine (2017 MacBook Pro with 4-core Intel Core i7 @ 2.9GHz).

Directly to the database:

```
root@566b7c06abcf:/go/src/github.com/kgilpin/secretless# PGPASSWORD=conjur PGSSLMODE=disable pgbench -h pg -U conjur -T 10 -c 12 -j 12 postgres
starting vacuum...end.
transaction type: TPC-B (sort of)
scaling factor: 1
query mode: simple
number of clients: 12
number of threads: 12
duration: 10 s
number of transactions actually processed: 13371
latency average: 8.988 ms
tps = 1335.119350 (including connections establishing)
tps = 1337.527786 (excluding connections establishing)
```

Through the `secretless` proxy:

```
root@566b7c06abcf:/go/src/github.com/kgilpin/secretless# PGPASSWORD=alice PGSSLMODE=disable pgbench -h 172.18.0.9 -U alice -T 10 -c 12 -j 12 postgres
starting vacuum...end.
transaction type: TPC-B (sort of)
scaling factor: 1
query mode: simple
number of clients: 12
number of threads: 12
duration: 10 s
number of transactions actually processed: 9622
latency average: 12.502 ms
tps = 959.835445 (including connections establishing)
tps = 962.082570 (excluding connections establishing)
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

14% fewer tps (excluding establishing connections) via Secretless. Establishing connections takes a relatively long time because the credentials are being looked up in Conjur. These can be cached in Secretless as a future optimization.

Changing the `-c` (number of clients) and `-j` (number of threads) didn't have much effect on the relative throughput, though increasing these from 1 to 12 does approximately double the tps in both direct and proxied scenarios. 

# Continuous Integration

In the project root directory there is a Makefile. Run `make all` to build and test Secretless. Binaries are built to the `./bin` directory.

The `./build` directory contains CI scripts:

* **build.sh** Builds the Go binaries. Expects to run in a Linux environment. 
* **test.sh** Runs the test suite. Expects to run in an environment with `docker-compose` available, and expects Go binaries to be built.
* **test\_in_container.sh** A helper script which you probably won't run directly.
