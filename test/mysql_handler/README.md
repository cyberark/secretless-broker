# MySQL Handler Development

## Usage / known limitations

- The MySQL handler is currently limited to connections via Unix domain socket 

### To use the MySQL handler:
#### Start your MySQL server
From this directory, call
```
docker-compose up -d mysql
```
This will automatically start a MySQL server in a Docker container at `localhost:$(docker-compose port mysql 3306)`.

It will also configure the MySQL server as follows:
- Create a `testuser` user (with password `testpass`)
- Authorize the `testuser` user to connect to the database server from any IP and access any schema
- Create a table `test` in the `testdb` schema and add two rows

#### Start and configure secretless-broker
From the root project directory, build the Secretless Broker binaries for your platform:
```
platform=$(go run test/print_platform.go)
./bin/build $platform amd64
```

From this directory, start Secretless Broker:
```
./run_dev
```

#### Log in to the MySQL server via the MySQL handler
In another terminal, navigate to the `test/mysql_handler` directory and send a MySQL request via Unix socket:

_Note: Since the Secretless Broker container runs the daemon as a limited user, sockets should be mounted to `/sock` directory._

```
mysql --socket=sock/mysql.sock
```
or via TCP:
```
mysql -h 0.0.0.0 -P 13306 --ssl-mode=DISABLED
```
You may be prompted for a password, but you don't need to enter one; just hit return to continue.

Once logged in, you should be able to `SELECT * FROM testdb.test` and see the rows that were added to the sample table.

Note: this assumes you have a MySQL client installed locally on your machine. In the examples above and when you run the test suite locally, it is assumed you use one like [mysqlsh](https://dev.mysql.com/doc/refman/5.7/en/mysqlsh.html), which assumes SSL connections when possible by default (and has an `--ssl-mode` flag you can use to disable SSL).

If you use `mysqlsh`, you will need to create an executable `mysql` file in your `PATH` that contains the following in order to be able to run `run_dev_test` locally:
```
#!/bin/bash -ex

mysqlsh --sql "$@"
```
This will run the MySQL shell as a client in SQL mode.

## MySQL Handler Development

### Using VS Code

The easiest way to do Secretless Broker development is to use the VS Code debugger. As above, you will want to start up your MySQL server container before beginning development. To configure the Secretless Broker, you can provide VS Code with a `launch.json` file for debugging by copying the sample file below to `.vscode/launch.json`, replacing `[YOUR MYSQL PORT]` with the actual exposed port of your MySQL Docker container.

Sample `launch.json`:
```
{
  // Use IntelliSense to learn about possible attributes.
  // Hover to view descriptions of existing attributes.
  // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
  "version": "0.2.0",
  "configurations": [
    {
      "name": "MySQL Handler",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "remotePath": "",
      "port": 2345,
      "host": "127.0.0.1",
      "program": "${workspaceFolder}/cmd/secretless/",
      "env": { "MYSQL_HOST": "localhost", "MYSQL_PORT": "[YOUR MYSQL PORT]", "MYSQL_PASSWORD": "testpass" },
      "args": [ "-f", "/Users/gjennings/go/src/github.com/cyberark/secretless-broker/test/mysql_handler/secretless.dev.yml"],
      "showLog": true
    }
  ]
}
```

Once you start the debugger (which will automatically start the Secretless Broker with the dev MySQL Handler configuration), you can send requests to the MySQL server via a client as described above.

### Using Docker

You can also run:
```
cd test/mysql_handler/
./start
docker-compose run --rm secretless-dev
```

Then, to connect with MySQL you can run either
`mysql -h secretless -P 3306`
to connect via TCP (SSL mode is disabled by default), or
`mysql --socket=/sock/mysql.sock`
to connect via Unix socket.

## Running the test suite

#### Run the tests in Docker
Make sure you have built updated Secretless Broker binaries for Linux and updated Docker images before running the test suite.

To run the test suite in Docker, run:
```
./stop   # Remove all existing project containers
./start  # Stand up MySQL and Secretless Broker servers
./test   # Run tests in a test container
```
Make sure you build the project by running `./bin/build` in the project root
before running the tests so that the test container will be using updated
code. If you want to run using your local changes, you can run `./test -l`
instead, which will mount your local project directory as a volume in the
test container, overwriting the project directory built into the container
image.
