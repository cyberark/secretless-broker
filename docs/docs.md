---
title: Secretless
id: docs
layout: landing
description: Secretless Documentation
---
<div id="docs-overview">
  <p class="card-heading">Overview</p>
    <p>“Secretless” is a design to relieve client users and applications of the responsibility of interacting with a vault and having to manage secrets. This reduces the threat surface of secrets and also handles secrets rotation in a way that’s transparent to the client. The “Secretless” design accomplishes this without changing how clients connect to services, and allowing them to use standard libraries and tools.</p>
    <p class="card-documentation-heading">Why Secretless?</p>
    <p>Exposing plaintext secrets to clients (both users and machines) is hazardous from both a security and operational standpoint. First, by providing a secret to a client, the client becomes part of the threat surface. If the client is compromised, then the attacker has a good chance of obtaining the plaintext secrets and being able to establish direct connections to backend resources. To mitigate the severity of this problem, important secrets are (or should be) rotated (changed) on a regular basis. However, rotation introduces the operational problem of keeping applications up to date with changing passwords. This is a significant problem as many applications only read secrets on startup and are not prepared to handle changing passwords.</p>
    <p>When the client connects to the backend resource through Secretless:</p>
    <ul>
      <li>The client is not part of the threat surface. The client does not have direct access to the password, and therefore cannot reveal it</li>
      <li>The client does not have to know how to properly manage secrets. Handling secrets safely is very difficult. When every application needs to know how to handle secrets, accidents happen. Secretless centralizes the client-side management of secrets into one code base</li>
      <li>The client does not have to handle changing secrets. Secretless is responsible for establishing connections to the backend, and can handle secrets rotation in a way that’s transparent to the client</li>
    </ul>
</div>   