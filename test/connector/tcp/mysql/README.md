# MySQL Connector Development

## Building

From the project root directory, build the Secretless Broker containers.

```sh
./bin/build
```

## Testing

### Run the tests in Docker

To run the test suite in Docker, run:

```sh
./start  # Stand up MySQL and Secretless Broker servers
./test   # Run tests in a test container
./stop   # Clean up all running containers
```

## Developing

### Start the development environment

Run `./dev` from this directory. This will automatically start a MySQL server in
a Docker container at `localhost:3306`. It will also configure the MySQL server as follows:

- Create a `testuser` user (with password `testpass`)
- Authorize the `testuser` user to connect to the database server from any IP and access any schema
- Create a table `test` in the `testdb` schema and add two rows

It will also start a container running Secretless Broker and a test container
that can be used to send requests to the MySQL server via Secretless Broker.

Note: When you run `./dev`, it will start Secretless Broker wtih two services:
(See [fixtures/secretless.dev.yml](fixtures/secretless.dev.yml))

- On port 5555, it will start a service that connects to the `mysql` server via TCP
- On port 6666, it will start a service that connects to the `mysql_no_tls` server via TCP

This is different from the behavior of the `./start` script, which starts Secretless Broker with
a large number of services in order to run the full automated test suite. Therefore,
running `./test` will fail if you run it after running `./dev` instead of `./start`.

#### Log in to the MySQL server via the MySQL connector

To connect to the MySQL server directly, you can run:

```sh
docker-compose exec test mysql -h mysql -P 3306 -u testuser -ptestpass
```

To connect to the MySQL server via Secretless Broker, you can run:

```sh
# MySQL will prompt you for a password, but you can just press "Enter"
docker-compose exec test mysql -h secretless-dev -P 5555 # Will connect to the `mysql` container
docker-compose exec test mysql -h secretless-dev -P 6666 # Will connect to `mysql_no_tls`
```

## Debugging

### Using VS Code

The easiest way to do Secretless Broker development is to use the VS Code
debugger. As above, you will want to start up your MySQL server container before
beginning development. You can choose to run `./dev` to start the entire environment
or just start the MySQL server by running `docker-compose up -d mysql` from this directory.

This repository includes a VS Code launch configuration in
`/.vscode/launch.json` for the MySQL connector. Once you start the debugger in
VS Code (which will automatically start the Secretless Broker with the dev MySQL
Connector configuration), you can send requests to the MySQL server using a
MySQL client pointed towards the port you specified in the
`fixtures/secretless.debug.yml` file (in this case, `7777`).

```sh
# Connect directly to MySQL server, should require correct password ("testpass").
mysql -h 0.0.0.0 -P 3306
# Connect to Secretless Broker, press "Enter" if prompted for password.
# Should work regardless of whether you enter a password.
mysql -h 0.0.0.0 -P 7777
```

You can now set breakpoints in your code and step through the code using the
debugger.
