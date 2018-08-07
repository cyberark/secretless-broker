---
title: Reference
id: reference
layout: docs
description: Secretless Documentation
permalink: docs/reference
---

## Services
The following target services are currently supported by Secretless:

- MySQL (Socket and TCP)
- PostgreSQL (Socket and TCP)
- SSH
- SSH-Agent
- HTTP (Basic auth, Conjur, and AWS authorization strategies)

With many others in the planning stages!

If there is a specific target service that you would like to be included in this project, please open a [GitHub issue](https://github.com/conjurinc/secretless/issues) with your request.

## Using Secretless

Secretless relies on YAML configuration files to specify which target services it can connect to and how it should retrieve the access credentials to authenticate with those services.

Each Secretless configuration file is composed of two sections:

* `listeners`: A list of protocol Listeners, each one on a Unix socket or TCP port.
* `handlers`: A list of Handlers to process the requests received by each Listener. Handlers implement the protocol for the target services and are configured to obtain the backend connection credentials from one or more Providers.
