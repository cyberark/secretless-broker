# Postgresql Handler Development

## Quick Test (Laptop environment)

Run Postgres in Docker:

```sh-session
$ docker-compose up -d pg
```

Run Secretless locally and execute tests:

```sh-session
$ ./run_dev_test
...
ok      github.com/conjurinc/secretless/test/pg_handler   0.048s
2018/01/11 15:06:56 Caught signal terminated: shutting down.
```

## Local Environment (Laptop)

These instructions show how to develop the Secretless PG handler on you local machine. This way you can use niceties such as IDE features.

First you'll need a Postgres server. You can run one natively, or using Docker:

```sh-session
$ docker-compose up -d pg
```

Now you can run `secretless` in a terminal:

```sh-session
$ ./run_dev
...
2018/01/10 16:33:09 pg listener 'pg_tcp' listening at: [::]:15432
2018/01/10 16:33:09 pg listener 'pg_socket' listening at: ./run/postgresql/.s.PGSQL.5432
```

Now run a client in another terminal.

Connect over a Unix socket:

```sh-session
$ psql -h $PWD/run/postgresql postgres
psql (9.6.5, server 9.3.20)
Type "help" for help.

postgres=> \q
```

Connect over TCP:

```sh-session
$ PGSSLMODE=disable psql -p 15432 -h localhost postgres
psql (9.6.5, server 9.3.20)
Type "help" for help.

postgres=> \q
```

### Docker-hosted Environment

You can also develop in Docker. This option doesn't require any Go tools on your local machine. 

First, run `pg`:

```sh-session
$ docker-compose up -d pg
```

Then run a `dev` container:

```sh-session
$ docker-compose run --rm dev
Starting pghandler_pg_1 ... done
secretless # cd test/pg_handler
pg_handler # 
```

Now you can run the secretless server:

```sh-session
pg_handler# PG_ADDRESS=pg:5432 \
  ../../bin/linux/amd64/secretless \
  -config secretless.dev.yml
2018/01/10 21:25:15 Secretless starting up...
...
2018/01/10 21:25:15 pg listener 'pg_tcp' listening at: [::]:15432
2018/01/10 21:25:15 pg listener 'pg_socket' listening at: ./run/postgresql/.s.PGSQL.5432
```

Now run another `dev` container as the client:

```sh-session
$ docker-compose run --rm dev
Starting pghandler_pg_1 ... done
secretless# cd test/pg_handler/
pg_handler#
```

Connect to Postgres using psql, over a Unix socket:

```sh-session
pg_handler# psql -h $PWD/run/postgresql/ postgres
psql (9.4.15, server 9.3.20)
Type "help" for help.

postgres=> \q
```

And over TCP (note: you'll need the IP address of the `secretless` container):

```sh-session
pg_handler# PGSSLMODE=disable psql -p 15432 -h 172.23.0.3 postgres
psql (9.4.15, server 9.3.20)
Type "help" for help.

postgres=> \q
```
