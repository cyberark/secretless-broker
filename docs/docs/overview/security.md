---
title: Security of the Secretless Broker
id: docs
layout: docs
description: Secretless Broker Documentation
permalink: docs/overview/security.html
---

Protection of Secretless Broker credential resolution is based on industry-standard practices of keeping the data for the minimal duration needed and hardening of the execution environment. This is ensured by various strategies listed below.

_Keep in mind though that this is an area of functionality that is currently under heavy development and is likely to be updated frequently._

## Hardened container

### Limited User

Our Docker container not only runs within an isolated Docker environment but is also running within a limited-user context. This ensures a layered system of proven SECCOMP kernel security on top of Linux standard user privilege context limitations.

### Minimal Container Runtime

Our container is also based on [Alpine Linux](https://alpinelinux.org/) which has an extremely limited amount of built-in tools, making it much harder for malware to operate in such an environment. Combined with the limited-user runner, the processes within the container are also unable to add additional packages to the container without circumventing Linux ["ring 3"](https://en.wikipedia.org/wiki/Protection_ring) isolation.

## Minimal Credential Retention

One of the biggest security risks for an application is keeping credentials around longer than they are needed. Because Secretless Broker off-loads credential management from the app itself, there will be a window of time where such credentials are retrieved and injected into the backend connection but majority of the time, your credentials would be only stored in your providers that are built for their long-term storage. Since each listener/handler combination is responsible for credential lifecycles these are the lifecycles of credentials for each of the built-in listeners:

- `http`: Listener fetches the credentials on each request and they are stored only for the duration of an individual connection authentication, after which they are zeroized.
- `mysql`: Credentials are loaded for each connection but garbage-collected after connecting to the backend.
- `pg`: Credentials are loaded for each connection but garbage-collected after connecting to the backend.
- `ssh`: Loaded on each new each ssh connection and stored for the duration of the indvidual connection.
- `ssh-agent`: Loaded at listener instantiation time. [Future work](https://github.com/cyberark/secretless-broker/issues/270) will include instantiation of keyring on each connection instead of having a single one that is loaded at start.

[Additional work](https://github.com/cyberark/secretless-broker/issues/271) on zeroization is also pending to ensure that credentials are safe after garbage collection.

## Hardened Networking

You can use both localhost listening address and/or socket files to exchange information between the applications and the Secretless Broker which provide a communication channel that does not leave the host/pod. By having an isolated communication channel between them, you won't have to worry much about external unauthorized access to the Secretless Broker. Additionaly, more security layers can be added to this system (e.g. [encrypted overlay network](https://docs.docker.com/network/overlay/#create-an-overlay-network), Kubernetes pod collocation, etc) for improvement but since they are specific to individual infrastructure deployments and as such are outside of the scope of this document.

_Please note that running the Secretless Broker with a configuration that listens to all interfaces in a non-collocated app deployment allows any other container in the network to connect to the Secretless Broker and authenticate using its configured providers. Use extreme care when your target environment is configured in this way._

## Future Work

Protections beyond the currently-implemented ones for the Secretless Broker are also under consideration (e.g. kernel memory isolation, CPU-specific enclaves, read-only filesystems, etc) and may be added to future releases as the project evolves.

## Additional Notes

While we have taken great care to secure the Secretless Broker and plan to do much more in the future, compared to the apps that would be consumers of the Secretless Broker connectivity, the Secretless Broker itself has a relatively minimal attack surface. Because of this, hardening efforts would in most cases have the best payoff when focused on the application itself and the infrastructure hosting your services.
