- [Tests cases](#tests-cases)
  - [Secretless Binary](#secretless-binary)
  - [In-process Proxy Service](#in-process-proxy-service)
- [Debugging Secretless Broker as it is Running in a Container](#debugging-secretless-broker-as-it-is-running-in-a-container)
  - [Start your MSSQL server and Secretless Broker Debug Image](#start-your-mssql-server-and-secretless-broker-debug-image)
  - [Configure Your Intellij / Goland IDE](#configure-your-intellij--goland-ide)
    - [Create a `Go Remote` Run Configuration](#create-a-go-remote-run-configuration)
    - [Add Breakpoint(s)](#add-breakpoints)
  - [Run an `sqlcmd` To Start Debugging](#run-an-sqlcmd-to-start-debugging)

# MSSQL Integration tests

This folder and its sub-folders hold the integration tests for the MSSQL service
connector.

A number of integration tests rely on the existence of MSSQL clients in the
environment and pre-running services, i.e. Secretless, MSSQL server instances
etc. We have decided to standardise on **docker-compose** to run these services,
and to run the tests inside a test container image that can have any other
dependencies the tests might have. The biggest benefit for us with this setup is
ease of reproduction.

`docker-compose.yml` files contain the declarations of the aforementioned
services. The accompanying `./start`, `./test` and `./stop` scripts are used to
start the services, run the tests and clean up, respectively. At present, the test
container image contains a variety of MSSQL clients, these are needed for some
of the tests.

## Tests cases

The test cases in this folder are integration tests for the MSSQL service
connector. The goal of these tests is to exercise functionality at a relatively
high level. The tests are fully-fledged Go tests, and therefore follow the
conventions of Go testing; they are defined in Go files with the suffix
`_test.go`, and can be run by calling `go test`.

There are 2 types of tests in this folder. The test cases are generally
composed of a **database client**, **Secretless Proxy** and a **target
service**. A realistic target service is provided through docker-compose.
There's also a mock target service use to capture the packets an actual server
might receive.

Below are 2 sections, each covering a separate type of integration test.

### Secretless Binary

This type of test validates functionality through the Secretless binary.
This is a sort of end-to-end test whose components are **database client**,
**Secretless binary**, **target service**. Its configuration is relatively
static.

This type of test requires:

- Setup of Secretless services by adding them to the Secretless configuration
  (`secretless.yml`).
- Coordination between the Secretless configuration and the test case in the
  test code.
- Invoking the client from the Go test code.

Though it can seem a bit cumbersome, this type of test takes its value
from being truly E2E because it consumes Secretless in its release packaging
as a container image.

### In-process Proxy Service

This type of test validates functionality through an in-process
Proxy Service. This is very similar to approach above, except that it
takes the **Secretless binary** component and moves the interface further
inward. The Secretless binary is the combination of the many internal
parts of Secretless. Roughly speaking, these are the parts:

1. Parse Config into Service and Credential Specs
1. Each Credential Spec is used to create a closure that fetches
   credentials using a Credential Provider and returns them as a map of
   type `map[string][]byte`, string to byte-slice key-value pairs.
1. Each Service Spec is used to create a Proxy Service. A Proxy Service
   listens on a particular network socket, and handles the creation of
   authenticated connections and the piping of authenticated connections.

A Proxy Service comes bundled with a Service Connector, network socket
listener and the credential fetching closure mentioned earlier. It is the Service
Connector which has the protocol specific logic to create an authenticated
connection. Some Proxy Services are not configurable and the Proxy Service
is tightly integrated with the Service Connector e.g. SSH. Other Proxy
Services such as the TCP Proxy Service can be configured to use a particular
Service Connector e.g. MSSQL, Postgres, MySQL etc. For MSSQL, the Proxy
Service is made up of a MSSQL Service Connector on top of a TCP Connector.

Parts (1) and (2) above generally do not require testing on a per Service
Connector basis. For this type of test we do away with these parts and
instead of relying on Secretless we create the Proxy Service ourselves
in-process. You likely won't ever have to worry about how this is done
since we've already written a wrapper in `./mssql_proxy_service.go` that:

1. Creates the network socket listener.
1. A convenient method for wrapping a credential map in a closure
   (as needed by the Proxy Service) that clones and returns it every time
   the closure is called.
1. Allows for easy creation (does (1) and (2)), starting and stopping of
   the Proxy Service.

This results in an extremely convenient and fast mechanism for testing
iterations of credential values against a Proxy Service. An example where
this has been useful is in testing the different scenarios that arise
in supporting TLS.

Below is an example test using this convenient wrapper.

```go
// Specify client request
clientRequest := clientRequest{
  database: "tempdb",
  readOnly: false,
  query: encryptionOptionQuery,
}

// Read self-signed server certificate
cert, _ := ioutil.ReadFile("./certs/server-cert.pem")

// Proxy Request through Secretless
out, _, err := clientRequest.proxyRequest(
  sqlcmdExec,
  map[string][]byte{
    "sslmode":     []byte("verify-full"),
    "sslrootcert": cert,
    "sslhost":     []byte("test"),
    "username":    []byte("sa"),
    "password":    []byte("yourStrong()Password"),
    "host":        []byte("localhost"),
    "port":        []byte("1433"),
  },
)
```

In the example the `proxyRequest` method on the `clientRequest` instance is
used. This method creates an MSSQL Proxy Service in-process, then runs a
client request through it, then cleans up. It returns the output from the
client request, and outputs the logs from the lifecycle of the Proxy Service.
This is wonderful for debugging.

It's worth noting that a similar test case would be more cumbersome if done
through the Secretless binary which would require creation of a service in
Secretless config; you'd have to pick arbitrary service names, service ports,
credential provider, and coordinate both the running of Secretless and the
credentials with your test cases.

## Debugging Secretless Broker as it is Running in a Container

Using a specially built "remote-debug" image for the Secretless Broker, it
is possible to connect a Delve-capable debugger such as Intellij or Goland
to a Secretless Broker process that is running inside a Docker container.
Once connected, you may debug Secretless Broker functionality, e.g. using
breakpoints, single-stepping, and examination of Golang data structures, etc.

The steps for starting the Secretless Broker and attaching a debugger are
described in the sections that follow.

### Start your MSSQL server and Secretless Broker Debug Image

From this directory, call
```
./start -D
```
or alternatively:
```
./remote_debug
```
This will automatically start a MSSQL server in a Docker container serving
at `localhost:1433`, and a remote-debug mode Secretless Broker serving
at `localhost:2223`.

The debug-mode Secretless Broker will be running a version of the secretless
broker binary that is compiled with optimization turned off (to enable
the best debugging experience). This secretless broker binary will be run
via Delve, which provides a debug link with Delve-capable debug IDEs
(e.g. Intellij and Goland).

### Configure Your Intellij / Goland IDE

Reference: [Debugging Containerized Go Applications](https://blog.jetbrains.com/go/2018/04/30/debugging-containerized-go-applications/)

#### Create a `Go Remote` Run Configuration

- In your IDE, select \<Run\> \<Edit Configurations...\> \<`+`\> \<Go Remote\>
- In the `Name:` box, enter `secretless-broker`
- In the `Port` box, enter `40000`
- Select \<OK\>

#### Add Breakpoint(s)

- In your IDE, navigate to a place in the source code where you would like
- Left-mouse-click in the column between the line number and the line of
  code, and you should see a red dot, indicating that a breakpoint has
  been added.

### Run an `sqlcmd` To Start Debugging

When running the `sqlcmd` manually for testing with the remote debugging,
use the `-t` and `-l` flags to disable timeouts on MSSQL transactions
and MSSQL handshakes, respectively:
```
sqlcmd -S "127.0.0.1,2223" -U "x" -P "x" -Q "SELECT 1+1" -t 0 -l 0
```

If all goes right, your IDE should hit your chosen breakpoint, indicated by
the line of code having a blue background.
