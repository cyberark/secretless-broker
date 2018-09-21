---
layout: post
title: "Announcing Secretless"
date: 2018-09-20 09:00:00 -0400
author: Dustin Byrne
categories: blog
published: true
excerpt: "CyberArk is incredibly pleased to announce Secretless, a new open source project that changes the way applications consume privileged credentials."
---

CyberArk is incredibly pleased to announce Secretless, a new open source project that changes the way applications consume privileged credentials. 

Applications and secrets don’t mix well, no matter how they’re implemented. We sometimes stack mitigating factors on top of our secrets to limit their misuse, but ultimately we’re at the mercy of the consuming application to protect the secrets it’s given. Once leaked, lateral movement and data breaches can occur in a matter of minutes. Secretless removes this risk by providing authenticated access to applications without exposing sensitive connection artifacts such as usernames and passwords.

## What’s the problem?
There are many different services and tools available to help facilitate the removal of hard coded secrets from applications and infrastructure. In fact, most organizations will admit to using a number of these tools simultaneously. Secrets can be delivered dynamically via a runtime API call, statically during configuration management or similar, either encrypted or unencrypted. As a developer, this often leads to additional work and dependencies to support each means of delivery.

Once a secret has been consumed by our application, there are a few more considerations to be made. What is the lifecycle of this secret? Does it expire after some time? Is there a chance it will change while it’s still in use? In some cases we can try and catch exceptions thrown as a result of using invalidated secrets, in others we might exit and let the underlying PaaS schedule another instance. Sometimes we endure the very manual process of instituting a change window where it is safe to update a secret and issue a rolling reboot of all our consuming applications. Due to variances in infrastructure, environment and application stack, there is often no standardized way to handle these events across an organization.

These bespoke solutions end up creating another headache down the line for security. How can we guarantee each application is handling its secrets properly? When an external vaulting service passes a secret value to an application’s memory space, that external vaulting service has effectively relinquished full control of that secret to the consuming application. At this point, it becomes much harder to guarantee the secret has not been leaked to logs or exposed through an undiscovered vulnerability.

## Going Secretless
Secretless solves these problems by removing secrets from the application space altogether. Instead, Secretless runs locally as an individual process which automatically authenticates inbound connections to a remote service. Credentials can be consumed seamlessly from a number of providers without any changes to application code. Lifecycle considerations such as expiration and rotation are handled for you behind the scenes. Secretless won’t log your passwords anywhere - and neither will your application (it can’t)!


Sound interesting? Head over to [secretless.io](https://secretless.io) for more information and instructions to get started.


