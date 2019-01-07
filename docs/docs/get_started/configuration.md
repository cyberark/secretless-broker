---
title: Configuration
id: configuration
layout: docs
description: Secretless Broker Documentation
redirect_to: https://docs.secretless.io/Latest/en/Content/Resources/_TopNav/cc_Home.htm
---

The Secretless Broker relies on its configuration to determine which Target Services
it can connect to and how it should retrieve the access credentials to authenticate
with those services.

Each Secretless Broker configuration includes two sections:

* `listeners`: A list of protocol Listeners, each one on a Unix socket or TCP port.
* `handlers`: A list of Handlers to process the requests received by each Listener. Handlers implement the protocol for the Target Services and are configured to obtain the backend connection credentials from one or more Credential Providers.

## Examples

In the examples below, we share the Secretless configurations that were used in
each of the [quick start demos](/docs/get_started/quick_start.html). For ease of
understanding we've broken them up into three separate configurations. In practice
you can configure Secretless Broker to handle as many types of connections as you
need; to see how we configured Secretless Broker to handle all three of these
connection types at once, check out the [actual configuration](https://github.com/cyberark/secretless-broker/blob/master/demos/quick-start/docker/etc/secretless.yml)
we used in building the quick start Docker image.

<div id="configuration-examples">
  <ul>
    <li><a href="#tabs-config-pg">PostgreSQL</a></li>
    <li><a href="#tabs-config-ssh">SSH</a></li>
    <li><a href="#tabs-config-http">HTTP</a></li>
  </ul>
  <div id="tabs-config-pg">
    <pre>
listeners:
  - name: pg_tcp
    protocol: pg
    address: 0.0.0.0:5454

handlers:
  - name: pg
    listener: pg_tcp
    credentials:
      - name: address
        provider: literal
        id: localhost:5432
      - name: username
        provider: env
        id: QUICKSTART_USERNAME
      - name: password
        provider: env
        id: QUICKSTART_PASSWORD
    </pre>
  </div>
  <div id="tabs-config-ssh">
    <pre>
listeners:
  - name: ssh
    protocol: ssh
    address: 0.0.0.0:2222

handlers:
  - name: ssh
    listener: ssh
    credentials:
      - name: address
        provider: literal
        id: localhost
      - name: user
        provider: literal
        id: user
      - name: privateKey
        provider: env
        id: SSH_PRIVATE_KEY
    </pre>
  </div>
  <div id="tabs-config-http">
    <pre>
listeners:
  - name: http_basic_auth
    protocol: http
    address: 0.0.0.0:8081

handlers:
  - name: http_basic_auth
    type: basic_auth
    listener: http_basic_auth
    match:
     - ^http\:\/\/quickstart\/
     - ^http\:\/\/localhost.*
    credentials:
      - name: username
        provider: env
        id: BASIC_AUTH_USERNAME
      - name: password
        provider: env
        id: BASIC_AUTH_PASSWORD
    </pre>
  </div>
</div>

## Configuring Secretless Broker

The [Configuration Managers](/docs/reference/config-managers/overview.html) section
in the Secretless Broker reference has more information about how to provide the Broker with
its configuration in practice.

<script>
  $( function() {
    $( "#configuration-examples" ).tabs();
  } );
</script>
