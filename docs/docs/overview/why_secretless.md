---
title: Why Secretless?
id: docs
layout: docs
description: Secretless Broker Documentation
permalink: docs/overview/why_secretless.html
---

Secretless Broker is primarily designed to solve two problems. The first is
**loss or theft of credentials from applications and services**, which can occur by:

- Accidental credential leakage (e.g. credential checked into source control, etc)
- An attack on a privileged user (e.g. phishing, developer machine compromises, etc)
- A vulnerability in an application (e.g. remote code execution, environment variable dump, etc)

The second problem is that **applications must know how to interact with secrets
vaults**. This often requires code to be changed to fetch needed credentials,
and time spent implementing and maintaining connections to vaults could be better
spent delivering business value.

## Prevent Credential Theft

Keeping secrets in a vault is a good practice. However, when a client “checks out” a secret from a vault, where does it go? Often, directly into unprotected application memory at which point that client app has now part of the threat surface. There are a stringent set of best practices and recommendations that should be be followed by every user and application which checks secrets out of a vault, but even if you harden the app in every possible way it does not protect you from 0-day vulnerabilities in the app, underlying framework, and/or the programming language used.

It can be a losing battle to try and train every user and application developer about how to safely handle a secret once they’ve obtained it even if you can mitigate all of the attack surfaces of the client app itself. There are so many things that can go wrong, and people (and applications) have better things to worry about than babysitting secrets (such as doing their jobs). Do you want developers to write features or try to be security engineers too?

## Focus on Adding Business Value

Secretless Broker is specifically written to be really good at interacting with a
variety of vaults and abstracting away the security to a specialized module that
can easily be audited and expanded with custom plugins. It does a really good job
of securing any secrets that it might obtain from a vault, and it does it the same
way no matter the developer, technology, or platform.

## Summary

In summary, when the client connects to the target service through Secretless Broker:

- **The client does not have to know how to fetch credentials**

  The client/app no longer has to directly interact with a secrets vault and no
  code must be changed to fetch credentials before opening a connection. If the
  vault used changes when the environment or platform changes no changes in code
  are required, thus reducing the probability of human error and bugs.

- **The client is not part of the threat surface**

  The client/app no longer has direct access to the password, and therefore cannot reveal it.

- **The client doesn't have to know how to properly manage secrets**

  Handling secrets safely involves some complexities, and when every application needs to know how to handle secrets, accidents happen. Secretless Broker centralizes the client-side management of secrets into one code base, making it easier for developers to focus on delivering features.
