---
title: Credential Providers
id: conjur
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/providers/conjur.html
---

## CyberArk Conjur
The Conjur provider (`conjur`) populates credentials from an external
[Conjur](https://www.conjur.org) service.

### Secretless Broker Configuration
To use the Conjur provider, Secretless must be configured to authenticate with
Conjur. The Secretless Broker currently supports two methods of authenticating
with Conjur:
- `CONJUR_AUTHN_TOKEN_FILE` environment variable
- `CONJUR_AUTHN_LOGIN` and `CONJUR_AUTHN_API_KEY` environment variables

Both methods also require `CONJUR_APPLIANCE_URL` and `CONJUR_ACCOUNT` to
be set in the environment of the Secretless Broker. You may optionally
also include any other configuration environment variables that are
allowed by the [Conjur Go Client Library](https://github.com/cyberark/conjur-api-go).

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
        provider: conjur
        id: postgres/my-service/address
      - name: username
        provider: conjur
        id: postgres/my-service/username
      - name: password
        provider: conjur
        id: postgres/my-service/password
```
