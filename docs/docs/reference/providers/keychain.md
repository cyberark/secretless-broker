---
title: Credential Providers
id: keychain
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/providers/keychain.html
---

## Keychain
### Overview
The Keychain provider (`keychain`) allows the use of your OS-level keychain as a
source of credentials.

**Note**: This provider is currently only available on macOS when built from
source. There are plans to integrate all major OS keychains into this provider
in a future release.

## Parameters
### macOS
The value of `id` must be provided in the format `<service>#<account>`.

## Examples
``` yaml
listeners:
  - name: ssh_agent_listener
    protocol: ssh-agent
    socket: /sock/.agent

handlers:
  - name: ssh_agent_handler
    listener: ssh_agent_listener
    credentials:
      - name: rsa
        provider: keychain
        id: identity#rsa-private-key
```
