---
title: Secretless
id: listeners
layout: docs
description: Secretless Documentation
permalink: docs/reference/listeners
---

# Listeners

You can configure the following kinds of Secretless *Listeners*:

1) `unix` Secretless serves the backend protocol on a Unix domain socket.

2) `tcp` Secretless serves the backend protocol on a TCP socket.

For example, PostgreSQL clients can connect to the PostgreSQL server either via Unix domain socket or over a TCP connection. If you are setting up Secretless to facilitate a connection to a PostgreSQL server, you can either configure it:

<ul>
  <li>To listen on a Unix socket as usual (default: <code>/var/run/postgresql/.s.PGSQL.5432</code>)
    <pre>
      listeners:
      - name: pg_socket
        protocol: pg
        socket: /sock/.s.PGSQL.5432
    </pre>
  In this case, the client would be configured to connect to the database URL <code>/sock</code>.
  </li>

  <li>To listen on a given port, which may be the PostgreSQL default 5432 or may be a different port to avoid conflicts with the actual PostgreSQL server
    <pre>
      listeners:
      - name: pg_tcp
        protocol: pg
        address: 0.0.0.0:5432
    </pre>
  In this case, the client would be configured to connect to the database URL <code>localhost:5432</code>
  </li>
</ul>

Note that in each case, **the client is not required to specify the username and password to connect to the target service**. It just needs to know where Secretless is listening, and it connects to Secretless directly via a local, unsecured connection.

In general, there are currently two strategies to redirect your client to connect to the target service via Secretless:

<ol>
  <li><strong>Connection URL</strong>
    <br/>
    Connections to the backend service are established by a connection URL. For example, PostgreSQL supports connection URLs such as <code>postgres://user@password:hostname:port/database</code>. <code>hostname:port</code> can also be a path to a Unix socket, and it can be omitted to use the default PostgreSQL socket <code>/var/run/postgresql/.s.PGSQL.5432</code>.
  </li>
  <li><strong>Proxy</strong>
    <br/>
    HTTP services support an environment variable or configuration setting <code>http_proxy=[url]</code> which will cause outbound traffic to route through the proxy URL on its way to the destination. Secretless can operate as an HTTP forward proxy, in which case it will place the proper authorization header on the outbound request. It can also optionally forward the connection using HTTPS. The client should always use plain <code>http://</code> URLs, otherwise Secretless cannot read the network traffic because it will encrypted.
  </li>
</ol>

Regardless of the connection strategy, the operating system provides security between the client and Secretless. It's important to configure the OS properly so that unauthorized processes and clients can't connect to Secretless. With Unix domain sockets, operating system file permissions protect the socket. With TCP connections, Secretless should be listening only on localhost.

The Listener configuration governs the _client to Secretless_ connection. The connection from Secretless to the PostgreSQL server is defined in the Handler configuration, where the actual address and credential information for the connection to the PostgreSQL server is defined.

At this time, the Secretless-to-target-service connection always happens over TCP by default.
