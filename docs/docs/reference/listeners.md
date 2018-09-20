---
title: Listeners
id: listeners
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/listeners.html
---

You can configure the following kinds of Secretless Broker *Listeners*:

1) `unix` Secretless Broker serves the backend protocol on a Unix domain socket.

2) `tcp` Secretless Broker serves the backend protocol on a TCP socket.

For example, PostgreSQL clients can connect to the PostgreSQL server either via Unix domain socket or over a TCP connection. If you are setting up Secretless Broker to facilitate a connection to a PostgreSQL server, you can either configure it:

- To listen on a Unix socket as usual (default: `/var/run/postgresql/.s.PGSQL.5432`)

  ```yml
    listeners:
    - name: pg_socket
      protocol: pg
      socket: /sock/.s.PGSQL.5432
  ```
  In this case, the client would be configured to connect to the database URL `/sock`.

- To listen on a given port, which may be the PostgreSQL default 5432 or may be a different port to avoid conflicts with the actual PostgreSQL server

  ```yml
    listeners:
    - name: pg_tcp
      protocol: pg
      address: 0.0.0.0:5432
  ```
  In this case, the client would be configured to connect to the database URL `localhost:5432`


Note that in each case, **the client is not required to specify the username and password to connect to the target service**. It just needs to know where the Secretless Broker is listening, and it connects to the Secretless Broker directly via a local, unsecured connection.

In general, there are currently two strategies to redirect your client to connect to the target service via the Secretless Broker:

- **Connection URL**
    <br/>
    Connections to the backend service are established by a connection URL. For example, PostgreSQL supports connection URLs such as `postgres://user@password:hostname:port/database`. `hostname:port` can also be a path to a Unix socket, and it can be omitted to use the default PostgreSQL socket `/var/run/postgresql/.s.PGSQL.5432`.

- **Proxy**
    <br/>
    HTTP services support an environment variable or configuration setting `http_proxy=[url]` which will cause outbound traffic to route through the proxy URL on its way to the destination. The Secretless Broker can operate as an HTTP forward proxy, in which case it will place the proper authorization header on the outbound request. It can also optionally forward the connection using HTTPS. The client should always use plain `http://` URLs, otherwise Secretless cannot read the network traffic because it will be encrypted.

## Listener Security

Regardless of the connection strategy, the operating system provides security between the client and Secretless. It's important to configure the OS properly so that unauthorized processes and clients can't connect to Secretless. With Unix domain sockets, operating system file permissions protect the socket. With TCP connections, Secretless should be listening only on localhost.

The Listener configuration governs the _client to Secretless_ connection. The connection from Secretless to the PostgreSQL server is defined in the Handler configuration, where the actual address and credential information for the connection to the PostgreSQL server is defined.

At this time, the Secretless-to-target-service connection always happens over TCP by default.
