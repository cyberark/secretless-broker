This folder includes smoke tests for the external plugin interface. It
implements an example plugin and verifies that it works as expected.

Currently tests:
 - Manager connection reject/allow
 - Sample Provider
 - Listener that passes a connection and injects a custom header in a HTTP-like
 connection

```
                                example plugin (manager)
                                           |
curl (or another client)  <---> example plugin (listener) <---> echo server
```

## Pieces:

### Echo server (`./echo`)

This tiny program listens on port 6174 and waits for connections. When something
connects to it, it waits for a double '\r\n' (Pseudo-http-ish) and then echoes back
the content to the sender and closes the connection.

Note: This server does not return proper HTTP responses.

You can test the server with:
```
$ curl -A Agent http://<agent_host>:6174
```

### Secretless plugin

The plugin opens two local ports (6175 and 6176) and forwards them to the backend echo
service. Manager closes any connections to even ports (in this case 6176). Allowed
connections are intercepted and two mock headers added. Tests listen for these headers to ensure
that the plugin is operating as expected.

Plugin is generally built and placed in `/usr/local/lib/secretless` as a `.so` library.


#### Listener (`./example/listener.go`)

This listener waits on two ports (6175 and 6176) and forwards them to the backend. Since the
expected traffic is HTTP-like, the message is parsed, and a mock header `Example-Header: IsSet`
is injected to the traffic to the echo server. The traffic is sent back from echo server
unmodified.

Note: If either side of the connection is forcibly closed, the whole tunnel goes down. If there
is also too much delay in sending the messages to the backend echo server, the listener
will close the connection.

#### Provider (`./example/provider.go`)

This provider resolves variables by appending `Provider` to the variable id

#### Manager (`./example/manager.go`)

This manager doesn't really do much but if a connection is attempted on an even port, it will
close it and allow it otherwise. This plugin is registered as part of the example shared
library.
