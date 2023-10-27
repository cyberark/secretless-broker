# PostgeSQL Connector Development

## Building

From the project root directory, build the Secretless Broker containers.

```sh
./bin/build
```

## Testing

### Run the tests in Docker

To run the test suite in Docker, run:

```sh
./start  # Stand up Postgres and Secretless Broker servers
./test   # Run tests in a test container
./stop   # Clean up all running containers
```

## Developing

### Start the development environment

Run `./dev` from this directory. This will automatically start a Postgres server
in a Docker container at `localhost:5432`. It will
also configure the Postgres server as follows:

- Create a `test` user (with password `test`)
- Create a table `test` in the `test` schema and add bunch of rows

It will also start a container running Secretless Broker and a test container
that can be used to send requests to the Postgres server via Secretless Broker.

Note: When you run `./dev`, it will start Secretless Broker wtih two services:
(See [fixtures/secretless.dev.yml](fixtures/secretless.dev.yml))

- On port 5555, it will start a service that connects to the `pg` server
- On port 6666, it will start a service that connects to the `pg_no_tls` server

This is different from the behavior of the `./start` script, which starts Secretless Broker with
a large number of services in order to run the full automated test suite. Therefore,
running `./test` will fail if you run it after running `./dev` instead of `./start`.

#### Log in to the Postgres server via the Postgres connector

To connect to the Postgres server directly, you can run:

```sh
docker-compose exec test psql -h pg -p 5432 -U test -d postgres # Password: "test"
```

To connect to the Postgres server via Secretless Broker, you can run:

```sh
docker-compose exec test psql -h secretless-dev -p 5555 -d postgres # Will connect to the `pg` container
docker-compose exec test psql -h secretless-dev -p 6666 -d postgres # Will connect to `pg_no_tls`
```

## Debugging

### Using VS Code

The easiest way to do Secretless Broker development is to use the VS Code
debugger. As above, you will want to start up your Postgres server container before
beginning development. You can choose to run `./dev` to start the entire environment
or just start the Postgres server by running `docker-compose up -d pg` from this directory.

This repository includes a VS Code launch configuration in
`/.vscode/launch.json` for the Postgres connector. Once you start the debugger in
VS Code (which will automatically start the Secretless Broker with the dev Postgres
Connector configuration), you can send requests to the Postgres server using a
Postgres client pointed towards the port you specified in the
`fixtures/secretless.debug.yml` file (in this case, `7777`).

```sh
# Connect directly to Postgres server, should require correct password ("test").
psql -h localhost -p 5432 -U test -d postgres
# Connect to Secretless Broker, should not require password
psql -h localhost -p 7777 -U test -d postgres
```

You can now set breakpoints in your code and step through the code using the
debugger.
