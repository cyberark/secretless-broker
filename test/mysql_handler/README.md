# MySQL Handler Development

## Usage / known limitations

- The MySQL handler is currently limited to connections via Unix domain socket 
- MySQL clients usually require a password to be included with a connection request; any non-empty value may be entered when connecting via the Secretless client. The Secretless MySQL handler will remove this dummy value and inject the correct password value it retrieves according to its configuration.

### To use the Secretless MySQL handler:
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

#### Start and configure secretless
From the root project directory, build the secretless binaries for your platform:
```
platform=$(go run test/print_platform.go)
./build/build.sh $platform amd64
```

From this directory, start secretless:
```
./run_dev.sh
```

#### Log in to the MySQL server via the Secretless MySQL handler
In another terminal, navigate to the `test/mysql_handler` directory and send a MySQL request via Unix socket:
```
mysql -u testuser --socket=run/mysql/mysql.sock
```

Once logged in, you should be able to `SELECT * FROM testdb.test` and see the rows that were added to the sample table.

Note: this assumes you have a MySQL client installed locally on your machine. A decent one is [mysqlsh](https://dev.mysql.com/doc/refman/5.7/en/mysqlsh.html); if you use `mysqlsh`, you will need to create an executable `mysql` file in your `PATH` that contains the following in order to be able to run `run_dev_test.sh` locally:
```
#!/bin/bash -ex

mysqlsh --sql "$@"
```

## MySQL Handler Development

The easiest way to do Secretless development is to use the VS Code debugger. As above, you will want to start up your MySQL server container before beginning development. To configure the Secretless server, you can provide VS Code with a `launch.json` file for debugging by copying the sample file below to `.vscode/launch.json`, replacing `[YOUR MYSQL PORT]` with the actual exposed port of your MySQL Docker container.

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
      "args": [ "-f", "/Users/gjennings/go/src/github.com/conjurinc/secretless/test/mysql_handler/secretless.dev.yml"],
      "showLog": true
    }
  ]
}
```

Once you start the debugger (which will automatically start Secretless with the dev MySQL Handler configuration), you can send requests to the MySQL server via a client as above.

## Running the test suite

#### Run the tests locally
Run MySQL in Docker:
```sh-session
$ docker-compose up -d mysql
```

Run Secretless locally and execute tests:
```sh-session
$ ./run_dev_test
...
ok      github.com/conjurinc/secretless/test/mysql_handler   0.048s
2018/01/11 15:06:56 Caught signal terminated: shutting down.
```


#### Run the tests in Docker
Make sure you have built updated Secretless binaries for your platform and updated Docker images before running the test suite.

To run the test suite in Docker, run:
```
./stop.sh   # Remove all existing project containers
./start.sh  # Stand up MySQL and Secretless servers
./test.sh   # Run tests in a test container
```