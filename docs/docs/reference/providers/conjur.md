---
title: Secretless Providers
id: conjur
layout: docs
description: Secretless Documentation
permalink: docs/reference/providers/conjur
---

# CyberArk Conjur
The Conjur provider (`conjur`) populates credentials from an external
[Conjur](https://www.conjur.org) service.

## Examples
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
        provider: conjur
        id: postgres/my-service/address
      - name: username
        provider: conjur
        id: postgres/my-service/username
      - name: password
        provider: conjur
        id: postgres/my-service/password
```
