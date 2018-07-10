---
title: Secretless
id: key-terms
layout: landing
description: Secretless Documentation
peramlink: key-terms
---
<div id="docs-key-terms">
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
        <li>Each application you run which is accessible to the outside world is also part of the threat surface. For example, a business website may be exploitable by an attacker to install malware on a server. Each other person that you give sensitive information to becomes part of the threat surface. The entry point to the Target hack was through an <a href="https://krebsonsecurity.com/2014/02/target-hackers-broke-in-via-hvac-company">HVAC contractor.</a></li>
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