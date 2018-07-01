---
title: Secretless
id: works
layout: landing
description: Secretless Documentation
---
<div id="docs-works">
  <div class="row">
    <p class="card-heading">How it Works?</p>
    <div class="col-md-8">
      <p>In a Secretless deployment, clients don’t directly obtain secrets or connect directly to protected services. Instead, when a client wants to communicate with an external service, it directs the request to a Secretless proxy service. The Secretless proxy is able to authenticate with a secrets vault on behalf of the client and obtain an identity credential. This identity credential is stored securely within the Secretless proxy, and used to obtain a backend connection secret such as database password from the secrets vault. The connection secrets are managed entirely within the Secretless service, and never exposed to the client. The Secretless proxy uses the connection secret to establishes a connection to the protected service and then transfers messages between the client and the service.</p>
    </div>
    <div class="col-md-4">
      <img class="introduction-img" src="/img/secretlessbrokerwhite.png">
    </div>
  </div>
  <ol>
    <li>A Client (user or code) connects to the Secretless Connection Broker (a local service) to obtain a connection to the Protected Resource (a database, server or web service)</li>
    <li>Secretless Connection Broker authenticates with the Secrets Vault to obtain an identity credential, which is held within an operating system keyring.</li>
    <li>Identity credential is used to check out secrets which allow access to the Protected Resource. Secrets are held by the Secretless Connection Broker within a secure operating system keyring</li>
    <li>Secretless Connection Broker uses the secrets to connect to the Protected Resource</li>
    <li>Secretless Connection Broker pipes traffic between the Client and the Protected Resource</li>
    <li>If a secret is changed, Secretless Connection Broker automatically checks out the new secret and uses it to establish new connections</li>
  </ol>
  <p>The Secretless Connection Broker is a proxy which intercepts traffic to the backend service and performs the authentication phase of the backend protocol. The data-transfer phases of the protocol are direct pass-through between the client and backend service</p>
  <p class="card-documentation-heading">Examples of protocols that can be brokered:</p>
  <ul>
    <li>HTTP via Authorization header</li>
    <li>SSH, via MITM or by implementing an ssh-agent</li>
    <li>Database protocols such as Oracle, Postgresql, MySQL, NoSQL flavors, etc</li>
  </ul>
  <p>Any published protocol can be supported in Secretless. Software code in Secretless is generally required for each new protocol.</p>
  <p>The Connection Broker typically runs locally alongside the client application. Authentication between the client and the proxy is managed by the operating system, e.g. local connection via Unix socket or HTTP connection to 127.0.0.1.  In container-managed environments such as Kubernetes, the Connection Broker can be a “sidecar” container which is linked to the application code.</p>
</div>

