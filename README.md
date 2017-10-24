# Secretless

Secretless is a proxy which replaces the native backend authentication of a backend service with Conjur authentication and authorization. The proxied backend can be a database, HTTP web service, or any other service which uses a TCP protocol. 

To provide Secretless access to a backend, a "provider" implements the protocol of the service, replacing the authentication handshake with a scheme based on Conjur access tokens. The client does not need or use a real password to the backend. Rather, it obtains a Conjur access token and uses it as the password. The Secretless provider passes the access token to Conjur and verifies that the client is both authenticated and authorized to access the backend service, which is modeled as a Conjur Webservice. If authentication and authorization is successful, the provider uses a URL, username and password stored in Conjur Variables to obtain a connection to the actual backend. This backend connection is then hooked up to the client, and data is shuttled back and forth between them as fast as possible. 

# Justification

Exposing plaintext database passwords to client code is hazardous from both a security and operational standpoint. First, by providing the password to the client, the client becomes part of the database threat surface. If the application is compromised, then the attacker has a good chance of obtaining the password and being able to establish direct connections to the database. To mitigate the severity of this problem, database passwords are (or should be) rotated on a regular basis. However, rotation introduces the operational problem of keeping the applications up to date with changing passwords. This is a significant problem as many applications only read secrets on startup and are not prepared to handle changing passwords.

When the client connects to the database through a Secretless proxy, we address both of these concerns.

* The client does not have direct access to the password, and therefore cannot reveal it.
* The Secretless proxy is responsible for establishing connections to the backend, and can handle password rotation in a way that's transparent to the client.

# Example

In this example, we setup a client to connect to Postgresql through a Secretless proxy. 

First, a Conjur policy creates the Host identity, the `pg` Webservice and the Variables which store the connection info:

```
# This is the client
- !host myapp

- !policy
  id: pg
  body:
  # Connection parameters to the pg backend
  - &variables
    - !variable url
    - !variable username
    - !variable password

  # Guards connections to the backend
  - !webservice

  # An identity which is used by the proxy
  - !host

  # Permit the proxy to read the connection parameters
  - !permit
    role: !host
    privilege: [ read, execute ]
    resources: *variables

# Permit the client to connect to the database
- !permit
  role: !host myapp
  privileges: [ execute ]
  resource: !webservice pg
```

Load this policy in the normal manner, e.g. `conjur policy load --replace root conjur.yml`.

Next, load the connection parameters:

```
$ conjur variable values add pg/username conjur
$ conjur variable values add pg/password conjur
$ conjur variable values add pg/url pg:5432
```

Now run Postgres on port 5432 in a container called `pg`, and run the Proxy in a container called `proxy`, also on port 5432. The `proxy` container needs environment variables `CONJUR_AUTHN_LOGIN=host/pg` and `CONJUR_AUTHN_API_KEY=<api key of host/pg>`.

The client (`myapp`) will connect to Postgres through the proxy, using `psql`. Of course, `myapp` doesn't have a Postgres password, so it will use a Conjur access token. You can obtain one like this:

```
$ token=$(conjur authn authenticate -H | base64 -w0)
```

Then use it as the password:

```
$ PGPASSWORD="$token" psql -U host/pg -h proxy
psql (9.4.13, server 9.3.19)
Type "help" for help.

postgres=>
```

In the `proxy` log, you can see the proxy authenticating the client with Conjur and then establishing the connection to `pg:5432` using the username and password obtained from Conjur. 

Keep in mind that *all* of the following must be in place for `myapp` to connect to Postgres:

* The proxy is running.
* `host:myapp` has `execute` permission on `webservice:pg`.
* `host:pg` has `execute` permission on the Variables which store the Posgtres parameters.
* The Postgres variables are loaded with valid connection data.

# Performance

Using Secretless reduces the transaction throughput by 28-30% on Postgresql. Once the connection to the backend database is established, the proxy runs 2 goroutines - one reads from the client and writes to the server, the other reads from the server and writes to the client. It's as simple as this:

```
    go stream(self.Client, self.Backend)
    go stream(self.Backend, self.Client)
```

So I am not sure if it can be optimized any further.

Here is some performance data created by running [pgbench](https://www.postgresql.org/docs/9.5/static/pgbench.html) in a Dockerized environment with the client, proxy and database running on a single machine (2017 MacBook Pro with 4-core Intel Core i7 @ 2.9GHz).

Directly to the database:

```
root@566b7c06abcf:/go/src/github.com/kgilpin/secretless-pg# PGPASSWORD=conjur PGSSLMODE=disable pgbench -h pg -U conjur -T 10 -c 12 -j 12 postgres
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
root@566b7c06abcf:/go/src/github.com/kgilpin/secretless-pg# PGPASSWORD=alice PGSSLMODE=disable pgbench -h 172.18.0.9 -U alice -T 10 -c 12 -j 12 postgres
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

Changing the `-c` (number of clients) and `-j` (number of threads) didn't have much effect on the relative throughput, though increasing these from 1 to 12 does approximately double the tps in both direct and proxied scenarios. 
