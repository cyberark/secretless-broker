---
title: Providers
id: vault
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/providers/vault.html
---

## HashiCorp Vault
The Vault provider (`vault`) populates credentials from an external
[Vault](https://www.vaultproject.io/) service.

### Parameters
The value of `id` must be provided in the format `<path>#<key>` _or_ the path
must have a key named `value`.

### Examples
``` yaml
listeners:
  - name: pg_listener
    protocol: pg
    address: 0.0.0.0:5432

handlers:
  - name: pg_handler
    listener: pg_listener
    credentials:
      - name: address
        provider: vault
        id: postgresql/creds#address
      - name: username
        provider: vault
        id: postgresql/creds#username
      - name: password
        provider: vault
        id: postgresql/creds#password
```
