---
title: Secretless
id: docs
layout: page
description: Secretless Documentation
---

<div class="container">
  <div class="card docs" id="docs">
    <p class="card-heading">Documentation</p>
    <div class="card-documentation-info">
      <p class="card-documentation-heading">Overview</p>
      <p>“Secretless” is a design to relieve client users and applications of the responsibility of interacting with a vault and having to manage secrets. This reduces the threat surface of secrets and also handles secrets rotation in a way that’s transparent to the client. The “Secretless” design accomplishes this without changing how clients connect to services, and allowing them to use standard libraries and tools.</p>
      <p class="card-documentation-heading">Why Secretless?</p>
      <p>Exposing plaintext secrets to clients (both users and machines) is hazardous from both a security and operational standpoint. First, by providing a secret to a client, the client becomes part of the threat surface. If the client is compromised, then the attacker has a good chance of obtaining the plaintext secrets and being able to establish direct connections to backend resources. To mitigate the severity of this problem, important secrets are (or should be) rotated (changed) on a regular basis. However, rotation introduces the operational problem of keeping applications up to date with changing passwords. This is a significant problem as many applications only read secrets on startup and are not prepared to handle changing passwords.</p>
      <p>When the client connects to the backend resource through Secretless:</p>
      <ul>
        <li>The client is not part of the threat surface. The client does not have direct access to the password, and therefore cannot reveal it</li>
        <li>The client does not have to know how to properly manage secrets. Handling secrets safely is very difficult. When every application needs to know how to handle secrets, accidents happen. Secretless centralizes the client-side management of secrets into one code base</li>
        <li>The client does not have to handle changing secrets. Secretless is responsible for establishing connections to the backend, and can handle secrets rotation in a way that’s transparent to the client</li>
      </ul>
      <!-- <p class="card-documentation-heading header">Get Started</p>

      <p> -->
      <p class="card-documentation-heading">How it Works?</p>
      <div class="col-md-8">
      <p>In a Secretless deployment, clients don’t directly obtain secrets or connect directly to protected services. Instead, when a client wants to communicate with an external service, it directs the request to a Secretless proxy service. The Secretless proxy is able to authenticate with a secrets vault on behalf of the client and obtain an identity credential. This identity credential is stored securely within the Secretless proxy, and used to obtain a backend connection secret such as database password from the secrets vault. The connection secrets are managed entirely within the Secretless service, and never exposed to the client. The Secretless proxy uses the connection secret to establishes a connection to the protected service and then transfers messages between the client and the service.</p>
      </div>
      <div class="col-md-4">
        <img class="introduction-img" src="/img/secretlessbrokerwhite.png">
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
      <p class="card-documentation-heading">Building Secretless</p>
    </div>
  </div>
</div>
<div class="container">
  <div class="card keyterms">
    <p class="card-heading">Key Terms</p>
    <div class="docs-keyterms-info">
    <p>Secretless:</p>
      <ul>
        <li>A random piece of data which is presented in order to gain access to a protected resource. Presentation of a secret (sometimes two) is usually sufficient to access protected data (for legitimate or illegitimate purposes.</li>
      </ul>
    <p>Connection broker:</p>
      <ul>
        <li>Software which negotiates connections to backend resources, such as web APIs and databases, without exposing the implementation details to the client</li>
      </ul>
    <p>Attack surface:</p>
      <ul>
        <li>The attack surface of a software environment refers to all the different ways that an attacker (i.e. malicious actor) might attempt to access protected resources. For example, if you run email on a server, then email is part of the attack surface - an attacker can try to send malicious emails to infect your machine.</li>
        <li>Each application you run which is accessible to the outside world is also part of the threat surface. For example, a business website may be exploitable by an attacker to install malware on a server. Each other person that you give sensitive information to becomes part of the threat surface. The entry point to the Target hack was through an <a href="https://krebsonsecurity.com/2014/02/target-hackers-broke-in-via-hvac-company">HVAC contractor</a>.</li>
      </ul>
    <p>Vault:</p>
      <ul>
        <li>A vault is a special type of database that is designed to store secret data. All the secret data in the vault is encrypted. The vault provides an API (programmatic interface) that clients (users and applications) can use to request access to secret data. The API requires clients to be both authenticated and authorized. </li>
        <li>When a client uses the API to interact with the vault, this interaction is written to a separate database called the audit. The audit can be used to search and report on historical interactions with secret data. Vaults may also be capable of automatically rotating the secrets. Or rotation may be performed by a separate application, acting through the Vault API.</li>
      </ul>
    <p>Listener:</p>
      <ul>
        <li>A listener is a named network location (e.g., port, unix domain socket, etc.) that can be connected to by clients. Secretless exposes one or more listeners that hosts connect to</li>
      </ul>
    <p>Provider:</p>
      <ul>
        <li>Secretless is vault-agnostic, and is not tied to any particular secrets source. Instead, sources are implemented as providers that are invoked by Secretless to fetch secret values</li>
      </ul>
    <p>Handler:</p>
      <ul>
        <li>When a new connection is received by a Listener, it's routed to a Handler for processing. The Handler is configured to obtain the backend connection credentials from one or more Providers</li>
      </ul>
  </div>
<div class="container">
  <div class="card getstarted" id="getstarted">
    <p class="card-heading">Get Started</p>
    <div class="docs-getstarted-info">
      <p>Quick start to run a simple example</p>
        <ul>
          <li>PostgreSQL</li>
          <li>Amazon RDS</li>
        </ul>
      <p>Simple Configuration</p>
      <p>Using the Secretless Docker image</p>

    </div>
  </div>
</div>
<div class="container">
  <div class="card installation">
    <p class="card-heading">Installation</p>
    <div class="docs-installation-info">
      <ul>
        <li>Building Docker Images</li>
        <li>Reference configurations</li>
      </ul>
    </div>
  </div>
</div>
<div class="container">
  <div class="card configuration">
    <p class="card-heading">Configuration</p>
    <div class="docs-configuration-info">
    <p>The Secretless broker configuration is written in YAML. The configuration for the broker is provided by specifying listeners and handlers, which defines which external services the broker will be connecting to, the credentials needed for each, and how clients will connect to the Secretless broker.</p>
    <p>[TO BE ADDED – details about writing a configuration YAML file]</p>
    </div>
  </div>
</div>
<div class="container">
  <div class="card backends">
    <p class="card-heading">Available Backends</p>
    <div class="docs-installation-info">
      <p>Handlers</p>
      <ul>
        <li>PostgreSQL</li>
        <li>MySQL</li>
        <li>SSH</li>
        <li>HTTP / AWS API</li>
        <li>HTTP / CONJUR API</li>
      </ul>
      <p>Providers</p>
      <ul>
        <li>CyberArk Conjur</li>
        <li>HashiCorp Vault</li>
        <li>OSX Keychain</li>
        <li>Environment</li>
        <li>File</li>
        <li>Lateral</li>
      </ul>
    </div>
  </div>
</div>
<div class="container">
  <div class="card plugins" id="plugins">
    <p class="card-heading">Plugins</p>
    <div class="docs-plugins-info">
      <p>TODO</p>
    </div>
  </div>
</div>
<div class="container">
  <div class="card examples" id="examples">
    <p class="card-heading">Examples</p>
    <div class="docs-examples-info">
      <p>Quick start to run a simple example</p>
        <ul>
          <li>PostgreSQL (TODO)</li>
          <li>AWS API (TODO)</li>
          <li>Full Demo (TODO)</li>
        </ul>
    </div>
  </div>
</div>
