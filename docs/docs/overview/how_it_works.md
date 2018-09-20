---
title: How Does It Work?
id: how_it_works
layout: docs
description: Secretless Broker Documentation
permalink: docs/overview/how_it_works.html
---

In a Secretless Broker deployment, when a client needs access to a Target Service it doesn't try to make a direct connection to it. Instead, it sends the request through its Secretless Broker.

<div class="docs-img">
  <img src="/img/secretless_architecture.svg" alt="Secretless Broker Architecture">
</div>

The Secretless Broker authenticates with a secrets vault and obtains an identity credential. This identity credential is managed securely within the Secretless Broker, and used to obtain a secret to access the Target Service such as database password from the secrets vault. The connection secrets are managed entirely within the Secretless Broker process, and never exposed to the client. The Secretless Broker uses the connection secret to establish a connection to the Target Service and then transmits data between the client and the target.

## Standard Workflow

1. A Client (user or code) connects to the Secretless Broker (a local service) to obtain a connection to the Target Service (a database, server or web service).
1. If needed, the Secretless Broker authenticates with an external vault to obtain an identity credential, which is managed securely within the Secretless Broker process.
1. The Secretless Broker uses the identity credential to obtain secrets which allow access to the Target Service. Secrets are managed securely by the Secretless Broker process.
1. Secretless Broker uses the secrets to open a connection to the Target Service.
1. Secretless Broker pipes traffic between the Client and the Target Service.
1. If a secret is changed, the Broker automatically obtains the new secret and uses it when establishing new connections.


The Secretless Broker is a proxy that intercepts traffic to the Target Service and performs the authentication phase of the backend protocol. The data-transfer phases of the protocol are direct pass-through between the client and Target Service.

Examples of protocols that can be brokered:  

-  Database protocols such as Oracle, Postgresql, MySQL, NoSQL flavors, etc.
-  HTTP via Authorization header
-  SSH, via MITM or by implementing an ssh-agent   

Any published protocol can be supported in the Secretless Broker. Software code in the Secretless Broker is generally required for each new protocol. For a list of currently supported Target Services, please see our documentation on <a href="/docs/reference/handlers/overview.html">Handlers</a>.

The Secretless Broker typically runs locally alongside the client application. Authentication between the Client and the Secretless Broker is managed by the operating system, e.g. local connection via Unix socket or HTTP connection to 127.0.0.1.  In container-managed environments such as Kubernetes, the Secretless Broker can be a “sidecar” container which is securely networked to the application container.


## Internal Architecture

<img src="/img/secretless_internal_architecture.svg" alt="Secretless Broker Internal Architecture">

Internally, when the Secretless Broker receives a new connection request:
1. The Proxy determines the Listener to send the request to, based on the port / socket the request was sent to (there is a Listener for each potential connection to a Target Service)
1. The Listener receives the connection request, and forwards it to its Handler
1. The Handler retrieves credentials using a Credential Provider
1. The Handler injects the credentials into a new connection request and opens a new connection to the Target Service
1. The Handler streams the connection

## Secretless Broker Configuration

The Secretless Broker relies on its configuration to determine which Target Services
it can connect to and how it should retrieve the access credentials to authenticate
with those services.

Each Secretless Broker configuration includes two sections:

* `listeners`: A list of protocol Listeners, each one on a Unix socket or TCP port.
* `handlers`: A list of Handlers to process the requests received by each Listener. Handlers implement the protocol for the Target Services and are configured to obtain the backend connection credentials from one or more Credential Providers.

The [Configuration Managers](/docs/reference/config-managers/overview.html) section
in the Secretless Broker reference has more information about how to provide the Broker with
its configuration in practice.
