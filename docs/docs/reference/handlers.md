---
title: Secretless
id: handlers
layout: docs
description: Secretless Documentation
permalink: docs/reference/handlers
---

# Handlers

When Secretless receives a new request on a defined Listener, it automatically passes the request on to the Handler defined in the Secretless configuration for processing. Each Listener in the Secretless configuration should therefore have a corresponding Handler.

The Handler configuration specifies the Listener that the Handler is handling connections for and any credentials that will be needed for that connection. Several credential sources are currently supported; see the [Credential Providers](/docs/reference/providers) section for more information.

The example below defines a Handler to process connection requests from the `pg_socket` Listener, and it has three credentials: `address`, `username`, and `password`. The `address` and `username` are literally specified in this case, and the `password` is taken from the environment of the running Secretless process.
<pre>
handlers:
  - name: pg_via_socket
    listener: pg_socket
    credentials:
      - name: address
        provider: literal
        id: pg:5432
      - name: username
        provider: literal
        id: myuser
      - name: password
        provider: env
        id: PG_PASSWORD
</pre>

In production you would want your credential information to be pulled from a vault, and Secretless currently supports multiple vault Credential Providers.

When a Handler receives a new connection requests, it retrieves any required credentials using the specified Provider(s), injects the correct authentication credentials into the connection request, and opens up a connection to the target service. From there, the Handler simply transparently shuttles data between the client and service.

Select the tab for the Handler you are interested in below to learn about the credentials it accepts in its configuration file.

## Handler Credential Configuration

<div id="handler-tabs">
  <ul>
    <li><a href="#tabs-mysql">MySQL</a></li>
    <li><a href="#tabs-pg">PostgreSQL</a></li>
    <li><a href="#tabs-ssh">SSH</a></li>
    <li><a href="#tabs-ssh-agent">SSH Agent</a></li>
    <li><a href="#tabs-http">HTTP</a></li>
  </ul>

  <div id="tabs-mysql">
    <p>The required credentials for the MySQL Handler are:</p>
    <ul>
      <li><code>host</code>  - Host name of MySQL server</li>
      <li><code>port</code> - Port of MySQL server</li>
      <li><code>username</code> - Username of MySQL account</li>
      <li><code>password</code> - Password of MySQL account</li>
    </ul>
  </div>

  <div id="tabs-pg">
    <p>The required credentials for the PostgreSQL Handler are:</p>
    <ul>
      <li><code>address</code> - Connection string of the form <code>host[:port][/dbname]</code></li>
      <li><code>username</code> - Username of the PostgreSQL account</li>
      <li><code>password</code> - Password of the PostgreSQL account</li>
    </ul>
  </div>

  <div id="tabs-ssh">
    <p>The required credentials for the SSH Handler are:</p>
    <ul>
      <li><code>address</code> - Server address of the form <code>host[:port]</code> (default: port 22)</li>
      <li><code>privateKey</code> - PEM encoded private key</li>
      <li><code>user</code> - <em>optional</em>; defaults to <code>root</code></li>
      <li><code>hostKey</code> - <em>optional</em>; accepts any host key if not included</li>
    </ul>
  </div>

  <div id="tabs-ssh-agent">
    <p>The required credentials for the SSH-Agent Handler are:</p>
    <ul>
      <li><code>rsa</code> or <code>ecdsa</code> - RSA or ECDSA private key</li>
      <li><code>comment</code>  - <em>optional</em>; free-form string</li>
      <li><code>lifetime</code> - <em>optional</em>; if not 0, the number of seconds the agent will store the key for</li>
      <li><code>confirm</code> - <em>optional</em>; confirms with user before using if true</li>
    </ul>
  </div>

  <div id="tabs-http">
    <ul>
      <li>Basic Auth Credentials</li>
        <ul>
          <li><code>username</code> - Username for service</li>
          <li><code>password</code> - Password for service</li>
          <li><code>forceSSL</code> - <em>optional</em>; boolean, forces connection over https if true</li>
        </ul>

      <li>Conjur Credentials</li>
        <ul>
          <li><code>accessToken</code> - Conjur access token</li>
          <li><code>forceSSL</code> - <em>optional</em>; boolean, forces connection over https if true</li>
        </ul>

      <li>AWS Credentials</li>
        <ul>
          <li><code>accessKeyID</code> - AWS access key ID</li>
          <li><code>secretAccessKey</code> - AWS secret access key</li>
          <li><code>accessToken</code> - AWS session token</li>
        </ul>
    </ul>
  </div>
</div>

<script>
  $( function() {
    $( "#handler-tabs" ).tabs();
  } );
</script>
