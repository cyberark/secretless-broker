# Secretless

Secretless is a multiplexing server which relieves client applications of the need to directly handle secrets to backend services. The backend service can be a database, HTTP web service, or any other service which uses a TCP protocol. 

To provide Secretless access to a backend service, a "provider" implements the protocol of the backend service, replacing the authentication handshake. The client does not need to know or use a real password to the backend. Instead, it authenticates to the Secretless server. If Secretless decides that the client is both authenticated and authorized to access the backend service, the provider obtains credentials to the Backend service from Conjur Variables. These credentials are used to establish a connection to the actual backend, and the Secretless server then rapidly shuttles data back and forth between the client and backend. 

# Why Secretless?

Exposing plaintext secrets to clients (both users and machines) is hazardous from both a security and operational standpoint. First, by providing a secret to a client, the client becomes part of the threat surface. If the client is compromised, then the attacker has a good chance of obtaining the plaintext secrets and being able to establish direct connections to backend resources. To mitigate the severity of this problem, important secrets are (or should be) rotated (changed) on a regular basis. However, rotation introduces the operational problem of keeping applications up to date with changing passwords. This is a significant problem as many applications only read secrets on startup and are not prepared to handle changing passwords.

When the client connects to a backend resource through Secretles:

* **The client is not part of the threat surface** The client does not have direct access to the password, and therefore cannot reveal it.
* **The client does not have to handle changing secrets** Secretless is responsible for establishing connections to the backend, and can handle secrets rotation in a way that's transparent to the client.

# Example

In this example, we setup a client to connect to Postgresql through Secretless. 

First, a Conjur policy creates a `secretless` Host identity along with Variables which store the database connection parameters:

```
# The Secretless server
- !host secretless

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

# Permit Secretless to read the pg variables
- !grant
  role: !group pg/secrets-users
  member: !host secretless
```

Load this policy in the normal manner, e.g. `conjur policy load --replace root conjur.yml`.

Next, load the connection parameters:

```
$ conjur variable values add pg/username conjur
$ conjur variable values add pg/password conjur
$ conjur variable values add pg/url pg:5432
```

Now run Postgres on port 5432 in a container called `pg`, and run the Secretless in a container called `secretless`, also on port 5432. The `secretless` container needs environment variables `CONJUR_AUTHN_LOGIN=host/secretless` and `CONJUR_AUTHN_API_KEY=<api key of host/secretless>`.

The client container will establish a `psql` connection to Postgresql, connecting through Secretless. Secretless will listen on the standard Postgresql Unix domain socket `/var/run/postgresql/.s.PGSQL.5432`. The client container and Secretless share the socket via Docker volume share:

```
$ psql
psql (9.4.13, server 9.3.19)
Type "help" for help.

postgres=>
```

In the `secretless` log, you can see Secretless establishing the connection to `pg:5432` using the username and password obtained from Conjur Variables. 

# Development

A development environment is provided in the `docker-compose.yml`. 

First build it using `docker-compose build`. Then run Conjur with `docker-compose up -d pg conjur` and obtain the admin API key:

```sh-session
$ docker-compose logs conjur | grep "API key"
conjur_1                    | API key for admin: 2jm5tvn2fmbmme14mfyg83z529kk1vj9x5h25e4m7djx1j7k376nejn
```

Then start a Conjur client container with `docker-compose run --rm client`.

Once in the client, login as `admin` and load the policy and populate the variables:

```sh-session
root@c88091df1304:/# cd ./work/
root@c88091df1304:/# cd ./work/
root@c88091df1304:/work# ./example/conjur.sh
+ conjur policy load root example/conjur.yml
Loaded policy 'root'
{
  "created_roles": {
    ...
    "dev:host:secretless": {
      "id": "dev:host:secretless",
      "api_key": "14y15dx20p3pxg3z99ffw26qqhx01pky7tz2y8jvt3cvxy2c3ebed5y"
    }
  },
  "version": 1
}
+ conjur variable values add pg/username conjur
Value added
+ conjur variable values add pg/password conjur
Value added
+ conjur variable values add pg/url pg:5432
Value added
```

In a new shell, run a `secretless_dev` container with `docker-compose run --rm secretless_dev`. Once in the container, build the Linux binary `./bin/linux/amd64/secretless`:

```sh-session
root@91353a15ccb1:/go/src/github.com/kgilpin/secretless# ./build.sh
+ godep restore
+ go install
+ mkdir -p bin/linux/amd64
+ cp /go/bin/secretless bin/linux/amd64
```

Back in the host shell, build the rest of the container images:

```sh-session
$ docker-compose build
```

Still in the host shell, start the `secretless_test` container:

```sh-session
$ export CONJUR_AUTHN_API_KEY=<API key of dev:host:secretless>
$ docker-compose up secretless_test
```

Back in the `secretless_dev` container, you are ready to run the tests:

```
# export CONJUR_AUTHN_API_KEY=<API key of dev:user:admin>
# godep restore
...
# go test
2017/10/27 14:17:07 Provide a statically configured password
2017/10/27 14:17:07 Provide the wrong value for a statically configured password
2017/10/27 14:17:07 Provide a Conjur access token as the password
2017/10/27 14:17:07 Provide a Conjur access token for an unauthorized user
PASS
ok    github.com/kgilpin/secretless  0.318s
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
* **test_in_container.sh** A helper script which you probably won't run directly.
