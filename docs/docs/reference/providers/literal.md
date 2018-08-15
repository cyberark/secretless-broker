---
title: Credential Providers
id: Literal
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/providers/literal.html
---

## Literal
The literal provider (`literal`) allows use of hard-coded values as credential
sources.

**Note**: This provider is not intended to be a source of sensitive information.
Secrets stored in this manner are considered insecure and should be provided
through a secure solution such as [Conjur](conjur.html). Take caution and be
mindful of the type of information supplied by this provider.

### Examples
``` yaml
listeners:
  - name: ssh_listener
    protocol: ssh
    address: 0.0.0.0:22

handlers:
  - name: ssh_handler
    listener: ssh_listener
    credentials:
      - name: address
        provider: literal
        id: my-service.myorg.com:29341
      - name: privateKey
        provider: conjur
        id: my-service/ssh-key
```
