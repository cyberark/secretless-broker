---
title: Secretless
id: docs
layout: docs
description: Secretless Documentation
permalink: docs
---

  <p class="card-heading">Why Secretless?</p>

  “Secretless” is designed to solve two problems. The first is **loss or theft of credentials from applications and services**, which can occur by:

  - Accidental credential leakage (e.g. credential checked into source control, etc)
  - An attack on a privileged user (e.g. phishing, developer machine compromises, etc)
  - A vulnerability in an application (e.g. remote code execution, environment variable dump, etc)

  The second is **downtime caused when applications or services do not respond to credential rotation** and crash or get locked out of target services as a result.

  Keeping secrets in a vault is a good practice. However, when a client “checks out” a secret from a vault, where does it go? Often, directly into unprotected application memory at which point that client app has now part of the threat surface. There are a stringent set of best practices and recommendations that should be be followed by every user and application which checks secrets out of a vault but even if you harden the app in every possible way, it does not protect you from 0-day vulnerabilities in the app, underlying framework, and/or the programming language used.

  Exposing plaintext secrets to clients (both users and machines) is hazardous from both a security and operational standpoint. First, by providing a secret to a client, the client becomes part of the threat surface. If the client is compromised, then the attacker has a good chance of obtaining the plaintext secrets and being able to establish direct connections to backend resources. To mitigate the severity of this problem, important secrets are (or should be) rotated (changed) on a regular basis. However, rotation introduces the operational problem of keeping applications up to date with changing passwords. This is a significant problem as many applications only read secrets on startup and are not prepared to handle changing passwords.

  In the long run, it’s a losing battle to try and train every user and application developer about how to safely handle a secret once they’ve obtained it even if you can mitigate all of the attack surfaces of the client app itself. There are so many things that can go wrong, and people (and applications) have better things to worry about than babysitting secrets (such doing their jobs). Do you want developers to write features or try to be security engineers too?

  Secretless alleviates users and applications from the burden of handling secrets. Instead of checking out secrets, when a user or application needs access to something important, it uses the Secretless proxy to establish the connection to the target service on its behalf.

  Secretless is specifically written to be really good at interacting with a variety of vaults and abstracting away the security to a specialized module that can easily be audited and expanded with custom plugins. It does a really good job of securing any secrets that it might obtain from a vault, and it does it the same way no matter the developer, technology, or platform. Additionally, it knows what to do when a secret has been changed so that the client users and code never have to respond to changing secrets or even know that they are changing.

  In summary, when the client connects to the target service through the Secretless broker:

  - **The client is not part of the threat surface**

    The client/app no longer has direct access to the password, and therefore cannot reveal it.

  - **The client doesn't have to know how to properly manage secrets**

    Handling secrets safely involves some complexities, and when every application needs to know how to handle secrets, accidents happen. The Secretless broker centralizes the client-side management of secrets into one code base, making it easier for developers to focus on delivering features.

  - **The client doesn't have to manage secret rotation**

    The Secretless broker is responsible for establishing connections to the backend, and can handle secrets rotation in a way that's transparent to the client.
