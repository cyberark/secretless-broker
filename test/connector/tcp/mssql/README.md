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
* In your IDE, select <Run> <Edit Configurations...> <`+`> <Go Remote>
* In the `Name:` box, enter `secretless-broker`
* In the `Port` box, enter `40000`
* Select <OK>

#### Add Breakpoint(s)
* In your IDE, navigate to a place in the source code where you would like
* Left-mouse-click in the column between the line number and the line of
  code, and you should see a red dot, indicating that a breakpoint has
  been added.

### Run an `sqlcmd` To Start Debugging
When running the `sqlcmd` manually for testing with the remote debugging,
use the `-t` and '-l` flags to disable timeouts on MSSQL transactions
and MSSQL handshakes, respectively:
```
sqlcmd -S "127.0.0.1,2223" -U "x" -P "x" -Q "SELECT 1+1" -t 0 -l 0
```

If all goes right, your IDE should hit your chosen breakpoint, indicated by
the line of code having a blue background.
